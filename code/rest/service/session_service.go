package service

import (
	"box/code/core"
	"box/code/rest/bean"
	"box/code/rest/dao"
)

// @Service
type SessionService struct {
	bean.BaseBean
	userDao    *dao.UserDao
	sessionDao *dao.SessionDao
}

func (this *SessionService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*dao.UserDao); ok {
		this.userDao = b
	}

	b = core.CONTEXT.GetBean(this.sessionDao)
	if b, ok := b.(*dao.SessionDao); ok {
		this.sessionDao = b
	}

}

// System cleanup.
func (this *SessionService) Cleanup() {

	this.Logger.Info("[SessionService] clean up. Delete all Session. total:%d", core.CONTEXT.GetSessionCache().Count())

	core.CONTEXT.GetSessionCache().Truncate()
}
