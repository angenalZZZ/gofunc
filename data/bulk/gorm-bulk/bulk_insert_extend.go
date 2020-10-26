package gormbulk

import (
	"github.com/angenalZZZ/gofunc/f"
	"github.com/dop251/goja"
	"github.com/jinzhu/gorm"

	"time"
)

func BulkInsertByJs(db *gorm.DB, objects []map[string]interface{}, chunkSize int, javascript string, interval time.Duration) error {
	var (
		vm  = goja.New()
		res goja.Value
		err error
	)

	// Split records with specified size not to exceed Database parameter limit
	for _, records := range f.SplitObjectMaps(objects, chunkSize) {
		// Input records
		vm.Set("records", records)

		// Output sql
		if res, err = vm.RunString(javascript); err != nil {
			return err
		}

		switch sql := res.Export().(type) {
		case string:
			if err := db.Exec(sql).Error; err != nil {
				return err
			}
		case []string:
			for _, s := range sql {
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
