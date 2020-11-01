package js

import (
	"testing"

	"github.com/angenalZZZ/gofunc/configfile"
	"github.com/dop251/goja"
	"github.com/jmoiron/sqlx"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

func TestConsole(t *testing.T) {
	r := goja.New()
	defer func() { r.ClearInterrupt() }()
	Console(r)

	if v, err := r.RunString(`console.log('hello world,', new Date)`); err != nil {
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
	var rows *sqlx.Rows
	sql, values := "SELECT * FROM logtest ORDER BY Id DESC LIMIT ?", []interface{}{1}
	if rows, err = d.Queryx(sql, values...); err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		result := make(map[string]interface{})
		if err = rows.MapScan(result); err != nil {
			t.Fatal(err)
		}
		t.Logf("Max(Id): %v", result["Id"])
	}

	Db(r, d)
	script := `console.log('Max(Id): '+db.q('SELECT * FROM logtest ORDER BY Id DESC LIMIT ?',1).Id)`
	if v, err := r.RunString(script); err != nil {
		t.Fatal(err)
	} else if !v.Equals(goja.Undefined()) {
		t.Fail()
	}
}
