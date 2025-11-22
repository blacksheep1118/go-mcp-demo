// internal/db/export.go
package db

import (
	"reflect"

	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/query"
	"gorm.io/gorm"
)

// RawDB 返回底层 *gorm.DB
func RawDB() *gorm.DB {
	// 1. 拿到 *DB[*query.Query]
	wrapper := instance.(*DB[*query.Query])
	// 2. 反射取 query 字段（第一个字段就是 *gorm.DB）
	v := reflect.ValueOf(wrapper.query).Elem()
	return v.Field(0).Interface().(*gorm.DB)
}
