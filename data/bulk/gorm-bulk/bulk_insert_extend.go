package gormbulk

import (
	"errors"
	"fmt"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

func BulkInsertByTbl(db *gorm.DB, objects []map[string]interface{}, chunkSize int, interval time.Duration, tableName string, excludeColumns ...string) error {
	// Split records with specified size not to exceed Database parameter limit
	for _, objSet := range f.SplitObjectMaps(objects, chunkSize) {
		if err := InsertObjsByTbl(db, objSet, tableName, excludeColumns...); err != nil {
			return err
		}
		if interval > 0 {
			time.Sleep(interval)
		}
	}
	return nil
}

func InsertObjsByTbl(db *gorm.DB, objects []map[string]interface{}, tableName string, excludeColumns ...string) error {
	if len(objects) == 0 {
		return nil
	}

	db = db.Table(tableName)
	firstAttrs := objects[0]
	attrSize := len(firstAttrs)

	// Scope to eventually run SQL
	mainScope := db.NewScope(objects[0])
	// Store placeholders for embedding variables
	placeholders := make([]string, 0, attrSize)

	// Replace with database column name
	dbColumns := make([]string, 0, attrSize)
	for _, key := range f.MapKeySorted(firstAttrs) {
		dbColumns = append(dbColumns, mainScope.Quote(key))
	}

	for _, obj := range objects {
		objAttrs := obj

		// If object sizes are different, SQL statement loses consistency
		if len(objAttrs) != attrSize {
			return errors.New("attribute sizes are inconsistent")
		}

		scope := db.NewScope(obj)

		// Append variables
		variables := make([]string, 0, attrSize)
		for _, key := range f.MapKeySorted(objAttrs) {
			scope.AddToVars(objAttrs[key])
			variables = append(variables, "?")
		}

		valueQuery := "(" + strings.Join(variables, ", ") + ")"
		placeholders = append(placeholders, valueQuery)

		// Also append variables to mainScope
		mainScope.SQLVars = append(mainScope.SQLVars, scope.SQLVars...)
	}

	insertOption := ""
	if val, ok := db.Get("gorm:insert_option"); ok {
		strVal, ok := val.(string)
		if !ok {
			return errors.New("gorm:insert_option should be a string")
		}
		insertOption = strVal
	}

	mainScope.Raw(fmt.Sprintf("INSERT INTO %s (%s) VALUES %s %s",
		mainScope.QuotedTableName(),
		strings.Join(dbColumns, ", "),
		strings.Join(placeholders, ", "),
		insertOption,
	))

	return db.Exec(mainScope.SQL, mainScope.SQLVars...).Error
}
