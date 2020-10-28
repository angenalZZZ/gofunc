package main

import (
	"fmt"
	"strings"
	"time"

	bulk "github.com/angenalZZZ/gofunc/data/bulk/gorm-bulk"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"github.com/dop251/goja"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	json "github.com/json-iterator/go"
)

type handler struct{}

// Handle run default handler
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

	script, fn := configInfo.Db.Table.Script, "sql"
	isFn := strings.Contains(script, "function "+fn)

	bulkSize := configInfo.Db.Table.Bulk
	bulkRecords, dataIndex := make([]map[string]interface{}, 0, bulkSize), 0
	for i := 0; i < count; i++ {
		obj := records[i]
		bulkRecords = append(bulkRecords, obj)
		if dataIndex++; dataIndex == bulkSize || dataIndex == count {
			// bulk handle
			if isFn {
				if err = bulk.BulkInsertByJsFunction(db, bulkRecords, bulkSize, script, fn, time.Microsecond); err != nil {
					return err
				}
			} else {
				if err = bulk.BulkInsertByJs(db, bulkRecords, bulkSize, script, time.Microsecond); err != nil {
					return err
				}
			}
			// reset data
			bulkRecords, dataIndex = make([]map[string]interface{}, 0, bulkSize), 0
		}
	}

	return nil
}

// CheckJs run check javascript
func (hub *handler) CheckJs(script string) error {
	var (
		fnName  = "sql"
		isFn    = strings.Contains(script, "function "+fnName)
		objects []map[string]interface{}
		vm      = goja.New()
	)

	defer func() { vm.ClearInterrupt() }()

	if isFn {
		if _, err := vm.RunString(script); err != nil {
			return err
		}

		val := vm.Get(fnName)
		if val == nil {
			return fmt.Errorf("js function %q not found", fnName)
		}

		var fn func([]map[string]interface{}) interface{}
		if err := vm.ExportTo(val, &fn); err != nil {
			return fmt.Errorf("js function %q not exported %s", fnName, err.Error())
		}

		v := fn(objects)
		if v == nil {
			return nil
		}

		switch sql := v.(type) {
		case string:
			if sql != "" {
				return fmt.Errorf("js function %q return string must be empty", fnName)
			}
		case []string:
			if len(sql) > 0 {
				return fmt.Errorf("js function %q return string array must be empty", fnName)
			}
		default:
			return fmt.Errorf("js function %q return type must be string or string array", fnName)
		}
	} else {
		fnName = "records"
		vm.Set(fnName, objects)

		if res, err := vm.RunString(script); err != nil {
			return fmt.Errorf("the table script error, must contain array %q, error: %s", fnName, err.Error())
		} else if res == nil {
			return nil
		} else {
			v := res.Export()
			if v == nil {
				return nil
			}

			switch sql := v.(type) {
			case string:
				if sql != "" {
					return fmt.Errorf("js with array %q return string must be empty", fnName)
				}
			case []string:
				if len(sql) > 0 {
					return fmt.Errorf("js with array %q return string array must be empty", fnName)
				}
			default:
				return fmt.Errorf("js with array %q return type must be string or string array", fnName)
			}
		}
	}

	return nil
}
