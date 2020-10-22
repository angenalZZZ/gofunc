package f_test

import (
	"github.com/angenalZZZ/gofunc/f"
	"testing"
)

func TestEach(t *testing.T) {
	t.Parallel()
	acc := 0
	data := []interface{}{1, 2, 3, 4, 5}
	var fn f.Iterator = func(value interface{}, index int) {
		acc = acc + value.(int)
	}
	f.Each(data, fn)
	if acc != 15 {
		t.Errorf("Expected Each(..) to be %v, got %v", 15, acc)
	}
}

func ExampleEach() {
	data := []interface{}{1, 2, 3, 4, 5}
	var fn f.Iterator = func(value interface{}, index int) {
		println(value.(int))
	}
	f.Each(data, fn)
}

func TestMaps(t *testing.T) {
	t.Parallel()
	data := []interface{}{1, 2, 3, 4, 5}
	var fn f.ResultIterator = func(value interface{}, index int) interface{} {
		return value.(int) * 3
	}
	result := f.Maps(data, fn)
	for i, d := range result {
		if d != fn(data[i], i) {
			t.Errorf("Expected Map(..) to be %v, got %v", fn(data[i], i), d)
		}
	}
}

func ExampleMaps() {
	data := []interface{}{1, 2, 3, 4, 5}
	var fn f.ResultIterator = func(value interface{}, index int) interface{} {
		return value.(int) * 3
	}
	_ = f.Maps(data, fn) // result = []interface{}{1, 6, 9, 12, 15}
}

func TestFind(t *testing.T) {
	t.Parallel()
	findElement := 96
	data := []interface{}{1, 2, 3, 4, findElement, 5}
	var fn1 f.ConditionIterator = func(value interface{}, index int) bool {
		return value.(int) == findElement
	}
	var fn2 f.ConditionIterator = func(value interface{}, index int) bool {
		value, _ = value.(string)
		return value == "govalidator"
	}
	val1 := f.Find(data, fn1)
	val2 := f.Find(data, fn2)
	if val1 != findElement {
		t.Errorf("Expected Find(..) to be %v, got %v", findElement, val1)
	}
	if val2 != nil {
		t.Errorf("Expected Find(..) to be %v, got %v", nil, val2)
	}
}

func TestFilter(t *testing.T) {
	t.Parallel()
	data := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	answer := []interface{}{2, 4, 6, 8, 10}
	var fn f.ConditionIterator = func(value interface{}, index int) bool {
		return value.(int)%2 == 0
	}
	result := f.Filter(data, fn)
	for i := range result {
		if result[i] != answer[i] {
			t.Errorf("Expected Filter(..) to be %v, got %v", answer[i], result[i])
		}
	}
}

func ExampleFilter() {
	data := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var fn f.ConditionIterator = func(value interface{}, index int) bool {
		return value.(int)%2 == 0
	}
	_ = f.Filter(data, fn) // result = []interface{}{2, 4, 6, 8, 10}
}

func TestCount(t *testing.T) {
	t.Parallel()
	data := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	count := 5
	var fn f.ConditionIterator = func(value interface{}, index int) bool {
		return value.(int)%2 == 0
	}
	result := f.Count(data, fn)
	if result != count {
		t.Errorf("Expected Count(..) to be %v, got %v", count, result)
	}
}

func ExampleCount() {
	data := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var fn f.ConditionIterator = func(value interface{}, index int) bool {
		return value.(int)%2 == 0
	}
	_ = f.Count(data, fn) // result = 5
}
