package errorcode

import "net/http"

// 错误码表.
const (
	// 缺少参数.
	MissingParameter = ToStatus(100401)
	// 参数无效.
	InvalidParameter = ToStatus(100402)

	// 账号未开通相应服务.
	OperationDenied = ToStatus(100403)
	// 账号已欠费，请充值.
	OperationDeniedSuspended = ToStatus(100404)

	// 后台发生未知错误，请稍后重试或联系客服解决.
	InternalError = ToStatus(100500)
)

// init 错误码表.
func init() {
	SetStatus(MissingParameter, "缺少参数.").SetHttpStatus(http.StatusBadRequest)
	SetStatus(InvalidParameter, "参数无效.").SetHttpStatus(http.StatusBadRequest)
	SetStatus(OperationDenied, "账号未开通相应服务.").SetHttpStatus(http.StatusForbidden)
	SetStatus(OperationDeniedSuspended, "账号已欠费，请充值.").SetHttpStatus(http.StatusForbidden)
	SetStatus(InternalError, "后台发生未知错误，请稍后重试或联系客服解决.").SetHttpStatus(http.StatusInternalServerError)
}
