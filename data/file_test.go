package data

import (
	"testing"
)

var testBufJSON = `{"Code":"用户登录","Type":2,"Message":"【管理员】登录","Exception":null,"ActionName":"Account.LoginWithCode","Data":"{\"Name\":\"admin\",\"Pwd\":\"96e79218965eb72c92a549dd5a330112\"}","CreateTime":"2020-10-01 16:49:32"}`

func TestObjectJSON(t *testing.T) {
	buf := testBufJSON
	if obj, err := ObjectJSON([]byte(buf)); err != nil {
		t.Fatal(err)
	} else {
		t.Log(obj)
	}
}

func TestListJSON(t *testing.T) {
	buf := "[" + testBufJSON + "]"
	if list, err := ListJSON([]byte(buf)); err != nil {
		t.Fatal(err)
	} else {
		t.Log(list)
	}
}

func TestListData(t *testing.T) {
	buf := "[" + testBufJSON + "]"
	if list, err := ListData([]byte(buf)); err != nil {
		t.Fatal(err)
	} else {
		for index, item := range list {
			t.Logf("%d: %s", index, item)
		}
	}
}
