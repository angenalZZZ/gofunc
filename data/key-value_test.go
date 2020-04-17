package data

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/data/kv"
	"os"
	"testing"
)

const (
	testDBPath       = "../test/data"
	testCountIncrKey = "count"
	testSomeKey      = "some"
)

func TestBadgerDB(t *testing.T) {
	var db KvDB = new(kv.BadgerDB)
	err := db.Open(testDBPath)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = db.Close()
		_ = os.RemoveAll(testDBPath)
	}()

	var count int64
	getSizeAndKeys(t, db)

	count, err = db.Incr(testCountIncrKey, 1)
	if err != nil {
		t.Error(err)
	}
	t.Logf("db.Incr = %d\n", count)

	someVal := "hello"
	genVal(t, db, testSomeKey, someVal)
	getVal(t, db, testSomeKey, someVal)

	getSizeAndKeys(t, db)

	err = db.Del([]string{testCountIncrKey, testSomeKey})
	if err != nil {
		t.Error(err)
	}

	_, err = db.Get(testSomeKey)
	if err == nil {
		t.Error(fmt.Errorf("db.Del = %t\n", false))
	} else {
		t.Logf("db.Del = %t\n", true)
	}

	getSizeAndKeys(t, db)
	//_ = db.GC()
}

func getSizeAndKeys(t *testing.T, db KvDB) {
	t.Helper()
	size, keys := db.Size(), db.Keys()
	t.Logf("db.Size = %d, db.Keys.Count = %d\n", size, len(keys))
}

func getVal(t *testing.T, db KvDB, key, expected string) {
	t.Helper()
	if get, err := db.Get(key); err != nil {
		t.Error(err)
	} else if get != expected {
		t.Errorf("Expected value (%v) was not returned from db, instead got %v", expected, get)
	}
}

func genVal(t *testing.T, db KvDB, key, expected string) {
	t.Helper()
	if err := db.Set(key, expected, 10); err != nil {
		t.Error(err)
	}
}
