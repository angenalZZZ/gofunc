package errorcode

import "testing"

func TestErrorCode_AddMsg(t *testing.T) {
	fun := "HTTP响应错误信息"
	msg := "成功获取到订单信息"
	err := OK.Msg(msg)
	if err.Code != int(OK) {
		t.Fatalf("%s OK.Msg > Code < %v", fun, err)
	}
	if err.GetHttpStatus() != OK.GetHttpStatus() {
		t.Fatalf("%s OK.Msg > GetHttpStatus < %v", fun, err)
	}

	err.AddMsg("总共%d条", 100)
	if err.Msg != msg+" 总共100条" {
		t.Fatalf("%s OK.Msg > AddMsg < %v", fun, err)
	}
}
