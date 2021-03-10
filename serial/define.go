package serial

import "reflect"

// 错误消息格式
type Error struct {
	Code int
	Msg  string
}

// 类型名称
func typeName(val interface{}) string {
	return reflect.Indirect(reflect.ValueOf(val)).Type().Name()
}
