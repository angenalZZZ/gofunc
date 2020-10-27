package gormbulk

import (
	"fmt"

	"github.com/angenalZZZ/gofunc/f"
	"github.com/dop251/goja"
	"github.com/jinzhu/gorm"

	"time"
)

// BulkInsertByJs executes the query to insert multiple records at once.
func BulkInsertByJs(db *gorm.DB, objects []map[string]interface{}, chunkSize int, javascript string, interval time.Duration) error {
	var (
		vm  = goja.New()
		res goja.Value
		err error
	)

	defer func() { vm.ClearInterrupt() }()

	// Split records with specified size not to exceed Database parameter limit
	for _, records := range f.SplitObjectMaps(objects, chunkSize) {
		// Input records
		vm.Set("records", records)

		// Output sql
		if res, err = vm.RunString(javascript); err != nil {
			return err
		} else if res == nil {
			continue
		}

		val := res.Export()
		if val == nil {
			continue
		}

		switch sql := val.(type) {
		case string:
			if len(sql) < 20 {
				continue
			}
			if err := db.Exec(sql).Error; err != nil {
				return err
			}
		case []string:
			for _, s := range sql {
				if len(s) < 20 {
					continue
				}
				if err := db.Exec(s).Error; err != nil {
					return err
				}
			}
		}

		if interval > 0 {
			time.Sleep(interval)
		}
	}
	return nil
}

// BulkInsertByJsFunction executes the query to insert multiple records at once.
func BulkInsertByJsFunction(db *gorm.DB, objects []map[string]interface{}, chunkSize int, javascript, functionName string, interval time.Duration) error {
	var (
		vm = goja.New()
		fn func([]map[string]interface{}) interface{}
	)

	defer func() { vm.ClearInterrupt() }()

	if _, err := vm.RunString(javascript); err != nil {
		return err
	}
	val := vm.Get(functionName)
	if val == nil {
		return fmt.Errorf("js function %s not found", functionName)
	}
	if err := vm.ExportTo(val, &fn); err != nil {
		return fmt.Errorf("js function %s not exported %s", functionName, err.Error())
	}

	// Split records with specified size not to exceed Database parameter limit
	for _, records := range f.SplitObjectMaps(objects, chunkSize) {
		// Output sql
		val := fn(records)
		if val == nil {
			continue
		}

		switch sql := val.(type) {
		case string:
			if len(sql) < 20 {
				continue
			}
			if err := db.Exec(sql).Error; err != nil {
				return err
			}
		case []string:
			for _, s := range sql {
				if len(s) < 20 {
					continue
				}
				if err := db.Exec(s).Error; err != nil {
					return err
				}
			}
		}

		if interval > 0 {
			time.Sleep(interval)
		}
	}
	return nil
}
