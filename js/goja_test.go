package js

import (
	"testing"

	"github.com/angenalZZZ/gofunc/configfile"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/dop251/goja"
	"github.com/jmoiron/sqlx"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

func TestConsole(t *testing.T) {
	r := goja.New()
	defer func() { r.ClearInterrupt() }()
	Console(r)

	if v, err := r.RunString(`console.log('hello world')`); err != nil {
		t.Fatal(err)
	} else if !v.Equals(goja.Undefined()) {
		t.Fail()
	}
}

func TestDb(t *testing.T) {
	r := goja.New()
	defer func() { r.ClearInterrupt() }()
	Console(r)

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
		t.Fatal(err)
	}
	defer func() { _ = d.Close() }()

	Db(r, d)

	var newId int64
	script := `db.i('insert into logtest(Code,Type,Message,Account,CreateTime) values(?,?,?,?,?)','test',2,'new message','admin','2020-10-10 10:20:30')`
	if res, err := r.RunString(script); err != nil {
		t.Fatal(err)
	} else {
		newId = res.ToInteger()
		t.Logf("inserted rows Id: %d", newId)
	}

	script = `console.log('update rows affected: '+db.x('update logtest set Code=? where Id=?','200',` + f.ToString(newId) + "))"
	if _, err := r.RunString(script); err != nil {
		t.Fatal(err)
	}

	script = `console.log('select rows: '+JSON.stringify(db.q('select * from logtest where Id=?',` + f.ToString(newId) + ")))"
	if _, err := r.RunString(script); err != nil {
		t.Fatal(err)
	}

	script = `console.log('delete rows affected: '+db.x('delete from logtest where Id=?',` + f.ToString(newId) + "))"
	if _, err := r.RunString(script); err != nil {
		t.Fatal(err)
	}
}
