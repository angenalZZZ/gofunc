package sqlxbulk

import (
	"fmt"
	"time"

	"github.com/angenalZZZ/gofunc/data/cache/store"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/js"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"github.com/dop251/goja"
	"github.com/jmoiron/sqlx"
)

// BulkInsertByJs executes the query to insert multiple records at once.
func BulkInsertByJs(db *sqlx.DB, objects []map[string]interface{}, chunkSize int, javascript string, interval time.Duration, jsObj map[string]interface{}, varRecords ...string) error {
	var (
		fnName = "records"
		vm     = goja.New()
		res    goja.Value
		err    error
	)

	if len(varRecords) > 0 {
		fnName = varRecords[0]
	}

	js.Console(vm)
	js.Db(vm, db)
	js.Ajax(vm)
	if nat.Conn != nil && nat.Subject != "" {
		js.Nats(vm, nat.Conn, nat.Subject)
	}
	if store.RedisClient != nil {
		js.Redis(vm, store.RedisClient)
	}
	if jsObj != nil {
		for k, v := range jsObj {
			vm.Set(k, v)
		}
	}

	defer func() { vm.ClearInterrupt() }()

	// Split records with specified size not to exceed Database parameter limit
	for _, records := range f.SplitObjectMaps(objects, chunkSize) {
		// Input records
		vm.Set(fnName, records)

		// Output sql
		if res, err = vm.RunString(javascript); err != nil {
			return fmt.Errorf("the table script error, must contain array %q, error: %s", fnName, err.Error())
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
			if _, err := db.Exec(sql); err != nil {
				return err
			}
		case []string:
			for _, s := range sql {
				if len(s) < 20 {
					continue
				}
				if _, err := db.Exec(s); err != nil {
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
func BulkInsertByJsFunction(db *sqlx.DB, objects []map[string]interface{}, chunkSize int, javascript, functionName string, interval time.Duration, jsObj map[string]interface{}) error {
	var (
		vm = goja.New()
		fn func([]map[string]interface{}) interface{}
	)

	js.Console(vm)
	js.Db(vm, db)
	js.Ajax(vm)
	if nat.Conn != nil && nat.Subject != "" {
		js.Nats(vm, nat.Conn, nat.Subject)
	}
	if store.RedisClient != nil {
		js.Redis(vm, store.RedisClient)
	}
	if jsObj != nil {
		for k, v := range jsObj {
			vm.Set(k, v)
		}
	}

	defer func() { vm.ClearInterrupt() }()

	if _, err := vm.RunString(javascript); err != nil {
		return err
	}

	val := vm.Get(functionName)
	if val == nil {
		return fmt.Errorf("js function %q not found", functionName)
	}

	if err := vm.ExportTo(val, &fn); err != nil {
		return fmt.Errorf("js function %q not exported %s", functionName, err.Error())
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
			if _, err := db.Exec(sql); err != nil {
				return err
			}
		case []string:
			for _, s := range sql {
				if len(s) < 20 {
					continue
				}
				if _, err := db.Exec(s); err != nil {
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
