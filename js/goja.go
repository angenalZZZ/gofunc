package js

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/angenalZZZ/gofunc/f"
	ht "github.com/angenalZZZ/gofunc/http"
	"github.com/dop251/goja"
	"github.com/go-resty/resty/v2"
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
// 	db.q: return ResultObject or Array of all rows
// 	db.q('select * from table1 where id=?',1)
// 	db.q('select * from table1 where id=:id',{id:1})
// 	db.g: return ResultValue of first column in first row
// 	db.g('select name from table1 where id=?',1)
// 	db.g('select name from table1 where id=:id',{id:1})
// 	db.i: return LastInsertId must int in number-id-column
// 	db.i('insert into table1 values(?,?)',1,'test')
// 	db.i('insert into table1 values(:id,:name)',{id:1,name:'test'})
//  db.x: return RowsAffected all inserted,updated,deleted
//  db.x('update table1 set name=? where id=?','test',1)
//  db.x('update table1 set name=:name where id=:id',{id:1,name:'test'})
func Db(r *goja.Runtime, d *sqlx.DB) {
	dbObj := r.NewObject()

	driver := make(map[string]interface{})
	driver["name"] = d.DriverName()
	_ = dbObj.Set("driver", driver)

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

		results := make([]map[string]interface{}, 0)
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
			results = append(results, result)
		}

		if len(results) == 1 {
			v = r.ToValue(results[0])
		} else {
			v = r.ToValue(results)
		}

		return v
	})

	_ = dbObj.Set("g", func(c goja.FunctionCall) goja.Value {
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
			if cols, err := rows.Columns(); err != nil {
				v = r.ToValue(result)
			} else if len(cols) > 0 {
				v = r.ToValue(result[cols[0]])
			}
			break
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

// Nats nats in javascript.
//  console.log(nats.subject)
// 	nats.pub('data'); nats.pub('subj','data')
// 	nats.req('data'); nats.pub('data',3); nats.pub('subj','data',3)
func Nats(r *goja.Runtime, nc *nats.Conn, subj string) {
	natsObj := r.NewObject()

	_ = natsObj.Set("name", nc.Opts.Name)
	_ = natsObj.Set("subject", subj)

	_ = natsObj.Set("pub", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l == 1 && subj != "" {
			data := c.Arguments[0].String()
			if err := nc.Publish(subj, f.Bytes(data)); err != nil {
				return r.ToValue(err)
			}
			return r.ToValue(0)
		} else if l == 2 {
			subj, data := c.Arguments[0].String(), c.Arguments[1].String()
			if err := nc.Publish(subj, f.Bytes(data)); err != nil {
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
			msg, err := nc.Request(subj, f.Bytes(data), 3*time.Second)
			if err != nil {
				return r.ToValue(err)
			}
			if msg.Data == nil {
				return v
			}
			return r.ToValue(string(msg.Data))
		} else if l == 2 && subj != "" {
			data, ms := c.Arguments[0].String(), c.Arguments[1].ToInteger()
			msg, err := nc.Request(subj, f.Bytes(data), time.Duration(ms)*time.Second)
			if err != nil {
				return r.ToValue(err)
			}
			if msg.Data == nil {
				return v
			}
			return r.ToValue(string(msg.Data))
		} else if l == 3 {
			subj, data, ms := c.Arguments[0].String(), c.Arguments[1].String(), c.Arguments[2].ToInteger()
			msg, err := nc.Request(subj, f.Bytes(data), time.Duration(ms)*time.Second)
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

// Ajax $ in javascript.
// 	$.q("get",url,data,"",function(res,statusCode)
// 	$.q("post",url,data,"json",function(res,statusCode)
func Ajax(r *goja.Runtime) {
	jObj := r.NewObject()

	_ = jObj.Set("trace", false)
	_ = jObj.Set("token", "")
	_ = jObj.Set("body", "")
	_ = jObj.Set("header", make(map[string]interface{}))
	_ = jObj.Set("cookie", make(map[string]interface{}))

	var trace = func(req *resty.Request) {
		if jObj.Get("trace").ToBoolean() {
			fmt.Printf("\r\n\t--- %s: %s \r\n\t%+v \r\n", req.Method, req.URL, req.TraceInfo())
		}
	}

	var setReq = func(req *resty.Request, data interface{}) {
		if header, ok := jObj.Get("header").Export().(map[string]interface{}); ok {
			for name, val := range header {
				req.SetHeader(name, f.ToString(val))
			}
		}
		if cookie, ok := jObj.Get("cookie").Export().(map[string]interface{}); ok {
			for name, val := range cookie {
				req.SetCookie(&http.Cookie{Name: name, Value: f.ToString(val)})
			}
		}

		if data != nil {
			switch tVal := data.(type) {
			case string:
				req.SetBody([]byte(tVal))
			default:
				if buf, err := json.Marshal(tVal); err == nil {
					req.SetBody(buf)
				}
			}
		} else {
			if body := jObj.Get("body").String(); body != "" {
				req.SetBody(body)
			}
		}

		if jObj.Get("trace").ToBoolean() {
			req.EnableTrace()
		}
	}

	var setRes = func(res *resty.Response) {
		if cookie, ok := jObj.Get("cookie").Export().(map[string]interface{}); ok {
			for _, cc := range res.Cookies() {
				cookie[cc.Name] = cc.Value
			}
			_ = jObj.Set("cookie", cookie)
		}
	}

	_ = jObj.Set("q", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 5 {
			return v
		}

		method, url := c.Arguments[0].String(), c.Arguments[1].String()
		if method == "" || url == "" {
			return v
		}

		var fn func(map[string]interface{}, int)
		if l == 5 {
			if err := r.ExportTo(c.Arguments[4], &fn); err != nil {
				return r.ToValue(err)
			}
		}

		req := ht.NewRestRequest()
		data, contentType := c.Arguments[2].Export(), c.Arguments[3].String()
		if strings.Contains(contentType, "json") {
			req.SetHeader("Content-Type", "application/json")
		} else if strings.Contains(contentType, "url") {
			req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
		} else if len(contentType) > 10 {
			req.SetHeader("Content-Type", contentType)
		}
		if token := jObj.Get("token").String(); token != "" {
			req.SetAuthToken(token)
		}
		setReq(req, data)
		method = strings.ToUpper(method)
		result := make(map[string]interface{})
		res, err := req.Execute(method, url)
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
		trace(req)
		setRes(res)
		buf := res.Body()
		if buf == nil || statusCode >= 400 {
			fn(result, statusCode)
			return v
		}

		buf = bytes.TrimSpace(buf)
		st0, st1 := buf[0], buf[len(buf)-1]
		if st0 == '[' && st1 == ']' {
			var records []map[string]interface{}
			if err := json.Unmarshal(buf, &records); err == nil {
				result["data"] = records
			} else {
				result["data"] = f.String(buf)
			}
		} else if st0 == '{' || st1 == '}' {
			var record map[string]interface{}
			if err := json.Unmarshal(buf, &record); err == nil {
				result["data"] = record
			} else {
				result["data"] = f.String(buf)
			}
		} else {
			result["data"] = f.String(buf)
		}

		fn(result, statusCode)
		return v
	})

	r.Set("$", jObj)
}
