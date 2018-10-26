package registry

import (
	"github.com/eclipse/che-machine-exec/exec/server"
	"sync"
)

// Exec in memory storage
type ExecRegistry struct {
	mutex   *sync.Mutex
	execMap map[int]*server.ServerExec
}

// Add new exec to storage
func (registry *ExecRegistry) Add(exec server.ServerExec)  {
	defer registry.mutex.Unlock()

	registry.execMap[exec.ID] = &exec
}

// Remove exec from storage
func (registry *ExecRegistry) Remove(id int)  {
	defer registry.mutex.Unlock()

	delete(registry.execMap, id)
}

// Get exec by Id or nil if exec with such id doesn't exists
func (registry *ExecRegistry) GetById(id int) *server.ServerExec {
	defer registry.mutex.Unlock()

	registry.mutex.Lock()
	return registry.execMap[id]
}
