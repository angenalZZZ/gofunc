package js

import (
	"bytes"
	"fmt"
	"time"

	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/http"
	"github.com/dop251/goja"
	"github.com/jmoiron/sqlx"
	json "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
)

// Console console.log in javascript.
func Console(r *goja.Runtime) {
	consoleObj := r.NewObject()

	_ = consoleObj.Set("log", func(c goja.FunctionCall) goja.Value {
		values := make([]interface{}, 0, len(c.Arguments))
		for _, a := range c.Arguments {
			values = append(values, a.Export())
		}
		fmt.Printf("    console.log: %+v\r\n", values...)
		return goja.Undefined()
	})

	r.Set("console", consoleObj)
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
	dbObj := r.NewObject()

	_ = dbObj.Set("driverName", d.DriverName())

	_ = dbObj.Set("q", func(c goja.FunctionCall) goja.Value {
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

	_ = dbObj.Set("i", func(c goja.FunctionCall) goja.Value {
		v, l := r.ToValue(-1), len(c.Arguments)
		if l == 0 {
			return v
		}

		var (
			sql      = c.Arguments[0].String()
			insertID int64
			value    map[string]interface{}
			hasValue bool
		)

		if l == 2 {
			value, hasValue = c.Arguments[1].Export().(map[string]interface{})
		}

		if hasValue {
			rows, err := d.Exec(sql, value)
			if err != nil {
				return r.ToValue(err)
			}
			insertID, _ = rows.LastInsertId()
		} else {
			values := make([]interface{}, 0, l-1)
			if l > 1 {
				for _, a := range c.Arguments[1:] {
					values = append(values, a.Export())
				}
			}
			rows, err := d.Exec(sql, values...)
			if err != nil {
				return r.ToValue(err)
			}
			insertID, _ = rows.LastInsertId()
		}
		v = r.ToValue(insertID)

		return v
	})

	_ = dbObj.Set("x", func(c goja.FunctionCall) goja.Value {
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
			rows, err := d.Exec(sql, value)
			if err != nil {
				return r.ToValue(err)
			}
			affected, _ = rows.RowsAffected()
		} else {
			values := make([]interface{}, 0, l-1)
			if l > 1 {
				for _, a := range c.Arguments[1:] {
					values = append(values, a.Export())
				}
			}
			rows, err := d.Exec(sql, values...)
			if err != nil {
				return r.ToValue(err)
			}
			affected, _ = rows.RowsAffected()
		}
		v = r.ToValue(affected)

		return v
	})

	r.Set("db", dbObj)
}

// Nats nats.pub,nats.req in javascript.
func Nats(r *goja.Runtime, nc *nats.Conn, subj string) {
	natsObj := r.NewObject()

	_ = natsObj.Set("name", nc.Opts.Name)
	_ = natsObj.Set("subj", subj)
	_ = natsObj.Set("subject", subj)

	_ = natsObj.Set("pub", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l == 1 && subj != "" {
			data := c.Arguments[0].String()
			if err := nc.Publish(subj, []byte(data)); err != nil {
				return r.ToValue(err)
			}
			return r.ToValue(0)
		} else if l == 2 {
			subj, data := c.Arguments[0].String(), c.Arguments[1].String()
			if err := nc.Publish(subj, []byte(data)); err != nil {
				return r.ToValue(err)
			}
			return r.ToValue(0)
		}
		return v
	})

	_ = natsObj.Set("req", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l == 1 && subj != "" {
			data := c.Arguments[0].String()
			msg, err := nc.Request(subj, []byte(data), 3*time.Second)
			if err != nil {
				return r.ToValue(err)
			}
			if msg.Data == nil {
				return v
			}
			return r.ToValue(string(msg.Data))
		} else if l == 2 && subj != "" {
			data, ms := c.Arguments[0].String(), c.Arguments[1].ToInteger()
			msg, err := nc.Request(subj, []byte(data), time.Duration(ms)*time.Microsecond)
			if err != nil {
				return r.ToValue(err)
			}
			if msg.Data == nil {
				return v
			}
			return r.ToValue(string(msg.Data))
		} else if l == 3 {
			subj, data, ms := c.Arguments[0].String(), c.Arguments[1].String(), c.Arguments[2].ToInteger()
			msg, err := nc.Request(subj, []byte(data), time.Duration(ms)*time.Microsecond)
			if err != nil {
				return r.ToValue(err)
			}
			if msg.Data == nil {
				return v
			}
			return r.ToValue(string(msg.Data))
		}
		return v
	})

	r.Set("nats", natsObj)
}

