// Copyright 2018 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mm

import (
	"fmt"
	"sync/atomic"

	"gvisor.googlesource.com/gvisor/pkg/atomicbitops"
	"gvisor.googlesource.com/gvisor/pkg/sentry/arch"
	"gvisor.googlesource.com/gvisor/pkg/sentry/context"
	"gvisor.googlesource.com/gvisor/pkg/sentry/limits"
	"gvisor.googlesource.com/gvisor/pkg/sentry/memmap"
	"gvisor.googlesource.com/gvisor/pkg/sentry/pgalloc"
	"gvisor.googlesource.com/gvisor/pkg/sentry/platform"
	"gvisor.googlesource.com/gvisor/pkg/sentry/usermem"
)

// NewMemoryManager returns a new MemoryManager with no mappings and 1 user.
func NewMemoryManager(p platform.Platform, mfp pgalloc.MemoryFileProvider) *MemoryManager {
	return &MemoryManager{
		p:           p,
		mfp:         mfp,
		haveASIO:    p.SupportsAddressSpaceIO(),
		privateRefs: &privateRefs{},
		users:       1,
		auxv:        arch.Auxv{},
		aioManager:  aioManager{contexts: make(map[uint64]*AIOContext)},
	}
}

// SetMmapLayout initializes mm's layout from the given arch.Context.
//
// Preconditions: mm contains no mappings and is not used concurrently.
func (mm *MemoryManager) SetMmapLayout(ac arch.Context, r *limits.LimitSet) (arch.MmapLayout, error) {
	layout, err := ac.NewMmapLayout(mm.p.MinUserAddress(), mm.p.MaxUserAddress(), r)
	if err != nil {
		return arch.MmapLayout{}, err
	}
	mm.layout = layout
	return layout, nil
}

