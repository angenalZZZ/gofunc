package errorcode

import "net/http"

// 错误码表.
const (
	// 请求成功.
	OK = ToStatus(0)

	// 请求无效.
	INVALID = ToStatus(400)

	// 请求未认证通过.
	UNAUTHORIZED = ToStatus(401)

	// 无权限执行该操作.
	Forbidden = ToStatus(403)

	// 请求发生错误.
	ERROR = ToStatus(500)

	// 服务不可用.
	ServiceUnAvailable = ToStatus(503)
)

// init 错误码表.
func init() {
	SetStatus(OK, "请求成功.").SetHttpStatus(http.StatusOK)
	SetStatus(INVALID, "请求无效.").SetHttpStatus(http.StatusBadRequest)
	SetStatus(UNAUTHORIZED, "请求未认证通过.").SetHttpStatus(http.StatusUnauthorized)
	SetStatus(Forbidden, "无权限执行该操作.").SetHttpStatus(http.StatusForbidden)
	SetStatus(ERROR, "请求发生错误.").SetHttpStatus(http.StatusInternalServerError)
	SetStatus(ServiceUnAvailable, "服务不可用.").SetHttpStatus(http.StatusServiceUnavailable)
}
