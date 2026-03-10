package tc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"prismServer/utils"
	"strconv"
	"strings"
	"time"
)

func NewClient() (*Client, error) {
	serverIP := utils.Config.ProcedureServer.IP
	port := utils.Config.ProcedureServer.PortNo
	path := utils.Config.ProcedureServer.Path
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	baseURL := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(serverIP, strconv.Itoa(port)),
		Path:   path,
	}
	return &Client{
		BaseURL:      baseURL.String(),
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
		PollInterval: 10 * time.Second,
	}, nil
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, body interface{}) (*response, error) {
	var buf io.ReadWriter
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		buf = bytes.NewBuffer(b)
	}
	fmt.Println(c.BaseURL+endpoint)

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+endpoint, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request failed with status %s: %s", resp.Status, string(respBody))
	}

	var ack response
	if err := json.Unmarshal(respBody, &ack); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &ack, nil
}

func (c *Client) createProcedure(ctx context.Context, name string, procedure string) error {
	reqBody := createRequest{
		ProcName:  name,
		Procedure: procedure,
	}
	ack, err := c.newRequest(ctx, http.MethodPost, "/createProcedure", reqBody)
	if err != nil {
		return err
	}

	if !ack.Ack {
		if strings.EqualFold(ack.ErrorMsg, "Procedure Already exists") {
			return nil
		}
		return fmt.Errorf("procedure creation not acknowledged: %s", ack.ErrorMsg)
	}
	return nil
}

func (c *Client) validateProcedure(ctx context.Context, name string, subsystem string) error {
	reqBody := validateRequest{
		ProcName:  name,
		ProcSrc:   "PRISM",
		Subsystem: subsystem,
	}
	ack, err := c.newRequest(ctx, http.MethodPost, "/validateProcedure", reqBody)
	if err != nil {
		return err
	}

	if !ack.Ack {
		return fmt.Errorf("procedure validation not acknowledged: %s", ack.ErrorMsg)
	}
	return nil
}

func (c *Client) executeProcedure(ctx context.Context, name string) (<-chan ProcedureResult, error) {
	reqBody := loadRequest{
		Action:       "execute",
		ProcName:     "rate-sm-1.tst",
		ProcSrc:      "PRISM",
		ProcMode:     "auto",
		ProcPriority: "high",
	}
	ack, err := c.newRequest(ctx, http.MethodPost, "/loadProcedure", reqBody)
	if err != nil {
		return nil, err
	}

	if !ack.Ack {
		return nil, fmt.Errorf("procedure execution not acknowledged: %s", ack.ErrorMsg)
	}

	resultChan := make(chan ProcedureResult, 1)
	go c.pollForStatus(ctx, name, resultChan)

	return resultChan, nil
}

func (c *Client) pollForStatus(ctx context.Context, name string, resultChan chan<- ProcedureResult) {
	defer close(resultChan)

	ticker := time.NewTicker(c.PollInterval)
	defer ticker.Stop()

	reqBody := loadRequest{
		Action:       "exestatus",
		ProcName:     "rate-sm-1.tst",
		ProcSrc:      "PRISM",
		ProcMode:     "auto",
		ProcPriority: "high",
	}

	for {
		select {
		case <-ctx.Done():
			resultChan <- ProcedureResult{Success: false, Status: StatusAborted, Err: ctx.Err()}
			return
		case <-ticker.C:
			ack, err := c.newRequest(ctx, http.MethodPost, "/getExeStatus", reqBody)
			if err != nil {
				resultChan <- ProcedureResult{Success: false, Err: fmt.Errorf("polling failed: %w", err)}
				return
			}

			if !ack.Ack {
				resultChan <- ProcedureResult{Success: false, Status: Status(ack.ExeStatus), Err: fmt.Errorf(ack.ErrorMsg)}
				return
			}

			status := Status(strings.ToLower(ack.ExeStatus))
			switch status {
			case StatusSuccess:
				resultChan <- ProcedureResult{Success: true, Status: status}
				return
			case StatusFailure, StatusAborted, StatusSuspended, StatusNotAvailable:
				resultChan <- ProcedureResult{Success: false, Status: status, Err: fmt.Errorf("procedure terminated with status: %s", status)}
				return
			case StatusInProgress, StatusQueued:
				// Continue polling
			default:
				resultChan <- ProcedureResult{Success: false, Status: StatusUnknown, Err: fmt.Errorf("received unknown procedure status: %s", status)}
				return
			}
		}
	}
}
