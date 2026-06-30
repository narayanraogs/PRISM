package server

import (
	"sync"
)

var (
	operationMutex     sync.Mutex
	IsOperationRunning bool
)

// TryLockOperation attempts to lock the system for operation execution.
// Returns true if successful, false if the system is already busy.
func TryLockOperation() bool {
	operationMutex.Lock()
	defer operationMutex.Unlock()
	if IsOperationRunning {
		return false
	}
	IsOperationRunning = true
	return true
}

// UnlockOperation unlocks the system.
func UnlockOperation() {
	operationMutex.Lock()
	defer operationMutex.Unlock()
	IsOperationRunning = false
}
