package js

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
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

// Console console.log,dump in javascript.
func Console(r *goja.Runtime) {
	consoleObj := r.NewObject()

	// console.log output content
	_ = consoleObj.Set("log", func(c goja.FunctionCall) goja.Value {
		for _, a := range c.Arguments {
			fmt.Printf("    console.log: %+v\n", a.Export())
		}
		return goja.Undefined()
	})

	r.Set("console", consoleObj)

	// dump output content
	r.Set("dump", func(c goja.FunctionCall) goja.Value {
		fmt.Println()
		for _, a := range c.Arguments {
			fmt.Printf("%+v\n", a.Export())
		}
		fmt.Println()
		return goja.Undefined()
	})
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
//  console.log(nats.name)
//  console.log(nats.subject)
// 	nats.pub('data'); nats.pub('subj','data')
// 	nats.req('data'); nats.pub('data',3); nats.pub('subj','data',3) // timeout:3s
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
// 	dump($.header, $.user, $.trace, $.body, $.cookie, $.token)
// 	var res = $.q("get",url)
// 	var res = $.q("get",url,param)
// 	var res = $.q("post",url,param,"json")
// 	$.q("get",url,param,"",function(data,status))
// 	$.q("post",url,param,"json",function(data,status))
func Ajax(r *goja.Runtime) {
	jObj := r.NewObject()

	header := make(map[string]interface{})
	header["Accept"] = "*/*"
	header["Accept-Language"] = "zh-CN,zh-TW,zh;q=0.9,zh;q=0.8,en;q=0.7"
	header["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36"
	_ = jObj.Set("header", header)

	user := make(map[string]string)
	user["username"] = ""
	user["password"] = ""
	_ = jObj.Set("user", user)
	_ = jObj.Set("body", "")
	_ = jObj.Set("cookie", "")
	_ = jObj.Set("token", "")
	_ = jObj.Set("trace", false)

	var trace = func(req *resty.Request, res *resty.Response, result map[string]interface{}) {
		if jObj.Get("trace").ToBoolean() {
			dump := "\r\n---- %s: %s \r\n%+v\r\n%+v\r\n%+v\r\n\n%+v\r\n\n%+v\r\n\n%s\r\n\n%+v\r\n\n"
			fmt.Printf(dump, req.Method, req.URL, jObj.Get("header").Export(), jObj.Get("cookie").Export(), req.Body, req.TraceInfo(), result, res.Body(), jObj.Get("cookie").Export())
		}
	}

	var setReq = func(req *resty.Request, data interface{}) {
		if jObj.Get("trace").ToBoolean() {
			req.EnableTrace()
		}

		if tObj := jObj.Get("header").Export(); tObj != nil {
			switch tVal := tObj.(type) {
			case map[string]interface{}:
				for name, val := range tVal {
					req.SetHeader(name, f.ToString(val))
				}
			case string:
				for _, line := range strings.Split(strings.TrimSpace(tVal), "\n") {
					if str := strings.Split(strings.TrimSpace(line), ":"); len(str) == 2 {
						req.SetHeader(strings.TrimSpace(str[0]), strings.TrimSpace(str[1]))
					}
				}
			}
		}

		if tObj := jObj.Get("cookie").Export(); tObj != nil {
			switch tVal := tObj.(type) {
			case map[string]interface{}:
				for name, val := range tVal {
					req.SetCookie(&http.Cookie{Name: name, Value: f.ToString(val)})
				}
			case string:
				for _, line := range strings.Split(strings.TrimSpace(tVal), "\n") {
					if str := strings.Split(strings.TrimSpace(line), ":"); len(str) == 2 {
						req.SetCookie(&http.Cookie{Name: strings.TrimSpace(str[0]), Value: strings.TrimSpace(str[1])})
					}
				}
			}
		}

		if data == nil {
			data = jObj.Get("body").Export()
		}
		if data != nil {
			switch tVal := data.(type) {
			case string:
				if tVal != "" {
					req.SetBody([]byte(tVal))
				}
			case map[string]interface{}:
				switch req.Header.Get("Content-Type") {
				case "application/json":
					if buf, err := json.Marshal(tVal); err == nil {
						req.SetBody(buf)
					}
				case "application/xml":
					if buf, err := xml.Marshal(tVal); err == nil {
						req.SetBody(buf)
					}
				default:
					items := make([]string, 0, len(tVal))
					for k, v := range tVal {
						items = append(items, url.QueryEscape(k)+"="+url.QueryEscape(f.ToString(v)))
					}
					req.SetBody(f.Bytes(strings.Join(items, "&")))
				}
			case map[interface{}]interface{}:
				switch req.Header.Get("Content-Type") {
				case "application/json":
					if buf, err := json.Marshal(tVal); err == nil {
						req.SetBody(buf)
					}
				case "application/xml":
					if buf, err := xml.Marshal(tVal); err == nil {
						req.SetBody(buf)
					}
				default:
					items := make([]string, 0, len(tVal))
					for k, v := range tVal {
						items = append(items, url.QueryEscape(f.ToString(k))+"="+url.QueryEscape(f.ToString(v)))
					}
					req.SetBody(f.Bytes(strings.Join(items, "&")))
				}
			default:
				if buf, err := json.Marshal(tVal); err == nil {
					req.SetBody(buf)
				}
			}
		}
	}

	var setRes = func(res *resty.Response) {
		cookie := make(map[string]interface{})
		for _, cc := range res.Cookies() {
			cookie[cc.Name] = cc.Value
		}
		_ = jObj.Set("cookie", cookie)
		_ = jObj.Set("body", "")
	}

	_ = jObj.Set("q", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 2 {
			return v
		}

		method, urlStr := c.Arguments[0].String(), c.Arguments[1].String()
		if method == "" || urlStr == "" {
			return v
		}

		var fn func(map[string]interface{}, int)
		callback := l == 5
		if callback {
			if err := r.ExportTo(c.Arguments[4], &fn); err != nil {
				return r.ToValue(err)
			}
		}

		var (
			cont string
			data interface{}
			req  = ht.NewRestRequest()
		)
		if l > 3 {
			data = c.Arguments[2].Export()
			if err := r.ExportTo(c.Arguments[3], &fn); err == nil {
				callback = true
			} else {
				cont = c.Arguments[3].String()
			}
		} else if l > 2 {
			if err := r.ExportTo(c.Arguments[2], &fn); err == nil {
				callback = true
			} else {
				data = c.Arguments[2].Export()
			}
		}

		if strings.Contains(cont, "json") {
			req.SetHeader("Content-Type", "application/json")
		} else if strings.Contains(cont, "xml") {
			req.SetHeader("Content-Type", "application/xml")
		} else if strings.Contains(cont, "form") || strings.Contains(cont, "url") {
			req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
		} else if strings.Contains(cont, "file") || strings.Contains(cont, "data") {
			req.SetHeader("Content-Type", "multipart/form-data")
		} else if strings.Contains(cont, "text") {
			req.SetHeader("Content-Type", "text/plain")
		} else if len(cont) > 10 {
			req.SetHeader("Content-Type", cont)
		}

		if token := jObj.Get("token").String(); token != "" {
			req.SetAuthToken(token)
		}

		if user, ok := jObj.Get("user").Export().(map[string]string); ok && user != nil && user["username"] != "" {
			req.SetBasicAuth(user["username"], user["password"])
		}

		setReq(req, data)
		method = strings.ToUpper(method)
		result := make(map[string]interface{})
		res, err := req.Execute(method, urlStr)
		if err != nil {
			result["error"] = err.Error()
			result["code"] = -1
			result["data"] = nil
			// request did not occur and could not be traced
			// trace(req, res, result)
			if callback {
				fn(result, -1)
				return v
			}
			return r.ToValue(result)
		}

		statusCode := res.StatusCode()
		result["error"] = res.Status()
		result["code"] = statusCode
		result["data"] = nil
		setRes(res)
		buf := res.Body()
		if buf == nil || statusCode >= 400 {
			trace(req, res, result)
			if callback {
				fn(result, statusCode)
				return v
			}
			return r.ToValue(result)
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

		result["error"] = nil
		trace(req, res, result)
		if callback {
			fn(result, statusCode)
			return v
		}
		return r.ToValue(result)
	})

	r.Set("$", jObj)
}
