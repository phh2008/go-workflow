package service

import (
	"context"
	"strings"

	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/pkg"
)

// IsVariable 判断传入字符串是否是变量（是否以 $ 开头）。
func (e *Engine) IsVariable(ctx context.Context, key string) bool {
	_ = ctx
	return strings.HasPrefix(key, "$")
}

// RemoveVarPrefix 去掉变量前缀 "$"。
func (e *Engine) RemoveVarPrefix(ctx context.Context, variable string) string {
	_ = ctx
	return strings.ReplaceAll(variable, "$", "")
}

// ResolveVariables 解析变量，获取并设置其 value，返回 map（非变量则原样存储）。
func (e *Engine) ResolveVariables(ctx context.Context, params model.ResolveVariablesParams) (map[string]string, error) {
	return e.repo.ResolveVariables(ctx, params.InstanceID, params.Variables)
}

// ParseVariable 解析变量，返回变量数组
func (e *Engine) ParseVariable(ctx context.Context, variablesJSON string) ([]model.Variable, error) {
	_ = ctx
	var variables []model.Variable
	if variablesJSON == "" {
		return variables, nil
	}
	if err := pkg.JSONToStruct(variablesJSON, &variables); err != nil {
		return nil, err
	}
	return variables, nil
}

// ParseVariableMap 解析变量，返回变量 map
func (e *Engine) ParseVariableMap(ctx context.Context, variablesJSON string) (map[string]string, error) {
	list, err := e.ParseVariable(ctx, variablesJSON)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, v := range list {
		m[v.Key] = v.Value
	}
	return m, nil
}
