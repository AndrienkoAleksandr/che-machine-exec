package registry

import (
	"github.com/eclipse/che-machine-exec/exec/server"
	"sync"
	"sync/atomic"
)

// Exec in memory storage
type ExecRegistry struct {
	mutex   *sync.Mutex
	execMap map[int]*server.ExecSession
	prevExecID uint64
}

func NewExecRegistry() *ExecRegistry {
	return &ExecRegistry{
		prevExecID:0,
		mutex:   &sync.Mutex{},
		execMap: make(map[int]*server.ExecSession),
	}
}

// Add new exec to storage
func (registry *ExecRegistry) Add(exec *server.ExecSession) int  {
	defer registry.mutex.Unlock()

	registry.mutex.Lock()
	exec.ID = int(atomic.AddUint64(&registry.prevExecID, 1))
	registry.execMap[exec.ID] = exec

	return exec.ID
}

// Remove exec from storage
func (registry *ExecRegistry) Remove(id int) {
	defer registry.mutex.Unlock()

	registry.mutex.Lock()
	delete(registry.execMap, id)
}

// Get exec by Id or nil if exec with such id doesn't exists
func (registry *ExecRegistry) GetById(id int) *server.ExecSession {
	defer registry.mutex.Unlock()

	registry.mutex.Lock()
	return registry.execMap[id]
}
