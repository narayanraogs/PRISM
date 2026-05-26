package tc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"prismServer/logger"
	"strings"
	"sync"
)

var manager ProcedureManager

func Init() {
	client, err := NewClient()
	if err != nil {
		logger.Log.Error("Cannot connect to Procedure Server", "error", err.Error())
	}
	manager = ProcedureManager{
		client:              client,
		validatedProcedures: make(map[string]bool),
	}
}

type ProcedureManager struct {
	client              *Client
	validatedProcedures map[string]bool
	mu                  sync.Mutex
}

func RunProcedure(ctx context.Context, name, procedure, subsystem string) (<-chan ProcedureResult, error) {
	return RunProcedureFromProvider(ctx, name, func() string { return procedure }, subsystem)
}

func RunProcedureFromProvider(ctx context.Context, name string, procedureProvider func() string, subsystem string) (<-chan ProcedureResult, error) {
	procedure := procedureProvider()

	// Compute SHA-256 hash of the procedure content
	hash := sha256.Sum256([]byte(procedure))
	// Convert the first 8 bytes to a 16-character hex string
	hashStr := hex.EncodeToString(hash[:8])

	// Create a unique name by appending the hash suffix (e.g., test-a1b2c3d4e5f6g7h8.tst)
	baseName := strings.TrimSuffix(name, ".tst")
	uniqueName := fmt.Sprintf("%s-%s.tst", baseName, hashStr)

	manager.mu.Lock()
	_, found := manager.validatedProcedures[uniqueName]
	manager.mu.Unlock()

	if !found {
		if err := manager.client.createProcedure(ctx, uniqueName, procedure); err != nil {
			return nil, err
		}

		if err := manager.client.validateProcedure(ctx, uniqueName, subsystem); err != nil {
			return nil, err
		}

		manager.mu.Lock()
		manager.validatedProcedures[uniqueName] = true
		manager.mu.Unlock()
	}

	return manager.client.executeProcedure(ctx, uniqueName)
}

func ExecuteExistingProcedure(ctx context.Context, name string) (<-chan ProcedureResult, error) {
	return manager.client.executeProcedure(ctx, name)
}
