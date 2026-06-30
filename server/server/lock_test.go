package server

import (
	"testing"
)

func TestGlobalLock(t *testing.T) {
	// Ensure lock starts unlocked
	UnlockOperation()
	if IsOperationRunning {
		t.Errorf("Expected IsOperationRunning to be false initially")
	}

	// Try acquiring lock
	ok := TryLockOperation()
	if !ok {
		t.Errorf("Expected to successfully acquire lock")
	}
	if !IsOperationRunning {
		t.Errorf("Expected IsOperationRunning to be true after locking")
	}

	// Try acquiring again (should fail)
	ok = TryLockOperation()
	if ok {
		t.Errorf("Expected locking to fail when already locked")
	}

	// Unlock
	UnlockOperation()
	if IsOperationRunning {
		t.Errorf("Expected IsOperationRunning to be false after unlocking")
	}

	// Lock again (should succeed)
	ok = TryLockOperation()
	if !ok {
		t.Errorf("Expected to lock successfully after unlock")
	}
	UnlockOperation() // Cleanup
}
