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

	count, size, keys := int64(0), db.Size(), db.Keys()
	t.Logf("db.Size = %d, db.Keys.Count = %d\n", size, len(keys))

	count, err = db.Incr(testCountIncrKey, 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("db.Incr = %d\n", count)

	someVal := "hello"
	err = db.Set(testSomeKey, someVal, 10)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Logf("db.Set = %s\n", someVal)
	}
	someVal, err = db.Get(testSomeKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("db.Get = %s\n", someVal)

	size, keys = db.Size(), db.Keys()
	t.Logf("db.Size = %d, db.Keys.Count = %d\n", size, len(keys))

	err = db.Del([]string{testCountIncrKey, testSomeKey})
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Get(testSomeKey)
	if err == nil {
		t.Fatal(fmt.Errorf("db.Del = %t\n", false))
	} else {
		t.Logf("db.Del = %t\n", true)
	}

	size, keys = db.Size(), db.Keys()
	t.Logf("db.Size = %d, db.Keys.Count = %d\n", size, len(keys))

	//_ = db.GC()
}
