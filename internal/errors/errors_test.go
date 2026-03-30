package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("something broke", FixableByAgent)
	if err.Message != "something broke" {
		t.Errorf("got message %q, want %q", err.Message, "something broke")
	}
	if err.FixableBy != FixableByAgent {
		t.Errorf("got fixableBy %q, want %q", err.FixableBy, FixableByAgent)
	}
	if err.Error() != "something broke" {
		t.Errorf("Error() = %q, want %q", err.Error(), "something broke")
	}
}

func TestNewf(t *testing.T) {
	err := Newf(FixableByHuman, "bad key %q", "xxx")
	if err.Message != `bad key "xxx"` {
		t.Errorf("got message %q", err.Message)
	}
	if err.FixableBy != FixableByHuman {
		t.Errorf("got fixableBy %q", err.FixableBy)
	}
}

func TestWrap(t *testing.T) {
	cause := fmt.Errorf("underlying")
	err := Wrap(cause, FixableByRetry)
	if err.Message != "underlying" {
		t.Errorf("got message %q", err.Message)
	}
	if err.FixableBy != FixableByRetry {
		t.Errorf("got fixableBy %q", err.FixableBy)
	}
	if err.Unwrap() != cause {
		t.Error("Unwrap did not return cause")
	}
}

func TestWrapNil(t *testing.T) {
	if Wrap(nil, FixableByAgent) != nil {
		t.Error("Wrap(nil) should return nil")
	}
}

func TestWithHint(t *testing.T) {
	err := New("bad", FixableByAgent).WithHint("try this")
	if err.Hint != "try this" {
		t.Errorf("got hint %q", err.Hint)
	}
}

func TestWithCause(t *testing.T) {
	cause := fmt.Errorf("root")
	err := New("wrapped", FixableByAgent).WithCause(cause)
	if err.Unwrap() != cause {
		t.Error("WithCause did not set cause")
	}
}

func TestAs(t *testing.T) {
	err := New("test", FixableByAgent)
	var target *APIError
	if !As(err, &target) {
		t.Error("As should match *APIError")
	}
	if target.Message != "test" {
		t.Errorf("got message %q", target.Message)
	}
}

func TestAsWrapped(t *testing.T) {
	inner := New("inner", FixableByHuman)
	outer := fmt.Errorf("outer: %w", inner)
	var target *APIError
	if !errors.As(outer, &target) {
		t.Error("errors.As should unwrap to *APIError")
	}
}

func TestAsNonAPIError(t *testing.T) {
	err := fmt.Errorf("plain error")
	var target *APIError
	if As(err, &target) {
		t.Error("As should not match plain error")
	}
}
