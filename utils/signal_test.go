package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSignalWait(t *testing.T) {
	signal := NewSignal()

	// Test signaling in a goroutine
	go func() {
		time.Sleep(10 * time.Millisecond) // Small delay to ensure Wait() is called first
		signal.Signal("test response")
	}()

	result := signal.Wait()
	assert.Equal(t, "test response", result)
}

func TestSignalMultipleWaits(t *testing.T) {
	signal := NewSignal()
	signal.Signal("first response")

	// First wait should get the response
	result1 := signal.Wait()
	assert.Equal(t, "first response", result1)

	// Second wait should block since channel is empty
	done := make(chan bool)
	go func() {
		signal.Wait() // This should block
		done <- true
	}()

	select {
	case <-done:
		t.Error("Second Wait() should have blocked")
	case <-time.After(50 * time.Millisecond):
		// Expected: Wait() blocked as expected
	}
}
