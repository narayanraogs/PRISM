package tc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateProcedures(t *testing.T) {
	// Create outputs in the current workspace or a folder we can inspect.
	// We will write them directly to /home/narayan/development/PRISM/server/tc or /home/narayan/development/PRISM/
	basePath := "/home/csrspdev/development/PRISM"

	tests := []struct {
		filename string
		commands int
	}{
		{"one.tst", 1},
		{"ten.tst", 10},
		{"long.tst", 300},
	}

	for _, tc := range tests {
		procFunc := CreateProcedure("Rx-Test", "CMD_SET", "CMD_RESET", tc.commands)
		content := procFunc()
		outputPath := filepath.Join(basePath, tc.filename)
		err := os.WriteFile(outputPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write to %s: %v", tc.filename, err)
		}
		t.Logf("Successfully wrote procedure with %d commands to %s", tc.commands, outputPath)
	}
}
