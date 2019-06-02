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

package fsutil

import (
	"math"

	"gvisor.googlesource.com/gvisor/pkg/sentry/context"
	"gvisor.googlesource.com/gvisor/pkg/sentry/memmap"
	"gvisor.googlesource.com/gvisor/pkg/sentry/platform"
	"gvisor.googlesource.com/gvisor/pkg/sentry/safemem"
	"gvisor.googlesource.com/gvisor/pkg/sentry/usermem"
)

// DirtySet maps offsets into a memmap.Mappable to DirtyInfo. It is used to
// implement Mappables that cache data from another source.
//
// type DirtySet <generated by go_generics>

// DirtyInfo is the value type of DirtySet, and represents information about a
// Mappable offset that is dirty (the cached data for that offset is newer than
// its source).
//
// +stateify savable
type DirtyInfo struct {
	// Keep is true if the represented offset is concurrently writable, such
	// that writing the data for that offset back to the source does not
	// guarantee that the offset is clean (since it may be concurrently
	// rewritten after the writeback).
	Keep bool
}

// dirtySetFunctions implements segment.Functions for DirtySet.
type dirtySetFunctions struct{}

// MinKey implements segment.Functions.MinKey.
func (dirtySetFunctions) MinKey() uint64 {
	return 0
}

// MaxKey implements segment.Functions.MaxKey.
func (dirtySetFunctions) MaxKey() uint64 {
	return math.MaxUint64
}

// ClearValue implements segment.Functions.ClearValue.
func (dirtySetFunctions) ClearValue(val *DirtyInfo) {
}

// Merge implements segment.Functions.Merge.
func (dirtySetFunctions) Merge(_ memmap.MappableRange, val1 DirtyInfo, _ memmap.MappableRange, val2 DirtyInfo) (DirtyInfo, bool) {
	if val1 != val2 {
		return DirtyInfo{}, false
	}
	return val1, true
}

// Split implements segment.Functions.Split.
func (dirtySetFunctions) Split(_ memmap.MappableRange, val DirtyInfo, _ uint64) (DirtyInfo, DirtyInfo) {
	return val, val
}

// MarkClean marks all offsets in mr as not dirty, except for those to which
// KeepDirty has been applied.
func (ds *DirtySet) MarkClean(mr memmap.MappableRange) {
	seg := ds.LowerBoundSegment(mr.Start)
	for seg.Ok() && seg.Start() < mr.End {
		if seg.Value().Keep {
			seg = seg.NextSegment()
			continue
		}
		seg = ds.Isolate(seg, mr)
		seg = ds.Remove(seg).NextSegment()
	}
}

// KeepClean marks all offsets in mr as not dirty, even those that were
// previously kept dirty by KeepDirty.
func (ds *DirtySet) KeepClean(mr memmap.MappableRange) {
	ds.RemoveRange(mr)
}

// MarkDirty marks all offsets in mr as dirty.
func (ds *DirtySet) MarkDirty(mr memmap.MappableRange) {
	ds.setDirty(mr, false)
}

// KeepDirty marks all offsets in mr as dirty and prevents them from being
// marked as clean by MarkClean.
func (ds *DirtySet) KeepDirty(mr memmap.MappableRange) {
	ds.setDirty(mr, true)
}

func (ds *DirtySet) setDirty(mr memmap.MappableRange, keep bool) {
	var changedAny bool
	defer func() {
		if changedAny {
			// Merge segments split by Isolate to reduce cost of iteration.
			ds.MergeRange(mr)
		}
	}()
	seg, gap := ds.Find(mr.Start)
	for {
		switch {
		case seg.Ok() && seg.Start() < mr.End:
			if keep && !seg.Value().Keep {
				changedAny = true
				seg = ds.Isolate(seg, mr)
				seg.ValuePtr().Keep = true
			}
			seg, gap = seg.NextNonEmpty()

		case gap.Ok() && gap.Start() < mr.End:
			changedAny = true
			seg = ds.Insert(gap, gap.Range().Intersect(mr), DirtyInfo{keep})
			seg, gap = seg.NextNonEmpty()

		default:
			return
		}
	}
}

