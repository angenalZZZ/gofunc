package data_test

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

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

func initDbSqlite3(t *testing.T) (db *sqlx.DB, err error) {
	var config map[interface{}]interface{}
	config, err = configfile.YamlToMap("../test/config/database.yaml")
	if err != nil {
		return nil, err
	}
	conn, ok := config["database"].(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("database config error")
	}

	data.DbType = "sqlite3"
	dbConn := conn[data.DbType].(string)
	db, err = sqlx.Open(data.DbType, dbConn)
	if err == nil {
		t.Logf("[%s] %s", data.DbType, dbConn)
	}
	return
}

func TestDb_test_sqlite3(t *testing.T) {
	dbo, err := initDbSqlite3(t)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = dbo.Close() }()

	err = dbo.Ping()
	if err != nil {
		t.Fatal(err)
	}

	var num int64
	sql, tbl := `SELECT * FROM sqlite_master WHERE type='table' AND name=?`, "logtest"
	if rows, err := dbo.Queryx(sql, tbl); err != nil {
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
				if res, err := dbo.Exec(sql); err != nil {
					t.Fatal(err)
				} else {
					num, _ = res.RowsAffected()
					t.Logf(`[%s] %q create table, rows affected %d , created by "logtest-sqlite.sql"`, data.DbType, tbl, num)
				}
			}
		}
	}

	sql = `INSERT INTO [logtest](Code,Type,Message,Account,CreateTime,CreateUser) VALUES(?,2,?,?,DATETIME(),?)`
	if res, err := dbo.Exec(sql, random.AlphaNumber(6), random.AlphaNumber(100), random.AlphaNumber(6), id.L36()); err != nil {
		t.Fatal(err)
	} else {
		num, _ = res.LastInsertId()
		t.Logf(`[%s] %q inserted rows [Id=%d]`, data.DbType, tbl, num)
	}
}

func TestBenchDb_insert_sqlite3(t *testing.T) {
	dbs := make([]*sqlx.DB, 2)
	// sets number
	dbN, number := len(dbs), 100

	for i := 0; i < dbN; i++ {
		dbo, err := initDbSqlite3(t)
		if err != nil {
			t.Fatal(err)
		}
		dbs[i] = dbo
	}

	wg := new(sync.WaitGroup)
	wg.Add(dbN)

	// start benchmark test
	t1 := time.Now()

	for i := 0; i < dbN; i++ {
		go func(idx, num int) {
			dbo := dbs[idx]
			defer func() {
				_ = dbo.Close()
				wg.Done()
			}()
			for i := 0; i < num; i++ {
				sql := `INSERT INTO [logtest](Code,Type,Message,Account,CreateTime,CreateUser) VALUES(?,2,?,?,DATETIME(),?)`
				if _, err := dbo.Exec(sql, random.AlphaNumber(6), random.AlphaNumber(100), random.AlphaNumber(6), id.L36()); err != nil {
					t.Fatal(err)
				}
			}
		}(i, number/dbN)
	}

	wg.Wait()
	t2 := time.Now()
	ts := t2.Sub(t1)
	time.Sleep(time.Millisecond)

	t.Logf("Take time %s, handle sql records %d qps", ts, 1000*int64(number)/ts.Milliseconds())
}
