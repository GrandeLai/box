package bean

import (
	"box/code/core"
	"box/code/rest/model"
	"box/code/tool/result"
	"box/code/tool/util"
	"net/http"
)

type BaseBean struct {
	Logger core.Logger
}

func (this *BaseBean) Init() {
	this.Logger = core.LOGGER
}

func (this *BaseBean) Bootstrap() {

}

// clean up the application.
func (this *BaseBean) Cleanup() {

}

// shortcut for panic check.
func (this *BaseBean) PanicError(err error) {
	core.PanicError(err)
}

// find the current user from request.
func (this *BaseBean) FindUser(request *http.Request) *model.User {

	//try to find from SessionCache.
	sessionId := util.GetSessionUuidFromRequest(request, core.COOKIE_AUTH_KEY)
	if sessionId == "" {
		return nil
	}

	cacheItem, err := core.CONTEXT.GetSessionCache().Value(sessionId)
	if err != nil {
		this.Logger.Warn("error while get from session cache. sessionId = %s, error = %v", sessionId, err)
		return nil
	}

	if cacheItem == nil || cacheItem.Data() == nil {

		this.Logger.Warn("cache item doesn't exist with sessionId = %s", sessionId)
		return nil
	}

	if value, ok := cacheItem.Data().(*model.User); ok {
		return value
	} else {
		this.Logger.Error("cache item not store the *User")
	}

	return nil

}

// find current error. If not found, panic the LOGIN error.
func (this *BaseBean) CheckUser(request *http.Request) *model.User {
	if this.FindUser(request) == nil {
		panic(result.LOGIN)
	} else {
		return this.FindUser(request)
	}
}
