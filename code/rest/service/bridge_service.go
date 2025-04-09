package service

import (
	"box/code/core"
	"box/code/rest/bean"
	"box/code/rest/dao"
	"box/code/rest/model"
)

// @Service
type BridgeService struct {
	bean.BaseBean
	bridgeDao *dao.BridgeDao
	userDao   *dao.UserDao
}

func (this *BridgeService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*dao.BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*dao.UserDao); ok {
		this.userDao = b
	}

}

func (this *BridgeService) Detail(uuid string) *model.Bridge {

	bridge := this.bridgeDao.CheckByUuid(uuid)

	return bridge
}
