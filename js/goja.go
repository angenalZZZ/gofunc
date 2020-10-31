package js

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func Console(r *goja.Runtime) {
	console := r.NewObject()

	_ = console.Set("log", func(c goja.FunctionCall) goja.Value {
		fmt.Println(c.Arguments)
		return goja.Undefined()
	})

	r.Set("console", console)
}

func Db(r *goja.Runtime, d *gorm.DB) {
	db := r.NewObject()

	_ = db.Set("query", func(c goja.FunctionCall) goja.Value {
		v, args := goja.Null(), c.Arguments
		if len(args) == 0 {
			return v
		}

		var values []interface{}
		sql := args[0].String()
		if len(args) > 1 {
			for _, s := range args[1:] {
				values = append(values, s.Export())
			}
		}

		if res := d.Exec(sql, values); res.Error != nil {
			return r.ToValue(res.Error)
		} else {
			v = r.ToValue(res.Value)
		}

		return v
	})

	r.Set("db", db)
}
