package js

import (
	"encoding/xml"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/angenalZZZ/gofunc/data"

	"github.com/angenalZZZ/gofunc/configfile"
	"github.com/angenalZZZ/gofunc/data/cache/store"
	"github.com/angenalZZZ/gofunc/data/id"
	"github.com/angenalZZZ/gofunc/data/random"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/dop251/goja"
	"github.com/go-redis/redis/v7"
	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func TestNewObject(t *testing.T) {
	vm := goja.New()
	defer func() { vm.ClearInterrupt() }()

	newObj := func(c goja.ConstructorCall) *goja.Object {
		_ = c.This.Set("name", c.Argument(0).String())
		_ = c.This.Set("display", func(c1 goja.FunctionCall) goja.Value {
			fmt.Println("    display: my name is", c.This.Get("name").Export())
			return goja.Undefined()
		})
		return nil
	}

	vm.Set("Obj", newObj)

	if v, err := vm.RunString(`obj = new Obj('GO'); obj.name`); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("obj.name = %q", v.Export())
	}

	vm.ClearInterrupt()

	if _, err := vm.RunString(`obj.display()`); err != nil {
		t.Fatal(err)
	}
}

func TestConsole(t *testing.T) {
	vm := goja.New()
	defer func() { vm.ClearInterrupt() }()
	Console(vm)

	if v, err := vm.RunString(`console.log('hello world')`); err != nil {
		t.Fatal(err)
	} else if !v.Equals(goja.Undefined()) {
		t.Fail()
	}

	if buf, err := xml.Marshal(&resty.User{Username: "Hi", Password: "***"}); err != nil {
		t.Fatal(err)
	} else if _, err := vm.RunString(`dump('` + f.String(buf) + `')`); err != nil {
		t.Fatal(err)
	}
}

func TestID(t *testing.T) {
	vm := goja.New()
	defer func() { vm.ClearInterrupt() }()
	Console(vm)
	ID(vm)
	RD(vm)

	shareObject := make(map[string]interface{})
	shareObject["id"] = id.L36()
	shareObject["name"] = random.AlphaNumber(10)
	vm.Set("shareObject", shareObject)

	script := `
dump(shareObject)
shareObject.id = ID()
shareObject.rd = RD()
shareObject.f1 = function (a) { console.log('this.id =', this.id, ', arguments =', a) };
shareObject.f1(111)
`

	if _, err := vm.RunString(script); err != nil {
		t.Fatal(err)
	} else {
		self := vm.Get("shareObject")
		if obj, ok := self.Export().(map[string]interface{}); ok {
			t.Logf("%+v", obj)
			if f1, ok := obj["f1"].(func(goja.FunctionCall) goja.Value); ok {
				f1(goja.FunctionCall{This: self, Arguments: []goja.Value{vm.ToValue(222)}})
			}
		}
	}
}

func TestDb(t *testing.T) {
	r := goja.New()
	defer func() { r.ClearInterrupt() }()
	Console(r)
	Cache(r, nil, filepath.Join(data.RootDir, ".nats01"))
	store.RedisClient = redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	Redis(r, store.RedisClient)

	config, err := configfile.YamlToMap("../test/config/database.yaml")
	if err != nil {
		t.Fatal(err)
	}
	conn, ok := config["database"].(map[interface{}]interface{})
	if !ok {
		t.SkipNow()
	}

	var d *sqlx.DB
	if d, err = sqlx.Connect("mysql", conn["mysql"].(string)); err != nil {
		t.Skip(err)
	}
	defer func() { _ = d.Close() }()

	Db(r, d)

	var ID string
	var newID int64
	script := `db.i("insert into logtest(Code,Type,Message,Account,CreateTime) values(?,?,?,?,?)",'test',2,'new message','admin','2020-10-10 10:20:30')`
	if res, err := r.RunString(script); err != nil {
		t.Fatal(err)
	} else {
		newID = res.ToInteger()
		ID = f.ToString(newID)
		t.Logf("inserted rows Id: %d", newID)
	}

	script = `console.log('update rows affected: '+db.x("update logtest set Code=? where Id=?",'200',` + ID + "))"
	if _, err := r.RunString(script); err != nil {
		t.Fatal(err)
	}

	script = `console.log('select rows: '+JSON.stringify(db.q2(2,"select * from logtest where Id=?",` + ID + ")))"
	if _, err := r.RunString(script); err != nil {
		t.Fatal(err)
	}
	script = `console.log('select rows: '+JSON.stringify(db.q2(0,"select * from logtest where Id=?",` + ID + ")))"
	if _, err := r.RunString(script); err != nil {
		t.Fatal(err)
	}

	script = `console.log('delete rows affected: '+db.x("delete from logtest where Id=?",` + ID + "))"
	if _, err := r.RunString(script); err != nil {
		t.Fatal(err)
	}
}

func TestAjax(t *testing.T) {
	r := goja.New()
	defer func() { r.ClearInterrupt() }()
	Console(r)
	Ajax(r)

	script := `
dump($.header, $.user, $.trace)
var res = $.q("get","https://postman-echo.com/time/now")
dump(res)
res = $.q("post","https://postman-echo.com/post","hello","text")
dump(res)
$.trace = true
res = $.q("post","https://postman-echo.com/post",{strange:'boom'},"url")
dump(res)
`

	if _, err := r.RunString(script); err != nil {
		t.Fatal(err)
	}
}

func TestCache(t *testing.T) {
	r := goja.New()
	defer func() { r.ClearInterrupt() }()
	Console(r)
	Cache(r, nil, filepath.Join(data.RootDir, ".nats01"))

	script := `
console.log("cache.dir =", cache.dir)
console.log("cache.cap =", cache.cap)
try { cache.load() } catch (e) { throw(e) }
console.log("key =", cache.get("key"))
cache.set("key",123)
console.log("key =", cache.get("key"))
cache.set("key",123456)
try { cache.save(); cache.load(); } catch (e) { throw(e) }
console.log("key =",cache.get("key"))
console.log("ok!")
cache.clear()
`

	if _, err := r.RunString(script); err != nil {
		t.Fatal(err)
	}
}

func TestRedis(t *testing.T) {
	r := goja.New()
	defer func() { r.ClearInterrupt() }()
	Console(r)
	store.RedisClient = redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	Redis(r, store.RedisClient)

	script := `
var k = 'key-123'
console.log(k+' =>', redis.get(k))
redis.setNX(k,'value-123',86400) // 1 days
var res = redis.get(k)
console.log(k+' =>', res, '( ttl =', redis.ttl(k), ')')
if (redis.del(k)) {
  console.log(k, 'is deleted.')
  console.log(k, (!redis.get(k) ? 'has been deleted.' : 'delete operation failed!'))
}
k = 'key-count'
res = redis.incr(k)
console.log(k+' =>', res)
res = redis.incr(k,2)
console.log(k+' =>', res)
`

	if _, err := r.RunString(script); err != nil {
		t.Fatal(err)
	}
}

func TestGoRuntime_loadModules(t *testing.T) {
	Runtime = NewRuntime(nil)
	v, err := Runtime.RunString(`
		var t = require('../test/js/load-modules.js');
		console.log(Object.getOwnPropertyNames(t));
		t.test();
	`)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(v)
	}
}
