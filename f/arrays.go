package f

// Iterator is the function that accepts element of slice/array and its index
type Iterator func(interface{}, int)

// ResultIterator is the function that accepts element of slice/array and its index and returns any result
type ResultIterator func(interface{}, int) interface{}

// ConditionIterator is the function that accepts element of slice/array and its index and returns boolean
type ConditionIterator func(interface{}, int) bool

// Each iterates over the slice and apply Iterator to every item
func Each(array []interface{}, iterator Iterator) {
	for index, data := range array {
		iterator(data, index)
	}
}

// Maps iterates over the slice and apply ResultIterator to every item. Returns new slice as a result.
func Maps(array []interface{}, iterator ResultIterator) []interface{} {
	var result = make([]interface{}, len(array))
	for index, data := range array {
		result[index] = iterator(data, index)
	}
	return result
}

// Find iterates over the slice and apply ConditionIterator to every item. Returns first item that meet ConditionIterator or nil otherwise.
func Find(array []interface{}, iterator ConditionIterator) interface{} {
	for index, data := range array {
		if iterator(data, index) {
			return data
		}
	}
	return nil
}

// Filter iterates over the slice and apply ConditionIterator to every item. Returns new slice.
func Filter(array []interface{}, iterator ConditionIterator) []interface{} {
	var result = make([]interface{}, 0)
	for index, data := range array {
		if iterator(data, index) {
			result = append(result, data)
		}
	}
	return result
}

// Count iterates over the slice and apply ConditionIterator to every item. Returns count of items that meets ConditionIterator.
func Count(array []interface{}, iterator ConditionIterator) int {
	count := 0
	for index, data := range array {
		if iterator(data, index) {
			count = count + 1
		}
	}
	return count
}

// SplitObjects Separate objects into several size.
func SplitObjects(array []interface{}, size int) [][]interface{} {
	var chunkSet [][]interface{}
	var chunk []interface{}

	for len(array) > size {
		chunk, array = array[:size], array[size:]
		chunkSet = append(chunkSet, chunk)
	}
	if len(array) > 0 {
		chunkSet = append(chunkSet, array[:])
	}
	return chunkSet
}

// StringsContains Check if string value is contained in slice.
func StringsContains(s []string, sub string) bool {
	for _, v := range s {
		if v == sub {
			return true
		}
	}
	return false
}
