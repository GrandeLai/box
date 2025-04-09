package service

import (
	"box/code/core"
	"box/code/rest/bean"
	"box/code/rest/dao"
	"box/code/rest/model"
)

// @Service
type PreferenceService struct {
	bean.BaseBean
	preferenceDao *dao.PreferenceDao
	preference    *model.Preference
	matterDao     *dao.MatterDao
	matterService *MatterService
	userDao       *dao.UserDao
	migrating     bool
}

func (this *PreferenceService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.preferenceDao)
	if b, ok := b.(*dao.PreferenceDao); ok {
		this.preferenceDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*dao.MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*dao.UserDao); ok {
		this.userDao = b
	}

}

func (this *PreferenceService) Fetch() *model.Preference {

	if this.preference == nil {
		this.preference = this.preferenceDao.Fetch()
	}

	return this.preference
}

// 清空单例配置。
func (this *PreferenceService) Reset() {

	this.preference = nil

}

// 清空单例配置。
func (this *PreferenceService) Save(preference *model.Preference) *model.Preference {

	preference = this.preferenceDao.Save(preference)

	//clean cache.
	this.Reset()

	return preference
}

// System cleanup.
func (this *PreferenceService) Cleanup() {

	this.Logger.Info("[PreferenceService] clean up. Delete all preference")

	this.Reset()
}
