package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Bunny3th/easy-workflow/internal/event"
	"github.com/Bunny3th/easy-workflow/internal/model"
)

// testNodeHandler returns nil
func testNodeHandler(ctx context.Context, id int, cur *model.Node, prev model.Node) error {
	return nil
}

// anotherNodeHandler returns nil
func anotherNodeHandler(ctx context.Context, id int, cur *model.Node, prev model.Node) error {
	return nil
}

// testProcHandler returns nil
func testProcHandler(ctx context.Context, instID int, uid string) error {
	return nil
}

// errorNodeHandler returns an error
func errorNodeHandler(ctx context.Context, id int, cur *model.Node, prev model.Node) error {
	return fmt.Errorf("node event error")
}

// errorProcHandler returns an error
func errorProcHandler(ctx context.Context, instID int, uid string) error {
	return fmt.Errorf("proc event error")
}

func TestRegisterNodeEvent(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)

	handler := event.NodeEventHandler(testNodeHandler)
	eng.RegisterNodeEvent("TestNodeEvent", handler)

	eng.eventMu.RLock()
	_, ok := eng.nodeEventPool["TestNodeEvent"]
	eng.eventMu.RUnlock()
	if !ok {
		t.Error("expected TestNodeEvent to be registered in nodeEventPool")
	}

	// verify it does NOT appear in procEventPool
	eng.eventMu.RLock()
	_, ok = eng.procEventPool["TestNodeEvent"]
	eng.eventMu.RUnlock()
	if ok {
		t.Error("expected TestNodeEvent NOT to be registered in procEventPool")
	}
}

func TestRegisterProcEvent(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)

	handler := event.ProcEventHandler(testProcHandler)
	eng.RegisterProcEvent("TestProcEvent", handler)

	eng.eventMu.RLock()
	_, ok := eng.procEventPool["TestProcEvent"]
	eng.eventMu.RUnlock()
	if !ok {
		t.Error("expected TestProcEvent to be registered in procEventPool")
	}

	// verify it does NOT appear in nodeEventPool
	eng.eventMu.RLock()
	_, ok = eng.nodeEventPool["TestProcEvent"]
	eng.eventMu.RUnlock()
	if ok {
		t.Error("expected TestProcEvent NOT to be registered in nodeEventPool")
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
		eng.RegisterNodeEvent("TestNodeEvent", testNodeHandler)
		eng.RegisterNodeEvent("AnotherEvent", anotherNodeHandler)
		err := eng.runNodeEvents(ctx, []string{"TestNodeEvent", "AnotherEvent"}, 1, cur, prev)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	// register errorNodeHandler with ignoreEventErr=true - should not return error
	t.Run("ignore event error", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.ignoreEventErr = true
		eng.RegisterNodeEvent("ErrorEvent", errorNodeHandler)
		err := eng.runNodeEvents(ctx, []string{"ErrorEvent"}, 1, cur, prev)
		if err != nil {
			t.Errorf("expected no error (ignored), got: %v", err)
		}
	})

	// register errorNodeHandler with ignoreEventErr=false - should return error
	t.Run("propagate event error", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.ignoreEventErr = false
		eng.RegisterNodeEvent("ErrorEvent", errorNodeHandler)
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
		eng.RegisterProcEvent("TestProcEvent", testProcHandler)
		err := eng.runProcEvents(ctx, []string{"TestProcEvent"}, 1, "user1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	// ignore event error
	t.Run("ignore event error", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.ignoreEventErr = true
		eng.RegisterProcEvent("ErrorProcEvent", errorProcHandler)
		err := eng.runProcEvents(ctx, []string{"ErrorProcEvent"}, 1, "user1")
		if err != nil {
			t.Errorf("expected no error (ignored), got: %v", err)
		}
	})

	// propagate event error
	t.Run("propagate event error", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.ignoreEventErr = false
		eng.RegisterProcEvent("ErrorProcEvent", errorProcHandler)
		err := eng.runProcEvents(ctx, []string{"ErrorProcEvent"}, 1, "user1")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	// unregistered event
	t.Run("unregistered event", func(t *testing.T) {
		eng := newTestEngine(repo)
		err := eng.runProcEvents(ctx, []string{"UnknownEvent"}, 1, "user1")
		if err == nil {
			t.Error("expected error for unregistered event, got nil")
		}
	})
}

func TestRunProcEvents_GetProcessNameOnce(t *testing.T) {
	// verify GetProcessNameByInstID is called only once, not per event
	callCount := 0
	repo := &mockRepo{
		GetProcessNameByInstIDFunc: func(ctx context.Context, instID int) (string, error) {
			callCount++
			return "test-process", nil
		},
	}
	ctx := context.Background()
	eng := newTestEngine(repo)
	eng.ignoreEventErr = true

	// register multiple proc events
	eng.RegisterProcEvent("Event1", testProcHandler)
	eng.RegisterProcEvent("Event2", testProcHandler)
	eng.RegisterProcEvent("Event3", testProcHandler)

	err := eng.runProcEvents(ctx, []string{"Event1", "Event2", "Event3"}, 1, "user1")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected GetProcessNameByInstID to be called once, got %d calls", callCount)
	}
}

func TestVerifyEvents(t *testing.T) {
	repo := &mockRepo{
		GetProcessResourceFunc: func(ctx context.Context, procID int) (string, error) {
			process := model.Process{
				ProcessName:  "test",
				RevokeEvents: []string{"RegisteredProcEvent"},
			}
			data, _ := json.Marshal(process)
			return string(data), nil
		},
	}
	ctx := context.Background()

	t.Run("all events registered", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.RegisterNodeEvent("RegisteredNodeEvent", testNodeHandler)
		eng.RegisterProcEvent("RegisteredProcEvent", testProcHandler)

		nodes := map[string]model.Node{
			"node1": {NodeStartEvents: []string{"RegisteredNodeEvent"}},
		}
		err := eng.verifyEvents(ctx, 1, nodes)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("unregistered node event", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.RegisterProcEvent("RegisteredProcEvent", testProcHandler)

		nodes := map[string]model.Node{
			"node1": {NodeStartEvents: []string{"MissingNodeEvent"}},
		}
		err := eng.verifyEvents(ctx, 1, nodes)
		if err == nil {
			t.Error("expected error for unregistered node event, got nil")
		}
	})

	t.Run("unregistered proc event", func(t *testing.T) {
		eng := newTestEngine(repo)
		eng.RegisterNodeEvent("RegisteredNodeEvent", testNodeHandler)

		nodes := map[string]model.Node{}
		err := eng.verifyEvents(ctx, 1, nodes)
		if err == nil {
			t.Error("expected error for unregistered proc event, got nil")
		}
	})
}
