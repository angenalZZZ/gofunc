package gormbulk

import (
	"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	json "github.com/json-iterator/go"
	"testing"
)

func TestBulkInsertByTbl(t *testing.T) {
	item := `{"Code":"Login","Type":2,"Message":"Admin Login","Account":"admin","CreateTime":"2020-10-25T11:29:43.5757388+08:00"}`
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(item), &obj); err != nil {
		t.Fatal(err)
	}

	t.Logf("%v", obj)

	// save database
	db, err := gorm.Open("mysql", "root:HGJ766GR767FKJU0@tcp(localhost:3306)/test?charset=utf8")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = db.Close() }()
	err = db.Exec("instert into logtest()").Create(obj).Error
	if err != nil {
		t.Fatal(err)
	}
}
