package local

import "testing"

func TestRun(t *testing.T) {
	if err := Run("echo", "hello"); err != nil {
		t.Error(err)
	}
	if err := Run("lalala", "hello"); err != nil {
		t.Error(err)
	}
}
