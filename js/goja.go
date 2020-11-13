package js

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/angenalZZZ/gofunc/data/cache/fastcache"
	"github.com/angenalZZZ/gofunc/data/id"
	"github.com/angenalZZZ/gofunc/data/random"
	"github.com/angenalZZZ/gofunc/f"
	ht "github.com/angenalZZZ/gofunc/http"
	"github.com/angenalZZZ/gofunc/log"
	"github.com/dop251/goja"
	"github.com/go-redis/redis/v7"
	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	json "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
)

// Logger use log in javascript.
// 	log.debug('%d', 123)
// 	log.info('%v', {name:'hello'})
// 	log.warn()
// 	log.error()
// 	log.fatal()
// 	log.panic()
// 	log.log()
func Logger(r *goja.Runtime, log *log.Logger) {
	logObj := r.NewObject()

	// log.debug output log
	_ = logObj.Set("debug", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Undefined(), len(c.Arguments)
		if l == 0 {
			return v
		}
		format, s := c.Arguments[0].String(), make([]interface{}, l-1)
		for i := 1; i < l; i++ {
			s = append(s, c.Arguments[i].Export())
		}
		log.Debug().Msgf(format, s...)
		return v
	})
	// log.info output log
	_ = logObj.Set("info", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Undefined(), len(c.Arguments)
		if l == 0 {
			return v
		}
		format, s := c.Arguments[0].String(), make([]interface{}, l-1)
		for i := 1; i < l; i++ {
			s = append(s, c.Arguments[i].Export())
		}
		log.Info().Msgf(format, s...)
		return v
	})
	// log.warn output log
	_ = logObj.Set("warn", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Undefined(), len(c.Arguments)
		if l == 0 {
			return v
		}
		format, s := c.Arguments[0].String(), make([]interface{}, l-1)
		for i := 1; i < l; i++ {
			s = append(s, c.Arguments[i].Export())
		}
		log.Warn().Msgf(format, s...)
		return v
	})
	// log.error output log
	_ = logObj.Set("error", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Undefined(), len(c.Arguments)
		if l == 0 {
			return v
		}
		format, s := c.Arguments[0].String(), make([]interface{}, l-1)
		for i := 1; i < l; i++ {
			s = append(s, c.Arguments[i].Export())
		}
		log.Error().Msgf(format, s...)
		return v
	})
	// log.fatal output log
	_ = logObj.Set("fatal", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Undefined(), len(c.Arguments)
		if l == 0 {
			return v
		}
		format, s := c.Arguments[0].String(), make([]interface{}, l-1)
		for i := 1; i < l; i++ {
			s = append(s, c.Arguments[i].Export())
		}
		log.Fatal().Msgf(format, s...)
		return v
	})
	// log.panic output log
	_ = logObj.Set("panic", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Undefined(), len(c.Arguments)
		if l == 0 {
			return v
		}
		format, s := c.Arguments[0].String(), make([]interface{}, l-1)
		for i := 1; i < l; i++ {
			s = append(s, c.Arguments[i].Export())
		}
		log.Panic().Msgf(format, s...)
		return v
	})
	// log.log output log
	_ = logObj.Set("log", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Undefined(), len(c.Arguments)
		if l == 0 {
			return v
		}
		format, s := c.Arguments[0].String(), make([]interface{}, l-1)
		for i := 1; i < l; i++ {
			s = append(s, c.Arguments[i].Export())
		}
		log.Log().Msgf(format, s...)
		return v
	})

	r.Set("log", logObj)
}

// Console use console.log,dump in javascript.
func Console(r *goja.Runtime) {
	consoleObj := r.NewObject()

	// console.log output content
	_ = consoleObj.Set("log", func(c goja.FunctionCall) goja.Value {
		fmt.Printf("    console.log:")
		for _, a := range c.Arguments {
			if v := a.Export(); v == nil {
				fmt.Print(" null")
			} else {
				fmt.Printf(" %+v", v)
			}
		}
		fmt.Println()
		return goja.Undefined()
	})

	r.Set("console", consoleObj)

	// dump output content
	r.Set("dump", func(c goja.FunctionCall) goja.Value {
		l := len(c.Arguments) - 1
		fmt.Println()
		for i, a := range c.Arguments {
			if v := a.Export(); v == nil {
				fmt.Print(" null")
			} else {
				fmt.Printf(" %+v", v)
			}
			if i < l {
				fmt.Println()
			}
		}
		fmt.Println()
		return goja.Undefined()
	})
}

