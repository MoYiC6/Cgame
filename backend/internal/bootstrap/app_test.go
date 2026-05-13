package bootstrap

import (
	"context"
	stderrors "errors"
	"testing"
)

type shutdownStub struct {
	err    error
	called bool
}

func (s *shutdownStub) Shutdown(ctx context.Context) error {
	s.called = true
	return s.err
}

func TestAppShutdownCallsAllShutdowners(t *testing.T) {
	first := &shutdownStub{}
	second := &shutdownStub{}
	app := NewApp(first, second)

	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	if !first.called || !second.called {
		t.Fatal("expected all shutdowners to be called")
	}
}

func TestAppShutdownReturnsJoinedError(t *testing.T) {
	app := NewApp(&shutdownStub{err: stderrors.New("db close failed")})

	if err := app.Shutdown(context.Background()); err == nil {
		t.Fatal("expected Shutdown to return error")
	}
}
