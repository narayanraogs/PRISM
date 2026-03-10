package tc

import (
	"context"
	"prismServer/logger"
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
	manager.mu.Lock()
	_, found := manager.validatedProcedures[name]
	manager.mu.Unlock()

	if !found {
		procedure := procedureProvider()

		if err := manager.client.createProcedure(ctx, name, procedure); err != nil {
			return nil, err
		}

		if err := manager.client.validateProcedure(ctx, name, subsystem); err != nil {
			return nil, err
		}

		manager.mu.Lock()
		manager.validatedProcedures[name] = true
		manager.mu.Unlock()
	}

	return manager.client.executeProcedure(ctx, name)
}

func ExecuteExistingProcedure(ctx context.Context, name string) (<-chan ProcedureResult, error) {
	return manager.client.executeProcedure(ctx, name)
}