// ID create a new random ID in javascript.
// 	ID(): return a new random UUID.
//  ID(9),ID(10),ID(20),ID(32),ID(36)
func ID(r *goja.Runtime) {
	r.Set("ID", func(c goja.FunctionCall) goja.Value {
		var l int64 = 36
		if len(c.Arguments) > 0 {
			l = c.Arguments[0].ToInteger()
		}
		switch l {
		case 9:
			return r.ToValue(id.L9())
		case 10:
			return r.ToValue(id.L10())
		case 20:
			return r.ToValue(id.L20())
		case 32:
			return r.ToValue(id.L32())
		case 36:
			return r.ToValue(id.L36())
		default:
			return r.ToValue(id.L36())
		}
	})
}

// RD create a new random string in javascript.
// 	RD(): return a new random string.
//  RD(3),ID(4),ID(5),ID(6),ID(7)...
func RD(r *goja.Runtime) {
	r.Set("RD", func(c goja.FunctionCall) goja.Value {
		var l int64 = 6
		if len(c.Arguments) > 0 {
			l = c.Arguments[0].ToInteger()
			if l < 2 {
				l = 2
			} else if l > 2000 {
				l = 2000
			}
		}
		return r.ToValue(random.AlphaNumber(int(l)))
	})
}

// Db use database and execute sql in javascript.
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

// Nats use nats in javascript.
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

// Ajax use $ in javascript.
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
			cookies := make(map[string]interface{})
			for _, cookie := range req.Cookies {
				name, val := cookie.Name, cookie.Value
				cookies[name] = val
			}
			dump := "\r\n---- %s: %s\r\n---- trace: %+v\r\n---- request-header: %+v\r\n---- request-cookie: %+v\r\n---- request-body: %s\r\n---- response-body: %s\r\n---- response-cookie: %+v\r\n---- response-result: %+v\r\n"
			fmt.Printf(dump, req.Method, req.URL, req.TraceInfo(), jObj.Get("header").Export(), cookies, req.Body, res.Body(), jObj.Get("cookie").Export(), result)
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
						name, val := strings.TrimSpace(str[0]), strings.TrimSpace(str[1])
						req.SetCookie(&http.Cookie{Name: name, Value: val})
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