// Jquery $.get,getJSON in javascript.
// 	$.get(url,data,function(res,status)
func Jquery(r *goja.Runtime) {
	jObj := r.NewObject()

	_ = jObj.Set("token", "")
	_ = jObj.Set("body", "")
	_ = jObj.Set("header", make(map[string]interface{}))

	_ = jObj.Set("get", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 2 {
			return v
		}

		url, data := c.Arguments[0].String(), c.Arguments[1].Export()
		if url == "" || data == nil {
			return v
		}

		var fn func(map[string]interface{}, int)
		if l == 3 {
			if err := r.ExportTo(c.Arguments[2], &fn); err != nil {
				return r.ToValue(err)
			}
		}

		token, header := jObj.Get("token").String(), jObj.Get("header").Export().(map[string]interface{})
		req := http.NewRestFormRequest(token)
		for name, val := range header {
			req.SetHeader(name, f.ToString(val))
		}
		if body := jObj.Get("body").String(); body != "" {
			req.SetBody(body)
		}

		result := make(map[string]interface{})
		res, err := req.Get(url)
		if err != nil {
			result["error"] = err.Error()
			result["code"] = -1
			result["data"] = nil
			fn(result, -1)
			return v
		}

		statusCode := res.StatusCode()
		result["error"] = res.Status()
		result["code"] = statusCode
		result["data"] = nil
		buf := res.Body()
		if buf == nil || statusCode >= 400 {
			fn(result, statusCode)
			return v
		}

		result["data"] = f.String(buf)

		fn(result, statusCode)
		return v
	})

	_ = jObj.Set("getJSON", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 2 {
			return v
		}

		url, data := c.Arguments[0].String(), c.Arguments[1].Export()
		if url == "" || data == nil {
			return v
		}

		var fn func(map[string]interface{}, int)
		if l == 3 {
			if err := r.ExportTo(c.Arguments[2], &fn); err != nil {
				return r.ToValue(err)
			}
		}

		token, header := jObj.Get("token").String(), jObj.Get("header").Export().(map[string]interface{})
		req := http.NewRestJsonRequest(token)
		for name, val := range header {
			req.SetHeader(name, f.ToString(val))
		}
		if body := jObj.Get("body").String(); body != "" {
			req.SetBody(body)
		}

		result := make(map[string]interface{})
		res, err := req.Get(url)
		if err != nil {
			result["error"] = err.Error()
			result["code"] = -1
			result["data"] = nil
			fn(result, -1)
			return v
		}

		statusCode := res.StatusCode()
		result["error"] = res.Status()
		result["code"] = statusCode
		result["data"] = nil
		buf := res.Body()
		if buf == nil || statusCode >= 400 {
			fn(result, statusCode)
			return v
		}

		buf = bytes.TrimSpace(buf)
		if buf[0] == '[' && buf[len(buf)-1] == ']' {
			var records []map[string]interface{}
			if err := json.Unmarshal(buf, &records); err == nil {
				result["data"] = records
			} else {
				result["data"] = f.String(buf)
			}
		} else {
			var record map[string]interface{}
			if err := json.Unmarshal(buf, &record); err == nil {
				result["data"] = record
			} else {
				result["data"] = f.String(buf)
			}
		}

		fn(result, statusCode)
		return v
	})

	r.Set("$", jObj)
}