// AllowClean allows MarkClean to mark offsets in mr as not dirty, ending the
// effect of a previous call to KeepDirty. (It does not itself mark those
// offsets as not dirty.)
func (ds *DirtySet) AllowClean(mr memmap.MappableRange) {
	var changedAny bool
	defer func() {
		if changedAny {
			// Merge segments split by Isolate to reduce cost of iteration.
			ds.MergeRange(mr)
		}
	}()
	for seg := ds.LowerBoundSegment(mr.Start); seg.Ok() && seg.Start() < mr.End; seg = seg.NextSegment() {
		if seg.Value().Keep {
			changedAny = true
			seg = ds.Isolate(seg, mr)
			seg.ValuePtr().Keep = false
		}
	}
}

// SyncDirty passes pages in the range mr that are stored in cache and
// identified as dirty to writeAt, updating dirty to reflect successful writes.
// If writeAt returns a successful partial write, SyncDirty will call it
// repeatedly until all bytes have been written. max is the true size of the
// cached object; offsets beyond max will not be passed to writeAt, even if
// they are marked dirty.
func SyncDirty(ctx context.Context, mr memmap.MappableRange, cache *FileRangeSet, dirty *DirtySet, max uint64, mem platform.File, writeAt func(ctx context.Context, srcs safemem.BlockSeq, offset uint64) (uint64, error)) error {
	var changedDirty bool
	defer func() {
		if changedDirty {
			// Merge segments split by Isolate to reduce cost of iteration.
			dirty.MergeRange(mr)
		}
	}()
	dseg := dirty.LowerBoundSegment(mr.Start)
	for dseg.Ok() && dseg.Start() < mr.End {
		var dr memmap.MappableRange
		if dseg.Value().Keep {
			dr = dseg.Range().Intersect(mr)
		} else {
			changedDirty = true
			dseg = dirty.Isolate(dseg, mr)
			dr = dseg.Range()
		}
		if err := syncDirtyRange(ctx, dr, cache, max, mem, writeAt); err != nil {
			return err
		}
		if dseg.Value().Keep {
			dseg = dseg.NextSegment()
		} else {
			dseg = dirty.Remove(dseg).NextSegment()
		}
	}
	return nil
}

// SyncDirtyAll passes all pages stored in cache identified as dirty to
// writeAt, updating dirty to reflect successful writes. If writeAt returns a
// successful partial write, SyncDirtyAll will call it repeatedly until all
// bytes have been written. max is the true size of the cached object; offsets
// beyond max will not be passed to writeAt, even if they are marked dirty.
func SyncDirtyAll(ctx context.Context, cache *FileRangeSet, dirty *DirtySet, max uint64, mem platform.File, writeAt func(ctx context.Context, srcs safemem.BlockSeq, offset uint64) (uint64, error)) error {
	dseg := dirty.FirstSegment()
	for dseg.Ok() {
		if err := syncDirtyRange(ctx, dseg.Range(), cache, max, mem, writeAt); err != nil {
			return err
		}
		if dseg.Value().Keep {
			dseg = dseg.NextSegment()
		} else {
			dseg = dirty.Remove(dseg).NextSegment()
		}
	}
	return nil
}

// Preconditions: mr must be page-aligned.
func syncDirtyRange(ctx context.Context, mr memmap.MappableRange, cache *FileRangeSet, max uint64, mem platform.File, writeAt func(ctx context.Context, srcs safemem.BlockSeq, offset uint64) (uint64, error)) error {
	for cseg := cache.LowerBoundSegment(mr.Start); cseg.Ok() && cseg.Start() < mr.End; cseg = cseg.NextSegment() {
		wbr := cseg.Range().Intersect(mr)
		if max < wbr.Start {
			break
		}
		ims, err := mem.MapInternal(cseg.FileRangeOf(wbr), usermem.Read)
		if err != nil {
			return err
		}
		if max < wbr.End {
			ims = ims.TakeFirst64(max - wbr.Start)
		}
		offset := wbr.Start
		for !ims.IsEmpty() {
			n, err := writeAt(ctx, ims, offset)
			if err != nil {
				return err
			}
			offset += n
			ims = ims.DropFirst64(n)
		}
	}
	return nil
}
