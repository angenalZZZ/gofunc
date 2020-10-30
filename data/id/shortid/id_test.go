package shortid

import "testing"

func TestGetDefault(t *testing.T) {
	id := GetDefault()
	if s, err := id.Generate(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(s)
	}
}

func TestSetDefault(t *testing.T) {
	id := MustNew(1, DefaultABC, 2048)
	SetDefault(id)
	s := MustGenerate()
	t.Log(s)
}

func TestMustNewAbc(t *testing.T) {
	abc := MustNewAbc(DefaultABC, 1)
	if id, err := abc.Encode(214235345234524356, 0, 6); err != nil {
		t.Error(err)
	} else if len(id) != 10 {
		t.Errorf("expected 10 symbols")
	}
}
