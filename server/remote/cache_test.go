package remote

import (
	"reflect"
	"testing"
)

func TestParameterCache(t *testing.T) {
	// Test single parameter
	SetParameter("Power", "-10")
	val, ok := GetParam("Power")
	if !ok || val != "-10" {
		t.Errorf("Expected -10, got %s (ok=%t)", val, ok)
	}

	// Test non-existent single parameter
	_, ok = GetParam("NonExistent")
	if ok {
		t.Errorf("Expected false for non-existent parameter")
	}

	// Test slice parameters
	SetParameters("Profile", []string{"Screenshot", "Magnitude"})
	vals, ok := GetParams("Profile")
	if !ok || !reflect.DeepEqual(vals, []string{"Screenshot", "Magnitude"}) {
		t.Errorf("Expected [Screenshot, Magnitude], got %v (ok=%t)", vals, ok)
	}

	// Test non-existent slice parameter
	_, ok = GetParams("NonExistentSlice")
	if ok {
		t.Errorf("Expected false for non-existent slice parameter")
	}
}