// Cache use cache in javascript.
// 	dump(cache.dir, cache.cap)
//  var val = cache.get("key")
//  var has = cache.has("key")
//  cache.set("key",123)
//  cache.del("key")
//  cache.reset(); cache.clear(); cache.clear('cache-01');
//  try { cache.save('cache-01'); cache.load('cache-01'); } catch (e) { throw(e) }
func Cache(r *goja.Runtime, cache *fastcache.Cache, cacheDir string, maxBytes ...int) {
	var err error
	// default directory
	currentDir := f.CurrentDir()
	defaultDir := filepath.Join(currentDir, ".nats")
	if cacheDir == "" {
		cacheDir = defaultDir
	}
	// creates a fast cache instance
	capacity := 1073741824 // 1GB cache capacity
	if cache == nil {
		if len(maxBytes) > 0 {
			capacity = maxBytes[0]
		}
		cache = fastcache.New(capacity)
	}

	cObj := r.NewObject()

	_ = cObj.Set("dir", cacheDir)
	_ = cObj.Set("cap", capacity)

	_ = cObj.Set("get", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 1 {
			return v
		}
		key := c.Arguments[0].String()
		if key == "" {
			return v
		}

		p := cache.Get(nil, f.Bytes(key))
		if p == nil || len(p) == 0 {
			return v
		}

		var val interface{}
		if err := json.Unmarshal(p, &val); err == nil {
			v = r.ToValue(val)
		}

		return v
	})

	_ = cObj.Set("set", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 2 {
			return v
		}
		key := c.Arguments[0].String()
		if key == "" {
			return v
		}

		val := c.Arguments[1].Export()
		if p, err := json.Marshal(val); err != nil {
			cache.Set(f.Bytes(key), []byte{})
		} else {
			cache.Set(f.Bytes(key), p)
		}

		return v
	})

	_ = cObj.Set("del", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Undefined(), len(c.Arguments)
		if l < 1 {
			return v
		}
		key := c.Arguments[0].String()
		if key == "" {
			return v
		}

		cache.Del(f.Bytes(key))

		return v
	})

	_ = cObj.Set("has", func(c goja.FunctionCall) goja.Value {
		v, l := r.ToValue(false), len(c.Arguments)
		if l < 1 {
			return v
		}
		key := c.Arguments[0].String()
		if key == "" {
			return v
		}

		p := cache.Has(f.Bytes(key))
		v = r.ToValue(p)

		return v
	})

	_ = cObj.Set("reset", func(c goja.FunctionCall) goja.Value {
		cache.Reset()
		return goja.Undefined()
	})
	_ = cObj.Set("clear", func(c goja.FunctionCall) goja.Value {
		cache.Reset()

		l, dir := len(c.Arguments), ""
		if l > 0 {
			dir = filepath.Join(currentDir, c.Arguments[0].String())
		}
		if dir == "" {
			dir = cacheDir
		}
		if f.IsDir(dir) {
			if err := os.RemoveAll(dir); err != nil {
				panic(err.Error())
			}
		}
		return goja.Undefined()
	})

	_ = cObj.Set("load", func(c goja.FunctionCall) goja.Value {
		l, dir := len(c.Arguments), ""
		if l > 0 {
			dir = filepath.Join(currentDir, c.Arguments[0].String())
		}
		if dir == "" {
			dir = cacheDir
			if f.PathExists(dir) == false {
				return goja.Undefined()
			}
		}
		if f.PathExists(dir) == false {
			panic("The specified directory does not exist")
		}
		if cache, err = fastcache.LoadFromFile(dir); err != nil {
			panic(err.Error())
		}
		return goja.Undefined()
	})

	_ = cObj.Set("save", func(c goja.FunctionCall) goja.Value {
		l, dir := len(c.Arguments), ""
		if l > 0 {
			dir = filepath.Join(currentDir, c.Arguments[0].String())
			// creates a new directory if does not exist
			if err = f.Mkdir(dir); err != nil {
				panic(err.Error())
			}
		}
		if dir == "" {
			dir = cacheDir
			// creates a new directory if does not exist
			if err = f.Mkdir(dir); err != nil {
				panic(err.Error())
			}
		}
		if f.PathExists(dir) == false {
			panic("The specified directory does not exist")
		}
		if err = cache.SaveToFileConcurrent(dir, 0); err != nil {
			panic(err.Error())
		}
		return goja.Undefined()
	})

	r.Set("cache", cObj)
}

