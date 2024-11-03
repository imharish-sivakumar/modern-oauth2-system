package constants

import (
	"testing"
)

func TestIsLocal(t *testing.T) {
	if !Environment("LOCAL").IsLocal() {
		t.Error("expected true but got false")
	}
}
