package data_test

import (
	"strings"
	"testing"

	"github.com/angenalZZZ/gofunc/configfile"
	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/data/id"
	"github.com/angenalZZZ/gofunc/data/random"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/jmoiron/sqlx"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func TestDbo(t *testing.T) {
	config, err := configfile.YamlToMap("../test/config/database.yaml")
	if err != nil {
		t.Fatal(err)
	}
	conn, ok := config["database"].(map[interface{}]interface{})
	if !ok {
		t.SkipNow()
	}

	data.DbType = "sqlite3"
	data.DbConn = conn[data.DbType].(string)
	data.Dbo, err = sqlx.Open(data.DbType, data.DbConn)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("[%s] %s", data.DbType, data.DbConn)
	defer func() { _ = data.Dbo.Close() }()

	err = data.Dbo.Ping()
	if err != nil {
		t.Fatal(err)
	}

	var num int64
	sql, tbl := `SELECT * FROM sqlite_master WHERE type='table' AND name=?`, "logtest"
	if rows, err := data.Dbo.Queryx(sql, tbl); err != nil {
		t.Fatal(err)
	} else {
		for rows.Next() {
			num++
			dest := make(map[string]interface{})
			_ = rows.MapScan(dest)
			t.Logf("[%s] %q table is exists", data.DbType, dest["name"])
		}
		if num == 0 {
			if buf, err := f.ReadFile("../test/sql/logtest-sqlite.sql"); err != nil {
				t.Fatal(err)
			} else {
				sql = strings.TrimSpace(f.String(buf))
				if res, err := data.Dbo.Exec(sql); err != nil {
					t.Fatal(err)
				} else {
					num, _ = res.RowsAffected()
					t.Logf(`[%s] %q create table, rows affected %d , created by "logtest-sqlite.sql"`, data.DbType, tbl, num)
				}
			}
		}
	}

	sql = `INSERT INTO [logtest](Code,Type,Message,Account,CreateTime,CreateUser) VALUES(?,2,?,?,DATETIME(),?)`
	if res, err := data.Dbo.Exec(sql, random.AlphaNumber(6), random.AlphaNumber(100), random.AlphaNumber(6), id.L36()); err != nil {
		t.Fatal(err)
	} else {
		num, _ = res.LastInsertId()
		t.Logf(`[%s] %q inserted rows [Id=%d]`, data.DbType, tbl, num)
	}
}
