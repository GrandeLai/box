package result

import (
	"box/code/tool/i18n"
	"fmt"
	"net/http"
	"strconv"
)

type WebResult struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (this *WebResult) Error() string {
	return this.Msg
}

type CodeWrapper struct {
	Code        string
	HttpStatus  int
	Description string
}

var (
	OK                     = &CodeWrapper{Code: "OK", HttpStatus: http.StatusOK, Description: "ok"}
	BAD_REQUEST            = &CodeWrapper{Code: "BAD_REQUEST", HttpStatus: http.StatusBadRequest, Description: "bad request"}
	NEED_SHARE_CODE        = &CodeWrapper{Code: "NEED_SHARE_CODE", HttpStatus: http.StatusUnauthorized, Description: "share code required"}
	SHARE_CODE_ERROR       = &CodeWrapper{Code: "SHARE_CODE_ERROR", HttpStatus: http.StatusUnauthorized, Description: "share code error"}
	LOGIN                  = &CodeWrapper{Code: "LOGIN", HttpStatus: http.StatusUnauthorized, Description: "not login"}
	USER_DISABLED          = &CodeWrapper{Code: "USER_DISABLED", HttpStatus: http.StatusForbidden, Description: "user disabled"}
	UNAUTHORIZED           = &CodeWrapper{Code: "UNAUTHORIZED", HttpStatus: http.StatusUnauthorized, Description: "unauthorized"}
	NOT_FOUND              = &CodeWrapper{Code: "NOT_FOUND", HttpStatus: http.StatusNotFound, Description: "404 not found"}
	METHOD_NOT_ALLOWED     = &CodeWrapper{Code: "METHOD_NOT_ALLOWED", HttpStatus: http.StatusMethodNotAllowed, Description: "405 method not allowed"}
	CONFLICT               = &CodeWrapper{Code: "CONFLICT", HttpStatus: http.StatusConflict, Description: "409 conflict"}
	PRECONDITION_FAILED    = &CodeWrapper{Code: "PRECONDITION_FAILED", HttpStatus: http.StatusPreconditionFailed, Description: "412 precondition failed"}
	UNSUPPORTED_MEDIA_TYPE = &CodeWrapper{Code: "UNSUPPORTED_MEDIA_TYPE", HttpStatus: http.StatusUnsupportedMediaType, Description: "415 conflict"}
	RANGE_NOT_SATISFIABLE  = &CodeWrapper{Code: "RANGE_NOT_SATISFIABLE", HttpStatus: http.StatusRequestedRangeNotSatisfiable, Description: "range not satisfiable"}
	NOT_INSTALLED          = &CodeWrapper{Code: "NOT_INSTALLED", HttpStatus: http.StatusInternalServerError, Description: "application not installed"}
	SERVER                 = &CodeWrapper{Code: "SERVER", HttpStatus: http.StatusInternalServerError, Description: "server error"}
	UNKNOWN                = &CodeWrapper{Code: "UNKNOWN", HttpStatus: http.StatusInternalServerError, Description: "server unknow error"}
)

func FetchHttpStatus(code string) int {
	if code == OK.Code {
		return OK.HttpStatus
	} else if code == BAD_REQUEST.Code {
		return BAD_REQUEST.HttpStatus
	} else if code == NEED_SHARE_CODE.Code {
		return NEED_SHARE_CODE.HttpStatus
	} else if code == SHARE_CODE_ERROR.Code {
		return SHARE_CODE_ERROR.HttpStatus
	} else if code == LOGIN.Code {
		return LOGIN.HttpStatus
	} else if code == USER_DISABLED.Code {
		return USER_DISABLED.HttpStatus
	} else if code == UNAUTHORIZED.Code {
		return UNAUTHORIZED.HttpStatus
	} else if code == NOT_FOUND.Code {
		return NOT_FOUND.HttpStatus
	} else if code == METHOD_NOT_ALLOWED.Code {
		return METHOD_NOT_ALLOWED.HttpStatus
	} else if code == CONFLICT.Code {
		return CONFLICT.HttpStatus
	} else if code == PRECONDITION_FAILED.Code {
		return PRECONDITION_FAILED.HttpStatus
	} else if code == UNSUPPORTED_MEDIA_TYPE.Code {
		return UNSUPPORTED_MEDIA_TYPE.HttpStatus
	} else if code == RANGE_NOT_SATISFIABLE.Code {
		return RANGE_NOT_SATISFIABLE.HttpStatus
	} else if code == NOT_INSTALLED.Code {
		return NOT_INSTALLED.HttpStatus
	} else if code == SERVER.Code {
		return SERVER.HttpStatus
	} else {
		//if this is an int. regard it as statusCode
		statusCode, err := strconv.Atoi(code)
		if err != nil {
			return UNKNOWN.HttpStatus
		} else {
			return statusCode
		}
	}
}

func ConstWebResult(codeWrapper *CodeWrapper) *WebResult {

	wr := &WebResult{
		Code: codeWrapper.Code,
		Msg:  codeWrapper.Description,
	}
	return wr
}

func CustomWebResultI18n(request *http.Request, codeWrapper *CodeWrapper, item *i18n.Item, v ...interface{}) *WebResult {

	return CustomWebResult(codeWrapper, fmt.Sprintf(item.Message(request), v...))

}

func CustomWebResult(codeWrapper *CodeWrapper, description string) *WebResult {

	if description == "" {
		description = codeWrapper.Description
	}
	wr := &WebResult{
		Code: codeWrapper.Code,
		Msg:  description,
	}
	return wr
}

// use standard http status code.
func StatusCodeWebResult(statusCode int, description string) *WebResult {
	if description == "" {
		description = http.StatusText(statusCode)
	}
	wr := &WebResult{
		Code: fmt.Sprintf("%d", statusCode),
		Msg:  description,
	}
	return wr
}

func BadRequestI18n(request *http.Request, item *i18n.Item, v ...interface{}) *WebResult {
	return CustomWebResult(BAD_REQUEST, fmt.Sprintf(item.Message(request), v...))
}

func BadRequest(format string, v ...interface{}) *WebResult {
	return CustomWebResult(BAD_REQUEST, fmt.Sprintf(format, v...))
}

func Unauthorized(format string, v ...interface{}) *WebResult {
	return CustomWebResult(UNAUTHORIZED, fmt.Sprintf(format, v...))
}

func NotFound(format string, v ...interface{}) *WebResult {
	return CustomWebResult(NOT_FOUND, fmt.Sprintf(format, v...))

}

// sever inner error
func Server(format string, v ...interface{}) *WebResult {
	return CustomWebResult(SERVER, fmt.Sprintf(format, v...))
}

// db error.
var (
	DB_ERROR_DUPLICATE_KEY  = "Error 1062: Duplicate entry"
	DB_ERROR_NOT_FOUND      = "record not found"
	DB_TOO_MANY_CONNECTIONS = "Error 1040: Too many connections"
	DB_BAD_CONNECTION       = "driver: bad connection"
)
