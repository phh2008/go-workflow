package service

import (
	"testing"
)

func TestEval(t *testing.T) {
	ev := NewExpressionEvaluator()

	// simple comparison: $days >= 3 with days = 5 => true
	result, err := ev.Eval("$days>=3", map[string]any{"days": 5})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !result {
		t.Errorf("expected true, got false")
	}

	// complex expression: $days>=3 and $level > 1
	result, err = ev.Eval("$days>=3 and $level > 1", map[string]any{"days": 5, "level": 2})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !result {
		t.Errorf("expected true, got false")
	}

	// non-existent variable should return error
	_, err = ev.Eval("$unknown > 0", map[string]any{"days": 5})
	if err == nil {
		t.Fatal("expected error for non-existent variable, got nil")
	}

	// invalid expression syntax
	_, err = ev.Eval("$days >=", map[string]any{"days": 5})
	if err == nil {
		t.Fatal("expected error for invalid syntax, got nil")
	}
}

func TestEvalWithRawEnv(t *testing.T) {
	ev := NewExpressionEvaluator()

	// number comparison with raw env (key has $ prefix, value is string)
	result, err := ev.EvalWithRawEnv("$days>=3", map[string]string{"$days": "5"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !result {
		t.Errorf("expected true, got false")
	}

	// string comparison: $name == "alice"
	result, err = ev.EvalWithRawEnv("$name == \"alice\"", map[string]string{"$name": "alice"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !result {
		t.Errorf("expected true, got false")
	}

	// bool variable
	result, err = ev.EvalWithRawEnv("$flag", map[string]string{"$flag": "true"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !result {
		t.Errorf("expected true, got false")
	}

	// bool variable false
	result, err = ev.EvalWithRawEnv("$flag", map[string]string{"$flag": "false"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result {
		t.Errorf("expected false, got true")
	}
}

func TestBuildEnv(t *testing.T) {
	ev := NewExpressionEvaluator()

	// string -> int64
	env := ev.buildEnv(map[string]string{"$count": "42"})
	if v, ok := env["count"].(int64); !ok || v != 42 {
		t.Errorf("expected int64 42, got %T %v", env["count"], env["count"])
	}

	// string -> float64
	env = ev.buildEnv(map[string]string{"$price": "3.14"})
	if v, ok := env["price"].(float64); !ok || v != 3.14 {
		t.Errorf("expected float64 3.14, got %T %v", env["price"], env["price"])
	}

	// "true"/"false" -> bool
	env = ev.buildEnv(map[string]string{"$active": "true"})
	if v, ok := env["active"].(bool); !ok || !v {
		t.Errorf("expected true, got %T %v", env["active"], env["active"])
	}

	env = ev.buildEnv(map[string]string{"$active": "false"})
	if v, ok := env["active"].(bool); !ok || v {
		t.Errorf("expected false, got %T %v", env["active"], env["active"])
	}

	// plain string stays string
	env = ev.buildEnv(map[string]string{"$name": "hello"})
	if v, ok := env["name"].(string); !ok || v != "hello" {
		t.Errorf("expected string 'hello', got %T %v", env["name"], env["name"])
	}
}

func TestCheckSafety(t *testing.T) {
	ev := NewExpressionEvaluator()

	// dangerous keywords
	dangerous := []string{"delete", "DROP", "SELECT", "insert", "truncate", "create", "update", "SET", "from", "grant", "call", "execute"}
	for _, kw := range dangerous {
		err := ev.checkSafety(kw)
		if err == nil {
			t.Errorf("expected error for dangerous keyword '%s', got nil", kw)
		}
	}

	// safe expressions should pass
	safe := []string{"days >= 3", "level > 1", "name == 'test'", "a and b", "x or y"}
	for _, expr := range safe {
		err := ev.checkSafety(expr)
		if err != nil {
			t.Errorf("expected no error for safe expression '%s', got: %v", expr, err)
		}
	}
}

func TestStripVarPrefix(t *testing.T) {
	ev := NewExpressionEvaluator()

	// single variable
	result := ev.stripVarPrefix("$days>=3")
	if result != "days>=3" {
		t.Errorf("expected 'days>=3', got '%s'", result)
	}

	// multiple variables
	result = ev.stripVarPrefix("$a > $b")
	if result != "a > b" {
		t.Errorf("expected 'a > b', got '%s'", result)
	}

	// no variables
	result = ev.stripVarPrefix("3 + 5")
	if result != "3 + 5" {
		t.Errorf("expected '3 + 5', got '%s'", result)
	}
}
