package entity

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// LocalTime 自定义时间类型，用于 GORM 中 DATETIME 字段的格式化处理。
//
// GORM 中使用 time.Time 类型查询返回的格式类似 2023-09-19T14:41:28+08:00，
// 对前端处理不友好，因此使用自定义类型对时间格式进行格式化处理。
type LocalTime time.Time

// Now 返回当前本地时间对应的 LocalTime。
func Now() LocalTime {
	return LocalTime(time.Now())
}

// MarshalJSON 实现 json.Marshaler 接口，将时间序列化为 "2006-01-02 15:04:05" 格式。
func (t *LocalTime) MarshalJSON() ([]byte, error) {
	tTime := time.Time(*t)
	return fmt.Appendf(nil, "\"%v\"", tTime.Format("2006-01-02 15:04:05")), nil
}

// Value 实现 driver.Valuer 接口，将 LocalTime 转换为数据库可接受的值。
// 如果时间为零值，则返回 nil。
func (t LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt, nil
}

// Scan 实现 sql.Scanner 接口，将数据库中的值扫描为 LocalTime。
func (t *LocalTime) Scan(v any) error {
	if value, ok := v.(time.Time); ok {
		*t = LocalTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

// NowInstance 返回当前本地时间对应的 LocalTime（实例方法）。
func (t *LocalTime) NowInstance() LocalTime {
	return LocalTime(time.Now())
}

// String 将 LocalTime 格式化为 "2006-01-02 15:04:05" 字符串。
// 如果接收者为 nil，则返回空字符串。
func (t *LocalTime) String() string {
	if t == nil {
		return ""
	}
	return time.Time(*t).Format("2006-01-02 15:04:05")
}
