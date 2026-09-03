package safego

import (
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func TestGoRecoversPanic(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	Go("test-go", func() {
		defer wg.Done()
		panic("boom")
	})
	wg.Wait() // The test process must survive the panic.
}

func TestRecoverInWaitGroupGo(t *testing.T) {
	var wg sync.WaitGroup
	wg.Go(func() {
		defer Recover("test-wg")
		panic("boom")
	})
	wg.Wait()
}

func TestRecoverPipeFailsTheReader(t *testing.T) {
	pr, pw := io.Pipe()
	go func() {
		defer RecoverPipe("test-pipe", pw)
		panic("boom")
	}()

	// Without RecoverPipe this read would block forever on a pipe nobody closes.
	_, err := io.ReadAll(pr)
	if err == nil {
		t.Fatal("expected the pipe reader to get an error, got nil")
	}
	if errors.Is(err, io.EOF) {
		t.Fatalf("expected a panic error, got EOF")
	}
}

func TestRecoverToSetsError(t *testing.T) {
	err := func() (err error) {
		defer RecoverTo(&err)
		panic("boom")
	}()

	if err == nil {
		t.Fatal("expected the panic to become an error, got nil")
	}
	if !errors.Is(err, hperrors.ErrPanic) {
		t.Errorf("expected an ErrPanic, got %v", err)
	}
}

func TestRecoverToJoinsExistingError(t *testing.T) {
	sentinel := errors.New("original failure")

	err := func() (err error) {
		defer RecoverTo(&err)
		err = sentinel
		panic("boom")
	}()

	if !errors.Is(err, sentinel) {
		t.Errorf("expected the original error to be kept, got %v", err)
	}
	if !errors.Is(err, hperrors.ErrPanic) {
		t.Errorf("expected the panic to be joined in, got %v", err)
	}
}

func TestRecoverToLeavesErrorAloneWithoutPanic(t *testing.T) {
	sentinel := errors.New("original failure")

	err := func() (err error) {
		defer RecoverTo(&err)
		return sentinel
	}()

	if !errors.Is(err, sentinel) || errors.Is(err, hperrors.ErrPanic) {
		t.Errorf("expected the error to pass through untouched, got %v", err)
	}
}

// RecoverTo(nil) is a misuse the linter rejects, but it must still contain the
// panic and log it rather than crash.
func TestRecoverToNilContainsThePanic(t *testing.T) {
	func() {
		defer RecoverTo(nil)
		panic("boom")
	}()
}
