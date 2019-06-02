// automatically generated by stateify.

package semaphore

import (
	"gvisor.googlesource.com/gvisor/pkg/state"
)

func (x *Registry) beforeSave() {}
func (x *Registry) save(m state.Map) {
	x.beforeSave()
	m.Save("userNS", &x.userNS)
	m.Save("semaphores", &x.semaphores)
	m.Save("lastIDUsed", &x.lastIDUsed)
}

func (x *Registry) afterLoad() {}
func (x *Registry) load(m state.Map) {
	m.Load("userNS", &x.userNS)
	m.Load("semaphores", &x.semaphores)
	m.Load("lastIDUsed", &x.lastIDUsed)
}

func (x *Set) beforeSave() {}
func (x *Set) save(m state.Map) {
	x.beforeSave()
	m.Save("registry", &x.registry)
	m.Save("ID", &x.ID)
	m.Save("key", &x.key)
	m.Save("creator", &x.creator)
	m.Save("owner", &x.owner)
	m.Save("perms", &x.perms)
	m.Save("opTime", &x.opTime)
	m.Save("changeTime", &x.changeTime)
	m.Save("sems", &x.sems)
	m.Save("dead", &x.dead)
}

func (x *Set) afterLoad() {}
func (x *Set) load(m state.Map) {
	m.Load("registry", &x.registry)
	m.Load("ID", &x.ID)
	m.Load("key", &x.key)
	m.Load("creator", &x.creator)
	m.Load("owner", &x.owner)
	m.Load("perms", &x.perms)
	m.Load("opTime", &x.opTime)
	m.Load("changeTime", &x.changeTime)
	m.Load("sems", &x.sems)
	m.Load("dead", &x.dead)
}

func (x *sem) beforeSave() {}
func (x *sem) save(m state.Map) {
	x.beforeSave()
	if !state.IsZeroValue(x.waiters) { m.Failf("waiters is %v, expected zero", x.waiters) }
	m.Save("value", &x.value)
	m.Save("pid", &x.pid)
}

func (x *sem) afterLoad() {}
func (x *sem) load(m state.Map) {
	m.Load("value", &x.value)
	m.Load("pid", &x.pid)
}

func (x *waiter) beforeSave() {}
func (x *waiter) save(m state.Map) {
	x.beforeSave()
	m.Save("waiterEntry", &x.waiterEntry)
	m.Save("value", &x.value)
	m.Save("ch", &x.ch)
}

func (x *waiter) afterLoad() {}
func (x *waiter) load(m state.Map) {
	m.Load("waiterEntry", &x.waiterEntry)
	m.Load("value", &x.value)
	m.Load("ch", &x.ch)
}

func (x *waiterList) beforeSave() {}
func (x *waiterList) save(m state.Map) {
	x.beforeSave()
	m.Save("head", &x.head)
	m.Save("tail", &x.tail)
}

func (x *waiterList) afterLoad() {}
func (x *waiterList) load(m state.Map) {
	m.Load("head", &x.head)
	m.Load("tail", &x.tail)
}

func (x *waiterEntry) beforeSave() {}
func (x *waiterEntry) save(m state.Map) {
	x.beforeSave()
	m.Save("next", &x.next)
	m.Save("prev", &x.prev)
}

func (x *waiterEntry) afterLoad() {}
func (x *waiterEntry) load(m state.Map) {
	m.Load("next", &x.next)
	m.Load("prev", &x.prev)
}

func init() {
	state.Register("semaphore.Registry", (*Registry)(nil), state.Fns{Save: (*Registry).save, Load: (*Registry).load})
	state.Register("semaphore.Set", (*Set)(nil), state.Fns{Save: (*Set).save, Load: (*Set).load})
	state.Register("semaphore.sem", (*sem)(nil), state.Fns{Save: (*sem).save, Load: (*sem).load})
	state.Register("semaphore.waiter", (*waiter)(nil), state.Fns{Save: (*waiter).save, Load: (*waiter).load})
	state.Register("semaphore.waiterList", (*waiterList)(nil), state.Fns{Save: (*waiterList).save, Load: (*waiterList).load})
	state.Register("semaphore.waiterEntry", (*waiterEntry)(nil), state.Fns{Save: (*waiterEntry).save, Load: (*waiterEntry).load})
}
