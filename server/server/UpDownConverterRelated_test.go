package server

import (
	"encoding/json"
	"testing"
)

func TestUCDCRequestCableLossesSerialization(t *testing.T) {
	reqJSON := `{
		"ConverterName": "UC-17-03-00",
		"DeviceProfile": "Standard",
		"ExternalSGName": "SG-01",
		"InputCableLoss": 0.5,
		"InputPower": -30.0,
		"LOCableLoss": 0.3,
		"OutputCableLoss": [0.1, 0.2, 0.3],
		"CableLosses": [
			{
				"InputCableLoss": 1.2,
				"LOCableLoss": 0.8,
				"OutputCableLoss": [0.4, 0.5, 0.6]
			},
			{
				"InputCableLoss": 2.1,
				"LOCableLoss": 1.5,
				"OutputCableLoss": [0.7, 0.8, 0.9]
			}
		],
		"StepSize": 5.0,
		"TestsSelected": ["UCDC_GAIN_INT_CABLE"]
	}`

	var req UCDCRequest
	err := json.Unmarshal([]byte(reqJSON), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal UCDCRequest: %v", err)
	}

	if len(req.CableLosses) != 2 {
		t.Fatalf("Expected 2 CableLosses sets, got %d", len(req.CableLosses))
	}

	if req.CableLosses[0].InputCableLoss != 1.2 {
		t.Errorf("Expected CableLosses[0].InputCableLoss=1.2, got %f", req.CableLosses[0].InputCableLoss)
	}
	if req.CableLosses[0].LOCableLoss != 0.8 {
		t.Errorf("Expected CableLosses[0].LOCableLoss=0.8, got %f", req.CableLosses[0].LOCableLoss)
	}
	if len(req.CableLosses[0].OutputCableLoss) != 3 || req.CableLosses[0].OutputCableLoss[1] != 0.5 {
		t.Errorf("Unexpected CableLosses[0].OutputCableLoss: %v", req.CableLosses[0].OutputCableLoss)
	}

	if req.CableLosses[1].InputCableLoss != 2.1 {
		t.Errorf("Expected CableLosses[1].InputCableLoss=2.1, got %f", req.CableLosses[1].InputCableLoss)
	}
	if req.CableLosses[1].LOCableLoss != 1.5 {
		t.Errorf("Expected CableLosses[1].LOCableLoss=1.5, got %f", req.CableLosses[1].LOCableLoss)
	}
	if len(req.CableLosses[1].OutputCableLoss) != 3 || req.CableLosses[1].OutputCableLoss[2] != 0.9 {
		t.Errorf("Unexpected CableLosses[1].OutputCableLoss: %v", req.CableLosses[1].OutputCableLoss)
	}
}
