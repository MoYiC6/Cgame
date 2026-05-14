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

func TestAppShutdownAggregatesErrors(t *testing.T) {
	firstErr := stderrors.New("http shutdown failed")
	secondErr := stderrors.New("db shutdown failed")
	first := &shutdownStub{err: firstErr}
	second := &shutdownStub{err: secondErr}
	third := &shutdownStub{}
	app := NewApp(first, nil, second, third)

	err := app.Shutdown(context.Background())
	if err == nil {
		t.Fatal("expected Shutdown to aggregate errors")
	}
	if !stderrors.Is(err, firstErr) {
		t.Fatalf("expected joined error to include first error, got %v", err)
	}
	if !stderrors.Is(err, secondErr) {
		t.Fatalf("expected joined error to include second error, got %v", err)
	}
	if !first.called || !second.called || !third.called {
		t.Fatal("expected all non-nil shutdowners to be called")
	}
}