// Fork creates a copy of mm with 1 user, as for Linux syscalls fork() or
// clone() (without CLONE_VM).
func (mm *MemoryManager) Fork(ctx context.Context) (*MemoryManager, error) {
	mm.metadataMu.Lock()
	defer mm.metadataMu.Unlock()
	mm.mappingMu.RLock()
	defer mm.mappingMu.RUnlock()
	mm2 := &MemoryManager{
		p:           mm.p,
		mfp:         mm.mfp,
		haveASIO:    mm.haveASIO,
		layout:      mm.layout,
		privateRefs: mm.privateRefs,
		users:       1,
		brk:         mm.brk,
		usageAS:     mm.usageAS,
		dataAS:      mm.dataAS,
		// "The child does not inherit its parent's memory locks (mlock(2),
		// mlockall(2))." - fork(2). So lockedAS is 0 and defMLockMode is
		// MLockNone, both of which are zero values. vma.mlockMode is reset
		// when copied below.
		captureInvalidations: true,
		argv:                 mm.argv,
		envv:                 mm.envv,
		auxv:                 append(arch.Auxv(nil), mm.auxv...),
		// IncRef'd below, once we know that there isn't an error.
		executable: mm.executable,
		aioManager: aioManager{contexts: make(map[uint64]*AIOContext)},
	}

	// Copy vmas.
	dstvgap := mm2.vmas.FirstGap()
	for srcvseg := mm.vmas.FirstSegment(); srcvseg.Ok(); srcvseg = srcvseg.NextSegment() {
		vma := srcvseg.Value() // makes a copy of the vma
		vmaAR := srcvseg.Range()
		// Inform the Mappable, if any, of the new mapping.
		if vma.mappable != nil {
			if err := vma.mappable.AddMapping(ctx, mm2, vmaAR, vma.off, vma.canWriteMappableLocked()); err != nil {
				mm2.removeVMAsLocked(ctx, mm2.applicationAddrRange())
				return nil, err
			}
		}
		if vma.id != nil {
			vma.id.IncRef()
		}
		vma.mlockMode = memmap.MLockNone
		dstvgap = mm2.vmas.Insert(dstvgap, vmaAR, vma).NextGap()
		// We don't need to update mm2.usageAS since we copied it from mm
		// above.
	}

	// Copy pmas. We have to lock mm.activeMu for writing to make existing
	// private pmas copy-on-write. We also have to lock mm2.activeMu since
	// after copying vmas above, memmap.Mappables may call mm2.Invalidate. We
	// only copy private pmas, since in the common case where fork(2) is
	// immediately followed by execve(2), copying non-private pmas that can be
	// regenerated by calling memmap.Mappable.Translate is a waste of time.
	// (Linux does the same; compare kernel/fork.c:dup_mmap() =>
	// mm/memory.c:copy_page_range().)
	mm2.activeMu.Lock()
	defer mm2.activeMu.Unlock()
	mm.activeMu.Lock()
	defer mm.activeMu.Unlock()
	dstpgap := mm2.pmas.FirstGap()
	var unmapAR usermem.AddrRange
	for srcpseg := mm.pmas.FirstSegment(); srcpseg.Ok(); srcpseg = srcpseg.NextSegment() {
		pma := srcpseg.ValuePtr()
		if !pma.private {
			continue
		}
		if !pma.needCOW {
			pma.needCOW = true
			if pma.effectivePerms.Write {
				// We don't want to unmap the whole address space, even though
				// doing so would reduce calls to unmapASLocked(), because mm
				// will most likely continue to be used after the fork, so
				// unmapping pmas unnecessarily will result in extra page
				// faults. But we do want to merge consecutive AddrRanges
				// across pma boundaries.
				if unmapAR.End == srcpseg.Start() {
					unmapAR.End = srcpseg.End()
				} else {
					if unmapAR.Length() != 0 {
						mm.unmapASLocked(unmapAR)
					}
					unmapAR = srcpseg.Range()
				}
				pma.effectivePerms.Write = false
			}
			pma.maxPerms.Write = false
		}
		fr := srcpseg.fileRange()
		mm2.incPrivateRef(fr)
		srcpseg.ValuePtr().file.IncRef(fr)
		addrRange := srcpseg.Range()
		mm2.addRSSLocked(addrRange)
		dstpgap = mm2.pmas.Insert(dstpgap, addrRange, *pma).NextGap()
	}
	if unmapAR.Length() != 0 {
		mm.unmapASLocked(unmapAR)
	}

	// Between when we call memmap.Mappable.AddMapping while copying vmas and
	// when we lock mm2.activeMu to copy pmas, calls to mm2.Invalidate() are
	// ineffective because the pmas they invalidate haven't yet been copied,
	// possibly allowing mm2 to get invalidated translations:
	//
	// Invalidating Mappable            mm.Fork
	// ---------------------            -------
	//
	// mm2.Invalidate()
	//                                  mm.activeMu.Lock()
	// mm.Invalidate() /* blocks */
	//                                  mm2.activeMu.Lock()
	//                                  (mm copies invalidated pma to mm2)
	//
	// This would technically be both safe (since we only copy private pmas,
	// which will still hold a reference on their memory) and consistent with
	// Linux, but we avoid it anyway by setting mm2.captureInvalidations during
	// construction, causing calls to mm2.Invalidate() to be captured in
	// mm2.capturedInvalidations, to be replayed after pmas are copied - i.e.
	// here.
	mm2.captureInvalidations = false
	for _, invArgs := range mm2.capturedInvalidations {
		mm2.invalidateLocked(invArgs.ar, invArgs.opts.InvalidatePrivate, true)
	}
	mm2.capturedInvalidations = nil

	if mm2.executable != nil {
		mm2.executable.IncRef()
	}
	return mm2, nil
}

// IncUsers increments mm's user count and returns true. If the user count is
// already 0, IncUsers does nothing and returns false.
func (mm *MemoryManager) IncUsers() bool {
	return atomicbitops.IncUnlessZeroInt32(&mm.users)
}

// DecUsers decrements mm's user count. If the user count reaches 0, all
// mappings in mm are unmapped.
func (mm *MemoryManager) DecUsers(ctx context.Context) {
	if users := atomic.AddInt32(&mm.users, -1); users > 0 {
		return
	} else if users < 0 {
		panic(fmt.Sprintf("Invalid MemoryManager.users: %d", users))
	}

	mm.aioManager.destroy()

	mm.metadataMu.Lock()
	exe := mm.executable
	mm.executable = nil
	mm.metadataMu.Unlock()
	if exe != nil {
		exe.DecRef()
	}

	mm.activeMu.Lock()
	// Sanity check.
	if atomic.LoadInt32(&mm.active) != 0 {
		panic("active address space lost?")
	}
	// Make sure the AddressSpace is returned.
	if mm.as != nil {
		mm.as.Release()
		mm.as = nil
	}
	mm.activeMu.Unlock()

	mm.mappingMu.Lock()
	defer mm.mappingMu.Unlock()
	// If mm is being dropped before mm.SetMmapLayout was called,
	// mm.applicationAddrRange() will be empty.
	if ar := mm.applicationAddrRange(); ar.Length() != 0 {
		mm.unmapLocked(ctx, ar)
	}
}
