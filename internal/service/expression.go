package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/expr-lang/expr"
)

// ExpressionEvaluator 表达式求值器，使用 expr-lang/expr 库替代原来基于 MySQL 的表达式计算。
type ExpressionEvaluator struct{}

// NewExpressionEvaluator 创建表达式求值器。
func NewExpressionEvaluator() *ExpressionEvaluator {
	return &ExpressionEvaluator{}
}

// Eval 计算布尔表达式。
// expression 为原始表达式，变量使用 $prefix 引用（如 "$days>=3"）。
// env 为变量名到值的映射（已去除 $ 前缀，值类型为 any）。
func (ev *ExpressionEvaluator) Eval(expression string, env map[string]any) (bool, error) {
	// 去除变量名中的 $ 前缀
	cleaned := ev.stripVarPrefix(expression)

	// 检查危险词，防止代码注入
	if err := ev.checkSafety(cleaned); err != nil {
		return false, err
	}

	// 编译表达式
	program, err := expr.Compile(cleaned, expr.Env(env), expr.AsBool())
	if err != nil {
		return false, fmt.Errorf("表达式编译失败: %w", err)
	}

	// 执行求值
	result, err := expr.Run(program, env)
	if err != nil {
		return false, fmt.Errorf("表达式执行失败: %w", err)
	}

	b, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("表达式结果类型不是 bool: %T", result)
	}

	return b, nil
}

// EvalWithRawEnv 便捷方法：接收原始字符串环境，自动处理 $ 前缀和类型转换。
// rawEnv 中 key 使用 $prefix 格式（如 "$days"），value 为字符串。
func (ev *ExpressionEvaluator) EvalWithRawEnv(expression string, rawEnv map[string]string) (bool, error) {
	env := ev.buildEnv(rawEnv)
	return ev.Eval(expression, env)
}

// buildEnv 从字符串值映射构建类型化环境。
// 自动去除 key 的 $ 前缀，并将值从字符串转为 int/float/bool 类型。
func (ev *ExpressionEvaluator) buildEnv(stringEnv map[string]string) map[string]any {
	env := make(map[string]any)
	for k, v := range stringEnv {
		key := strings.TrimPrefix(k, "$")
		env[key] = ev.autoConvert(v)
	}
	return env
}

// autoConvert 尝试将字符串值自动转换为更合适的类型。
func (ev *ExpressionEvaluator) autoConvert(v string) any {
	// 尝试转为 bool
	if v == "true" {
		return true
	}
	if v == "false" {
		return false
	}
	// 尝试转为 int
	if i, err := strconv.ParseInt(v, 10, 64); err == nil {
		return i
	}
	// 尝试转为 float64
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}
	return v
}

// stripVarPrefix 将表达式中 $开头的变量名去除 $ 前缀。
func (ev *ExpressionEvaluator) stripVarPrefix(expression string) string {
	reg := regexp.MustCompile(`\$([a-zA-Z_]\w*)`)
	return reg.ReplaceAllString(expression, `$1`)
}

// checkSafety 检查表达式中是否包含危险操作。
func (ev *ExpressionEvaluator) checkSafety(expression string) error {
	pattern := regexp.MustCompile(`(?i)\b(delete|truncate|insert|drop|create|select|update|set|from|grant|call|execute)\b`)
	match := pattern.FindString(expression)
	if match != "" {
		return fmt.Errorf("表达式中包含危险词: %s", match)
	}
	return nil
}
