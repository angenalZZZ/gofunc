package main

import (
	"strings"
	"time"

	bulk "github.com/angenalZZZ/gofunc/data/bulk/gorm-bulk"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	json "github.com/json-iterator/go"
)

type handler struct{}

func (hub *handler) Handle(list [][]byte) error {
	size := len(list)
	if size == 0 {
		return nil
	}

	// gets records
	debug := configInfo.Log.Level == "debug" || nat.Log.GetLevel() <= 0
	records := make([]map[string]interface{}, 0, size)
	for _, item := range list {
		if len(item) == 0 {
			break
		}
		if item[0] == '{' {
			if debug {
				nat.Log.Debug().Msgf("[nats] received on %q: %s", subject, string(item))
			}

			var obj map[string]interface{}
			if err := json.Unmarshal(item, &obj); err != nil {
				continue
			}

			records = append(records, obj)
		}
	}

	count := len(records)
	if count == 0 {
		return nil
	}

	// save database
	db, err := gorm.Open(configInfo.Db.Type, configInfo.Db.Conn)
	if err != nil {
		return err
	}

	if debug {
		db = db.Debug()
	}

	sqlDB := db.DB()
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Minute)
	defer func() { _ = db.Close() }()

	script := configInfo.Db.Table.Script
	isFunc := strings.Contains(script, "function sql(")

	bulkSize := configInfo.Db.Table.Bulk
	bulkRecords, dataIndex := make([]map[string]interface{}, 0, bulkSize), 0
	for i := 0; i < count; i++ {
		obj := records[i]
		bulkRecords = append(bulkRecords, obj)
		if dataIndex++; dataIndex == bulkSize || dataIndex == count {
			// bulk handle
			if isFunc {
				if err = bulk.BulkInsertByJsFunction(db, bulkRecords, configInfo.Db.Table.Bulk, script, "sql", time.Microsecond); err != nil {
					return err
				}
			} else {
				if err = bulk.BulkInsertByJs(db, bulkRecords, configInfo.Db.Table.Bulk, script, time.Microsecond); err != nil {
					return err
				}
			}
			// reset data
			bulkRecords, dataIndex = make([]map[string]interface{}, 0, bulkSize), 0
		}
	}

	return nil
}
