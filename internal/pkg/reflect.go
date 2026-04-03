package pkg

import "reflect"

// TypeIsError 判断给定的反射类型是否实现了 error 接口。
// 通过检查类型是否拥有无参数、返回 string 的 Error 方法来判定。
func TypeIsError(t reflect.Type) bool {
	if t.NumMethod() >= 1 {
		for method := range t.Methods() {
			if method.Name == "Error" {
				// 无传入参数
				if method.Type.NumIn() != 0 {
					return false
				}
				// 只有一个输出参数
				if method.Type.NumOut() != 1 {
					return false
				}
				// 输出参数为 string
				if method.Type.Out(0).Kind().String() != "string" {
					return false
				}
				return true
			}
		}
	}

	return false
}
