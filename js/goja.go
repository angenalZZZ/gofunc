package js

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/jmoiron/sqlx"
)

// Console console.log in javascript.
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

// Db execute sql in javascript.
// 	db.q: return ResultObject
// 	db.q('select * from table1 where id=?',1)
// 	db.q('select * from table1 where id=:id',{id:1})
// 	db.i: return LastInsertId
// 	db.i('insert into table1 values(?,?)',1,'test')
// 	db.i('insert into table1 values(:id,:name)',{id:1,name:'test'})
//  db.x: return RowsAffected
//  db.x('update table1 set name=? where id=?','test',1)
//  db.x('update table1 set name=:name where id=:id',{id:1,name:'test'})
func Db(r *goja.Runtime, d *sqlx.DB) {
	db := r.NewObject()

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
			for k, v := range result {
				if s, ok := v.([]byte); ok {
					result[k] = string(s)
				}
			}
			v = r.ToValue(result)
		}

		return v
	})

	_ = db.Set("i", func(c goja.FunctionCall) goja.Value {
		v, l := r.ToValue(-1), len(c.Arguments)
		if l == 0 {
			return v
		}

		var (
			sql      = c.Arguments[0].String()
			insertId int64
			value    map[string]interface{}
			hasValue bool
		)

		if l == 2 {
			value, hasValue = c.Arguments[1].Export().(map[string]interface{})
		}

		if hasValue {
			if rows, err := d.Exec(sql, value); err != nil {
				return r.ToValue(err)
			} else {
				insertId, _ = rows.LastInsertId()
			}
		} else {
			values := make([]interface{}, 0, l-1)
			if l > 1 {
				for _, a := range c.Arguments[1:] {
					values = append(values, a.Export())
				}
			}
			if rows, err := d.Exec(sql, values...); err != nil {
				return r.ToValue(err)
			} else {
				insertId, _ = rows.LastInsertId()
			}
		}
		v = r.ToValue(insertId)

		return v
	})

	_ = db.Set("x", func(c goja.FunctionCall) goja.Value {
		v, l := r.ToValue(-1), len(c.Arguments)
		if l == 0 {
			return v
		}

		var (
			sql      = c.Arguments[0].String()
			affected int64
			value    map[string]interface{}
			hasValue bool
		)

		if l == 2 {
			value, hasValue = c.Arguments[1].Export().(map[string]interface{})
		}

		if hasValue {
			if rows, err := d.Exec(sql, value); err != nil {
				return r.ToValue(err)
			} else {
				affected, _ = rows.RowsAffected()
			}
		} else {
			values := make([]interface{}, 0, l-1)
			if l > 1 {
				for _, a := range c.Arguments[1:] {
					values = append(values, a.Export())
				}
			}
			if rows, err := d.Exec(sql, values...); err != nil {
				return r.ToValue(err)
			} else {
				affected, _ = rows.RowsAffected()
			}
		}
		v = r.ToValue(affected)

		return v
	})

	r.Set("db", db)
}
