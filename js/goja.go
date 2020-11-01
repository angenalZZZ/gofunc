package js

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/jmoiron/sqlx"
)

func Console(r *goja.Runtime) {
	console := r.NewObject()

	_ = console.Set("log", func(c goja.FunctionCall) goja.Value {
		values := make([]interface{}, 0, len(c.Arguments))
		for _, a := range c.Arguments {
			values = append(values, a.Export())
		}
		fmt.Printf("    console.log: %+v\n", values...)
		return goja.Undefined()
	})

	r.Set("console", console)
}

func Db(r *goja.Runtime, d *sqlx.DB) {
	db := r.NewObject()

	// db.q('select * from table1 where id=?',1)
	// db.q('select * from table1 where id=:id',{id:1})
	_ = db.Set("q", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l == 0 {
			return v
		}

		var (
			sql      = c.Arguments[0].String()
			rows     *sqlx.Rows
			err      error
			value    map[string]interface{}
			hasValue bool
		)

		if l == 2 {
			value, hasValue = c.Arguments[1].Export().(map[string]interface{})
		}

		if hasValue {
			if rows, err = d.NamedQuery(sql, value); err != nil {
				return r.ToValue(err)
			}
		} else {
			values := make([]interface{}, 0, l-1)
			if l > 1 {
				for _, a := range c.Arguments[1:] {
					values = append(values, a.Export())
				}
			}
			if rows, err = d.Queryx(sql, values...); err != nil {
				return r.ToValue(err)
			}
		}

		for rows.Next() {
			result := make(map[string]interface{})
			if err = rows.MapScan(result); err != nil {
				return r.ToValue(err)
			}
			v = r.ToValue(result)
		}

		return v
	})

	r.Set("db", db)
}
