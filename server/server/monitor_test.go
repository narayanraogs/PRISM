package server

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"prismServer/executeTest"
	"prismServer/logger"
	"prismServer/reports"
	"prismServer/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type mockTester struct{}

func (m *mockTester) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	ctx.UpdateChannel <- executeTest.SingleTestProgress{CurrentStep: "MockTestStep"}
}
func (m *mockTester) Rollback() error { return nil }
func (m *mockTester) DBValidate() error { return nil }
func (m *mockTester) GetRollbackDetails() map[string]utils.CommandResponse { return nil }
func (m *mockTester) SetRollbackMap(r map[string]utils.CommandResponse) {}
func (m *mockTester) Measure(ctx context.Context) error {
	time.Sleep(500 * time.Millisecond)
	return nil
}
func (m *mockTester) GenerateReport() (map[string]reports.Result, error) { return nil, nil }
func (m *mockTester) GenerateFailureReport(err error) (map[string]reports.Result, error) { return nil, nil }
func (m *mockTester) SetParameters(map[string]interface{}) error { return nil }

func TestMonitorMultiplexing(t *testing.T) {
	// Initialize mock logger to prevent nil panic
	logger.Log = slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Register dummy test procedure to bypass DB validations
	executeTest.Register("DummyTest", "DummyCat", func() executeTest.Tester {
		return &mockTester{}
	})

	// Set up gin router
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/testProgress", testProgress)

	// Create test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Convert http URL to ws URL
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/testProgress"

	// Ensure lock starts unlocked
	UnlockOperation()

	// Connect first client (the executor)
	dialer := websocket.Dialer{}
	conn1, _, err := dialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to dial client 1: %v", err)
	}
	defer conn1.Close()

	// Send initial test registration to client 1 (executor)
	var req StartTestsRequest
	req.Tests = []TestDescription{
		{TestName: "DummyTest", TestCategory: "DummyCat", Configuration: "DummyCfg"},
	}
	err = conn1.WriteJSON(req)
	if err != nil {
		t.Fatalf("Client 1 failed to send registration: %v", err)
	}

	// Give the server a small moment to acquire the lock and register
	time.Sleep(100 * time.Millisecond)

	if !IsOperationRunning {
		t.Errorf("Expected IsOperationRunning to be true after client 1 registers tests")
	}

	// Connect second client (the monitor)
	conn2, _, err := dialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to dial monitor client 2: %v", err)
	}
	defer conn2.Close()

	// Send initial monitor registration JSON
	err = conn2.WriteJSON(StartTestsRequest{})
	if err != nil {
		t.Fatalf("Monitor client 2 failed to send registration: %v", err)
	}

	// Give server time to parse registration and register monitor connection
	time.Sleep(150 * time.Millisecond)

	// Check if monitor is registered in the list
	monitorMutex.Lock()
	monitorsCount := len(monitors)
	monitorMutex.Unlock()
	if monitorsCount != 1 {
		t.Errorf("Expected 1 monitor registered, got %d", monitorsCount)
	}

	// Broadcast dummy progress update
	var update executeTest.TestProgressResponse
	update.OK = true
	update.Message = "Mock progress status update"
	broadcastToMonitors(update)

	// Check if monitor receives the update
	var rcvUpdate executeTest.TestProgressResponse
	err = conn2.SetReadDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		t.Fatalf("Failed to set read deadline: %v", err)
	}
	err = conn2.ReadJSON(&rcvUpdate)
	if err != nil {
		t.Fatalf("Monitor failed to read broadcasted update: %v", err)
	}

	if rcvUpdate.Message != "Mock progress status update" {
		t.Errorf("Expected 'Mock progress status update', got: %s", rcvUpdate.Message)
	}

	// Disconnect monitor and check clean up
	conn2.Close()
	time.Sleep(100 * time.Millisecond)

	monitorMutex.Lock()
	monitorsCount = len(monitors)
	monitorMutex.Unlock()
	if monitorsCount != 0 {
		t.Errorf("Expected 0 monitors registered after disconnect, got %d", monitorsCount)
	}

	// Clean up executor
	conn1.Close()
	time.Sleep(100 * time.Millisecond)
	UnlockOperation()
}
