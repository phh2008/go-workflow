package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Bunny3th/easy-workflow/internal/model"
)

func TestIsVariable(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	if !eng.IsVariable(ctx, "$days") {
		t.Error("expected true for '$days'")
	}
	if eng.IsVariable(ctx, "days") {
		t.Error("expected false for 'days'")
	}
	if eng.IsVariable(ctx, "") {
		t.Error("expected false for empty string")
	}
}

func TestRemoveVarPrefix(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	result := eng.RemoveVarPrefix(ctx, "$days")
	if result != "days" {
		t.Errorf("expected 'days', got '%s'", result)
	}

	result = eng.RemoveVarPrefix(ctx, "$$days")
	if result != "days" {
		t.Errorf("expected 'days', got '%s'", result)
	}

	result = eng.RemoveVarPrefix(ctx, "days")
	if result != "days" {
		t.Errorf("expected 'days', got '%s'", result)
	}
}

func TestResolveVariables(t *testing.T) {
	ctx := context.Background()

	t.Run("resolve variables from repo", func(t *testing.T) {
		repo := &mockRepo{
			ResolveVariablesFunc: func(ctx context.Context, instID int, varNames []string) (map[string]string, error) {
				result := make(map[string]string)
				for _, name := range varNames {
					// strip $ prefix for lookup
					key := name
					if len(key) > 0 && key[0] == '$' {
						key = key[1:]
					}
					result[name] = "value_of_" + key
				}
				return result, nil
			},
		}
		eng := newTestEngine(repo)

		m, err := eng.ResolveVariables(ctx, model.ResolveVariablesParams{InstanceID: 1, Variables: []string{"$days", "literal"}})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if m["$days"] != "value_of_days" {
			t.Errorf("expected 'value_of_days', got '%s'", m["$days"])
		}
		if m["literal"] != "value_of_literal" {
			t.Errorf("expected 'value_of_literal', got '%s'", m["literal"])
		}
	})

	t.Run("missing variable returns error", func(t *testing.T) {
		repo := &mockRepo{
			ResolveVariablesFunc: func(ctx context.Context, instID int, varNames []string) (map[string]string, error) {
				return nil, fmt.Errorf("variable not found")
			},
		}
		eng := newTestEngine(repo)

		_, err := eng.ResolveVariables(ctx, model.ResolveVariablesParams{InstanceID: 1, Variables: []string{"$missing"}})
		if err == nil {
			t.Fatal("expected error for missing variable, got nil")
		}
	})
}

func TestInstanceVariablesSave(t *testing.T) {
	t.Run("valid JSON array", func(t *testing.T) {
		variablesJSON := `[{"Key":"x","Value":"1"},{"Key":"y","Value":"hello"}]`
		var variables []model.Variable
		if err := json.Unmarshal([]byte(variablesJSON), &variables); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(variables) != 2 {
			t.Fatalf("expected 2 variables, got %d", len(variables))
		}
		if variables[0].Key != "x" || variables[0].Value != "1" {
			t.Errorf("unexpected variable[0]: %+v", variables[0])
		}
	})

	t.Run("empty JSON array", func(t *testing.T) {
		variablesJSON := `[]`
		var variables []model.Variable
		if err := json.Unmarshal([]byte(variablesJSON), &variables); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(variables) != 0 {
			t.Errorf("expected 0 variables, got %d", len(variables))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		variablesJSON := `{not valid json}`
		var variables []model.Variable
		if err := json.Unmarshal([]byte(variablesJSON), &variables); err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
	})
}
