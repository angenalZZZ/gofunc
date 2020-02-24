package errorcode

import "fmt"

// ToStatus Usage: Error Code.
type ToStatus int

// ToHttpStatus Usage: HTTP Status Code.
type ToHttpStatus int

// ErrorCode HTTP响应错误信息.
type ErrorCode struct {
	// Code 错误码
	Code int `json:"code"`
	// Msg 错误信息
	Msg string `json:"msg"`
	// Name 错误名
	Name string `json:"name"`
}

// Msg 生成HTTP响应错误信息.
func (i ToStatus) Msg(msg string) *ErrorCode {
	e := ErrorCode{Code: int(i), Msg: msg}
	for name, code := range errorCodeData {
		if e.Code == code {
			e.Name = name
			break
		}
	}
	return &e
}

// AddMsg 添加错误信息.
func (e *ErrorCode) AddMsg(format string, param ...interface{}) *ErrorCode {
	return e.AddMsgWith(" ", format, param...)
}

// AddMsgWith 添加错误信息,并设置拆分字符.
func (e *ErrorCode) AddMsgWith(split string, format string, param ...interface{}) *ErrorCode {
	if e.Msg == "" {
		e.Msg = fmt.Sprintf(format, param...)
	} else {
		e.Msg = e.Msg + split + fmt.Sprintf(format, param...)
	}
	return e
}

// GetStatus 获取状态码.
func (e *ErrorCode) GetStatus() int {
	return GetStatus(e.Name)
}

// GetStatus 获取状态码.
func GetStatus(name string) int {
	if status, ok := errorCodeData[name]; ok {
		return status
	}
	return -1
}

// SetStatus 设置状态码.
func SetStatus(code ToStatus, name string) ToHttpStatus {
	errorCodeData[name] = int(code)
	return ToHttpStatus(code)
}

// GetHttpStatus 获取HTTP状态码.
func (e *ErrorCode) GetHttpStatus() int {
	return ToStatus(e.Code).GetHttpStatus()
}

// GetHttpStatus 获取HTTP状态码.
func (i ToStatus) GetHttpStatus() int {
	if status, ok := errorCodeToHttpStatus[int(i)]; ok {
		return status
	}
	return -1
}

// SetHttpStatus Map To HTTP Status Code.
func (i ToHttpStatus) SetHttpStatus(status int) {
	errorCodeToHttpStatus[int(i)] = status
}

// 存储错误信息, 请不要修改.
var errorCodeData = map[string]int{}

// 存储错误信息, 请不要修改.
var errorCodeToHttpStatus = map[int]int{}
