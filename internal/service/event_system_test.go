package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Bunny3th/easy-workflow/internal/model"
)

// testNodeEvent has valid node event methods
type testNodeEvent struct{}

func (e *testNodeEvent) ValidNodeEvent(id int, cur *model.Node, prev model.Node) error { return nil }
func (e *testNodeEvent) AnotherEvent(id int, cur *model.Node, prev model.Node) error    { return nil }

// testProcEvent has valid proc event methods
type testProcEvent struct{}

func (e *testProcEvent) ValidRevokeEvent(id int, uid string) error { return nil }

// testEventWithEngine has SetEngine
type testEventWithEngine struct{ eng *Engine }

func (e *testEventWithEngine) SetEngine(eng *Engine) { e.eng = eng }
func (e *testEventWithEngine) SomeNodeEvent(id int, cur *model.Node, prev model.Node) error {
	return nil
}

// badSignatureEvent has wrong method signatures
type badSignatureEvent struct{}

func (e *badSignatureEvent) WrongParams(a, b string) error { return nil }

// errorEvent returns error when called
type errorEvent struct{}

func (e *errorEvent) ErrorEvent(id int, cur *model.Node, prev model.Node) error {
	return fmt.Errorf("event error")
}

// errorProcEvent returns error when called
type errorProcEvent struct{}

func (e *errorProcEvent) ErrorRevokeEvent(id int, uid string) error {
	return fmt.Errorf("proc event error")
}


func TestRegisterEvents(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)

	eng.RegisterEvents(&testNodeEvent{})

	// verify methods are registered
	if _, ok := eng.getEvent("ValidNodeEvent"); !ok {
		t.Error("expected ValidNodeEvent to be registered")
	}
	if _, ok := eng.getEvent("AnotherEvent"); !ok {
		t.Error("expected AnotherEvent to be registered")
	}

	// verify running registered node events works
	ctx := context.Background()
	cur := &model.Node{NodeID: "node1", NodeName: "Node1"}
	prev := model.Node{NodeID: "node0", NodeName: "Node0"}
	err := eng.runNodeEvents(ctx, []string{"ValidNodeEvent"}, 1, cur, prev)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestRegisterEvents_SetEngine(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)

	evt := &testEventWithEngine{}
	eng.RegisterEvents(evt)

	if evt.eng == nil {
		t.Error("expected SetEngine to be called, eng is nil")
	}
}

func TestRunNodeEvents(t *testing.T) {
	repo := &mockRepo{}
	ctx := context.Background()
	cur := &model.Node{NodeID: "node1", NodeName: "Node1"}
	prev := model.Node{NodeID: "node0", NodeName: "Node0"}

	// register events and call runNodeEvents - no error
	t.Run("no error", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.RegisterEvents(&testNodeEvent{})
		err := eng.runNodeEvents(ctx, []string{"ValidNodeEvent", "AnotherEvent"}, 1, cur, prev)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	// register errorEvent with ignoreEventErr=true - should not return error
	t.Run("ignore event error", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.ignoreEventErr = true
		eng.RegisterEvents(&errorEvent{})
		err := eng.runNodeEvents(ctx, []string{"ErrorEvent"}, 1, cur, prev)
		if err != nil {
			t.Errorf("expected no error (ignored), got: %v", err)
		}
	})

	// register errorEvent with ignoreEventErr=false - should return error
	t.Run("propagate event error", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.ignoreEventErr = false
		eng.RegisterEvents(&errorEvent{})
		err := eng.runNodeEvents(ctx, []string{"ErrorEvent"}, 1, cur, prev)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	// unregistered event
	t.Run("unregistered event", func(t *testing.T) {
		eng := newTestEngine(repo)
		err := eng.runNodeEvents(ctx, []string{"UnknownEvent"}, 1, cur, prev)
		if err == nil {
			t.Error("expected error for unregistered event, got nil")
		}
	})
}

func TestRunProcEvents(t *testing.T) {
	repo := &mockRepo{
		GetProcessNameByInstIDFunc: func(ctx context.Context, instID int) (string, error) {
			return "test-process", nil
		},
	}
	ctx := context.Background()

	// no error
	t.Run("no error", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.RegisterEvents(&testProcEvent{})
		err := eng.runProcEvents(ctx, []string{"ValidRevokeEvent"}, 1, "user1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	// ignore event error
	t.Run("ignore event error", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.ignoreEventErr = true
		eng.RegisterEvents(&errorProcEvent{})
		err := eng.runProcEvents(ctx, []string{"ErrorRevokeEvent"}, 1, "user1")
		if err != nil {
			t.Errorf("expected no error (ignored), got: %v", err)
		}
	})

	// propagate event error
	t.Run("propagate event error", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.ignoreEventErr = false
		eng.RegisterEvents(&errorProcEvent{})
		err := eng.runProcEvents(ctx, []string{"ErrorRevokeEvent"}, 1, "user1")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestVerifyEvents(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)

	// register event with wrong signature
	eng.RegisterEvents(&badSignatureEvent{})

	// verifyNodeEventParams should detect wrong signatures
	em, ok := eng.getEvent("WrongParams")
	if !ok {
		t.Fatal("expected WrongParams to be registered")
	}
	err := eng.verifyNodeEventParams(em)
	if err == nil {
		t.Error("expected error for wrong method signature, got nil")
	}

	// verifyProcEventParams should also detect wrong signatures
	err = eng.verifyProcEventParams(em)
	if err == nil {
		t.Error("expected error for wrong proc event signature, got nil")
	}

	// verify valid node event
	eng.RegisterEvents(&testNodeEvent{})
	em, ok = eng.getEvent("ValidNodeEvent")
	if !ok {
		t.Fatal("expected ValidNodeEvent to be registered")
	}
	err = eng.verifyNodeEventParams(em)
	if err != nil {
		t.Errorf("expected no error for valid node event, got: %v", err)
	}

	// verify valid proc event
	eng.RegisterEvents(&testProcEvent{})
	em, ok = eng.getEvent("ValidRevokeEvent")
	if !ok {
		t.Fatal("expected ValidRevokeEvent to be registered")
	}
	err = eng.verifyProcEventParams(em)
	if err != nil {
		t.Errorf("expected no error for valid proc event, got: %v", err)
	}
}
