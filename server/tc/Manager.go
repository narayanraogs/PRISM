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

func getUniqueName(name string, procedure string) string {
	// Compute SHA-256 hash of the procedure content
	hash := sha256.Sum256([]byte(procedure))
	// Convert the first 4 bytes to an 8-character hex string (saves characters to stay within name limits)
	hashStr := hex.EncodeToString(hash[:4])

	// Create a unique name by appending the hash suffix (removing .tst if present)
	baseName := strings.TrimSuffix(name, ".tst")

	// The remote server expects procedure names to be less than 30 characters.
	// "-<hashStr>" takes 9 characters (1 + 8).
	// To ensure uniqueName is strictly less than 30 characters (i.e., <= 29),
	// baseName can be at most 29 - 9 = 20 characters.
	const maxBaseNameLen = 20
	if len(baseName) > maxBaseNameLen {
		baseName = baseName[:maxBaseNameLen]
	}
	return fmt.Sprintf("%s-%s", baseName, hashStr)
}

func RunProcedureFromProvider(ctx context.Context, name string, procedureProvider func() string, subsystem string) (<-chan ProcedureResult, error) {
	procedure := procedureProvider()
	uniqueName := getUniqueName(name, procedure)

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
