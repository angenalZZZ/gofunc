package f

import "testing"

func TestCMap_JSON(t *testing.T) {
	m := NewConcurrentMap()
	m.MSet(map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": 3,
	})
	s, _ := m.JSON()
	t.Logf("%s\n", s)

	m2, err2 := NewConcurrentMapFromJSON(s)
	if err2 != nil {
		t.Fatal(err2)
	}
	s2, _ := m2.JSON()
	t.Logf("%s\n", s2)
}
