package ui

import (
	"testing"
)

func TestInitialModel(t *testing.T) {
	model := initialModel()
	if model.isLoading != true {
		t.Error("Expected initial model to be in loading state")
	}
	if model.status != "Checking prerequisites..." {
		t.Errorf("Expected initial status to be 'Checking prerequisites...', got '%s'", model.status)
	}
}
