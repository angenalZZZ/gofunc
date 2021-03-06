package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/js"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	json "github.com/json-iterator/go"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type handler struct {
	context.Context
	running bool
	// js code
	js     string
	jsr    *js.GoRuntime
	jsMod  time.Time
	jsFile string
	jsName string
	jsFunc func([]map[string]interface{}) interface{}
	// js global variable
	jso map[string]interface{}
	// js runtime and register param
	jsp *js.GoRuntimeParam
	// natS subscriber
	sub *nat.SubscriberFastCache
}

// Handle run default handler
func (h *handler) Handle(list [][]byte) error {
	h.running = true
	size := len(list)
	if size == 0 {
		h.running = false
		return nil
	}

	// gets records
	records := make([]map[string]interface{}, 0, size)
	debug := configInfo.Log.Level == "debug" || nat.Log.GetLevel() < 1
	for _, item := range list {
		if len(item) == 0 {
			break
		}
		if item[0] == '{' {
			if debug {
				nat.Log.Debug().Msgf("[nats] received on %q: %s", h.jsp.NatSubject, item)
			}

			var obj map[string]interface{}
			if err := json.Unmarshal(item, &obj); err != nil {
				continue
			}

			records = append(records, obj)
		}
	}

	if len(records) == 0 {
		h.running = false
		return nil
	}

	script, sqlLen := h.js, 20

	if h.jsFunc != nil {
		// js runtime and register
		//vm := js.NewRuntime(h.jsp)
		//defer vm.Clear()
		//
		//if _, err := vm.Runtime.RunString(script); err != nil {
		//	h.running = false
		//	return err
		//}
		//var fn func([]map[string]interface{}) interface{}
		//val := vm.Runtime.Get(h.jsName)
		//if val == nil {
		//	h.running = false
		//	return fmt.Errorf("js function %q not found", fnName)
		//}
		//if err := vm.Runtime.ExportTo(val, &fn); err != nil {
		//	h.running = false
		//	return fmt.Errorf("js function %q not exported %s", fnName, err.Error())
		//}

		db := h.jsp.DB
		if db == nil && h.jsr != nil {
			db = h.jsr.DB
		}
		if db == nil {
			var err error
			db, err = sqlx.Connect(h.jsp.DbType, h.jsp.DbConn)
			if err != nil {
				return err
			}
			if db == nil {
				return fmt.Errorf("%s Connect Error", h.jsp.DbType)
			}
			defer func() { _ = db.Close() }()
		}

		// Split records with specified size not to exceed Database parameter limit
		for _, rs := range f.SplitObjectMaps(records, configInfo.Nats.Bulk) {
			// Output sql
			val := h.jsFunc(rs)
			if val == nil {
				continue
			}

			switch sql := val.(type) {
			case string:
				if len(sql) < sqlLen {
					continue
				}
				if _, err := db.Exec(sql); err != nil {
					h.running = false
					return err
				}
			case []string:
				for _, s := range sql {
					if len(s) < sqlLen {
						continue
					}
					if _, err := db.Exec(s); err != nil {
						h.running = false
						return err
					}
				}
			}

			time.Sleep(time.Microsecond)
		}
	} else {
		// js runtime and register
		vm, fnName := h.jsr, h.jsName
		if vm == nil {
			vm = js.NewRuntime(h.jsp)
		}
		defer vm.Clear()

		db := h.jsp.DB
		if db == nil {
			db = vm.DB
		}
		if db == nil {
			var err error
			db, err = sqlx.Connect(h.jsp.DbType, h.jsp.DbConn)
			if err != nil {
				return err
			}
			if db == nil {
				return fmt.Errorf("%s Connect Error", h.jsp.DbType)
			}
			defer func() { _ = db.Close() }()
		}

		// Split records with specified size not to exceed Database parameter limit
		for _, rs := range f.SplitObjectMaps(records, configInfo.Nats.Bulk) {
			// Input records
			vm.Runtime.Set(h.jsName, rs)

			// Output sql
			res, err := vm.Runtime.RunString(script)
			if err != nil {
				h.running = false
				return fmt.Errorf("the table script error, must contain array %q, error: %s", fnName, err.Error())
			}
			if res == nil {
				continue
			}

			val := res.Export()
			if val == nil {
				continue
			}

			switch sql := val.(type) {
			case string:
				if len(sql) < sqlLen {
					continue
				}
				if _, err := db.Exec(sql); err != nil {
					h.running = false
					return err
				}
			case []string:
				for _, s := range sql {
					if len(s) < sqlLen {
						continue
					}
					if _, err := db.Exec(s); err != nil {
						h.running = false
						return err
					}
				}
			}

			time.Sleep(time.Microsecond)
		}
	}

	h.running = false
	return nil
}

// Stop run
func (h *handler) Stop(ms ...int) {
	n := 10000
	if len(ms) > 0 {
		n = ms[0]
	}
	for ; h.running && n > 0; n-- {
		time.Sleep(time.Millisecond)
	}
}

func (h *handler) isScriptMod() bool {
	if h.jsFile == "" {
		return false
	}
	info, err := os.Stat(h.jsFile)
	if os.IsNotExist(err) {
		return false
	}
	if t := info.ModTime(); t.Unix() != h.jsMod.Unix() {
		h.jsMod = t
		return true
	}
	return false
}

func (h *handler) doScriptMod() error {
	if h.jsFile == "" {
		return nil
	}
	script, err := f.ReadFile(h.jsFile)
	if err != nil {
		return err
	}

	h.js = strings.TrimSpace(string(script))
	return nil
}
