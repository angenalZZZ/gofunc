package gormbulk

import (
	"github.com/angenalZZZ/gofunc/f"
	"github.com/jinzhu/gorm"
	"time"

	//_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	json "github.com/json-iterator/go"
	"testing"
)

func TestBulkInsertByJs(t *testing.T) {
	var record map[string]interface{}
	item := `{"Code":"Login","Type":2,"Message":"Admin Login","Account":"admin","CreateTime":"2020-10-25T11:29:43.5757388+08:00"}`
	if err := json.Unmarshal([]byte(item), &record); err != nil {
		t.Fatal(err)
	}

	records := make([]map[string]interface{}, 0, 10)
	f.Times(cap(records), func(i int) {
		records = append(records, record)
	})

	script := `"insert into logtest(Code,Type,Message,Account,CreateTime) values" 
		+ records.map(function(item){
		return "(" 
			+ "'" + item.Code + "'," 
			+ item.Type + "," 
			+ "'" + item.Message + "'," 
			+ "'" + item.Account + "'," 
			+ "'" + item.CreateTime.replace('T',' ').split('.')[0] + "'"
			+ ")";
	}).join(",") + ";"`

	// save database
	db, err := gorm.Open("mysql", "root:HGJ766GR767FKJU0@tcp(localhost:3306)/test?charset=utf8")
	if err != nil {
		t.Fatal(err)
	}

	db = db.Debug()
	defer func() { _ = db.Close() }()
	if err = BulkInsertByJs(db, records, 5, script, time.Second); err != nil {
		t.Fatal(err)
	}
}
