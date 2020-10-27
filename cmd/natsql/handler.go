package main

import (
	"time"

	bulk "github.com/angenalZZZ/gofunc/data/bulk/gorm-bulk"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"github.com/jinzhu/gorm"
	json "github.com/json-iterator/go"
)

type handler struct{}

func (hub *handler) Handle(list [][]byte) error {
	size := len(list)
	if size == 0 {
		return nil
	}

	// gets records
	records := make([]map[string]interface{}, 0, size)
	for _, item := range list {
		if len(item) == 0 {
			break
		}
		if item[0] == '{' {
			nat.Log.Debug().Msgf("[nats] received on %q: %s", subject, string(item))

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

	sqlDB := db.DB()
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Minute)
	defer func() { _ = db.Close() }()

	bulkSize := configInfo.Db.Table.Bulk
	bulkRecords, dataIndex := make([]map[string]interface{}, 0, bulkSize), 0
	for i := 0; i < count; i++ {
		obj := records[i]
		bulkRecords = append(bulkRecords, obj)
		if dataIndex++; dataIndex == bulkSize || dataIndex == count {
			// bulk handle
			if err = bulk.BulkInsertByJs(db, bulkRecords, configInfo.Db.Table.Bulk, configInfo.Db.Table.Script, time.Microsecond); err != nil {
				return err
			}
			// reset data
			bulkRecords, dataIndex = make([]map[string]interface{}, 0, bulkSize), 0
		}
	}

	return nil
}