// Redis use redis in javascript.
// 	redis.get(key)
// 	redis.del(key,key1,key2)
// 	redis.set(key,value,86400) // 1 days
// 	redis.setNX(key,value,86400)
// 	redis.incr(key), incr(key,2)
// 	redis.lpush(key,1,2,3)
// 	redis.rpush(key,1,2,3)
// 	redis.sort(key,0,10,'asc')
// 	redis.list(key,0,10)
// 	redis.do('SET', key, value)
// 	redis.eval('...')
//  http://www.runoob.com/redis/redis-tutorial.html
func Redis(r *goja.Runtime, client *redis.Client) {
	rObj := r.NewObject()

	// GET key
	_ = rObj.Set("get", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 1 {
			return v
		}

		res, err := client.Get(c.Arguments[0].String()).Result()
		if err == redis.Nil {
			return v
		} else if err != nil {
			return r.ToValue(err)
		}

		return r.ToValue(res)
	})

	// TTL key
	_ = rObj.Set("ttl", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 1 {
			return v
		}

		res, err := client.TTL(c.Arguments[0].String()).Result()
		if err == redis.Nil {
			return v
		} else if err != nil {
			return r.ToValue(err)
		}

		return r.ToValue(res)
	})

	// DEL key
	_ = rObj.Set("del", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 1 {
			return v
		}

		args := make([]string, 0, l)
		for _, a := range c.Arguments {
			args = append(args, a.String())
		}
		res, err := client.Del(args...).Result()
		if err == redis.Nil {
			return v
		} else if err != nil {
			return r.ToValue(err)
		}

		return r.ToValue(res)
	})

	// SET key value EX 10
	_ = rObj.Set("set", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 3 {
			return v
		}

		res, err := client.Set(c.Arguments[0].String(), c.Arguments[1].Export(), time.Duration(c.Arguments[2].ToInteger())*time.Second).Result()
		if err != nil {
			return r.ToValue(err)
		}

		return r.ToValue(res)
	})

	// SET key value EX 10 NX
	_ = rObj.Set("setNX", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 3 {
			return v
		}

		res, err := client.SetNX(c.Arguments[0].String(), c.Arguments[1].Export(), time.Duration(c.Arguments[2].ToInteger())*time.Second).Result()
		if err != nil {
			return r.ToValue(err)
		}

		return r.ToValue(res)
	})

	// INCR key, IncrBy key 10
	_ = rObj.Set("incr", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 1 {
			return v
		}

		if l == 1 {
			res, err := client.Incr(c.Arguments[0].String()).Result()
			if err != nil {
				return r.ToValue(err)
			}
			return r.ToValue(res)
		}

		res, err := client.IncrBy(c.Arguments[0].String(), c.Arguments[1].ToInteger()).Result()
		if err != nil {
			return r.ToValue(err)
		}
		return r.ToValue(res)
	})

	// LPUSH list 1 10 100
	_ = rObj.Set("lpush", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 2 {
			return v
		}

		args := make([]interface{}, 0, l-1)
		for i, a := range c.Arguments {
			if i == 0 {
				continue
			}
			args = append(args, a.Export())
		}
		res, err := client.LPush(c.Arguments[0].String(), args...).Result()
		if err != nil {
			return r.ToValue(err)
		}

		return r.ToValue(res)
	})

	// RPUSH list 1 10 100
	_ = rObj.Set("rpush", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 2 {
			return v
		}

		args := make([]interface{}, 0, l-1)
		for i, a := range c.Arguments {
			if i == 0 {
				continue
			}
			args = append(args, a.Export())
		}
		res, err := client.RPush(c.Arguments[0].String(), args...).Result()
		if err != nil {
			return r.ToValue(err)
		}

		return r.ToValue(res)
	})

	// SORT list LIMIT 0 2 ASC
	_ = rObj.Set("sort", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 4 {
			return v
		}

		sort := redis.Sort{Offset: c.Arguments[1].ToInteger(), Count: c.Arguments[2].ToInteger(), Order: strings.ToUpper(c.Arguments[3].String())}
		res, err := client.Sort(c.Arguments[0].String(), &sort).Result()
		if err == redis.Nil {
			return v
		} else if err != nil {
			return r.ToValue(err)
		}

		return r.ToValue(res)
	})

	// GetRange list 0 10
	_ = rObj.Set("list", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 3 {
			return v
		}

		res, err := client.GetRange(c.Arguments[0].String(), c.Arguments[1].ToInteger(), c.Arguments[2].ToInteger()).Result()
		if err == redis.Nil {
			return v
		} else if err != nil {
			return r.ToValue(err)
		}

		return r.ToValue(res)
	})

	_ = rObj.Set("do", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 2 {
			return v
		}

		args := make([]interface{}, 0, l)
		for _, a := range c.Arguments {
			args = append(args, a.Export())
		}

		res, err := client.Do(args...).Result()
		if err == redis.Nil {
			return v
		} else if err != nil {
			return r.ToValue(err)
		}

		return r.ToValue(res)
	})

	_ = rObj.Set("eval", func(c goja.FunctionCall) goja.Value {
		v, l := goja.Null(), len(c.Arguments)
		if l < 2 {
			return v
		}

		var script string
		keys, args := make([]string, 0, l-1), make([]interface{}, 0, l-2)
		for i, a := range c.Arguments {
			if i == 0 {
				script = a.String()
				continue
			}
			switch tVal := a.Export().(type) {
			case string:
				keys = append(keys, tVal)
			default:
				args = append(args, tVal)
			}
		}
		if script == "" {
			return v
		}

		res, err := client.Eval(script, keys, args...).Result()
		if err == redis.Nil {
			return v
		} else if err != nil {
			return r.ToValue(err)
		}

		return r.ToValue(res)
	})

	r.Set("redis", rObj)
}
