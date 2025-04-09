package service

import (
	"box/code/core"
	"box/code/rest/bean"
	"box/code/rest/dao"
	"box/code/rest/model"
)

// @Service
type SpaceMemberService struct {
	bean.BaseBean
	spaceMemberDao *dao.SpaceMemberDao
	matterDao      *dao.MatterDao
	bridgeDao      *dao.BridgeDao
	userDao        *dao.UserDao
}

func (this *SpaceMemberService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.spaceMemberDao)
	if b, ok := b.(*dao.SpaceMemberDao); ok {
		this.spaceMemberDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*dao.MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*dao.BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*dao.UserDao); ok {
		this.userDao = b
	}

}

func (this *SpaceMemberService) Detail(uuid string) *model.SpaceMember {

	spaceMember := this.spaceMemberDao.CheckByUuid(uuid)

	return spaceMember
}

// create space
func (this *SpaceMemberService) CreateMember(space *model.Space, user *model.User, spaceRole string) *model.SpaceMember {

	spaceMember := &model.SpaceMember{
		SpaceUuid: space.Uuid,
		UserUuid:  user.Uuid,
		Role:      spaceRole,
	}

	spaceMember = this.spaceMemberDao.Create(spaceMember)

	return spaceMember

}

// 当前用户对于此空间，是否有管理权限。
func (this *SpaceMemberService) CanManage(user *model.User, spaceUuid string) bool {
	if user.Role == model.USER_ROLE_ADMINISTRATOR {
		return true
	}
	if user.SpaceUuid == spaceUuid {
		return true
	}

	//only space's admin can add member.
	spaceMember := this.spaceMemberDao.FindBySpaceUuidAndUserUuid(spaceUuid, user.Uuid)
	return this.canManageBySpaceMember(user, spaceMember)
}

// 当前用户对于此空间，是否有可读权限。
func (this *SpaceMemberService) CanRead(user *model.User, spaceUuid string) bool {
	if user.Role == model.USER_ROLE_ADMINISTRATOR {
		return true
	}
	if user.SpaceUuid == spaceUuid {
		return true
	}

	//only space's admin can add member.
	spaceMember := this.spaceMemberDao.FindBySpaceUuidAndUserUuid(spaceUuid, user.Uuid)
	return this.canReadBySpaceMember(user, spaceMember)
}

// 当前用户对于此空间，是否有可写权限。
func (this *SpaceMemberService) canWrite(user *model.User, spaceUuid string) bool {
	if user.Role == model.USER_ROLE_ADMINISTRATOR {
		return true
	}
	if user.SpaceUuid == spaceUuid {
		return true
	}

	//only space's admin can add member.
	spaceMember := this.spaceMemberDao.FindBySpaceUuidAndUserUuid(spaceUuid, user.Uuid)
	return this.canWriteBySpaceMember(user, spaceMember)
}

// 当前用户对于此空间，是否有管理权限。
func (this *SpaceMemberService) canManageBySpaceMember(user *model.User, member *model.SpaceMember) bool {
	if user.Role == model.USER_ROLE_ADMINISTRATOR {
		return true
	}

	//only space's admin can add member.
	if member != nil && member.Role == model.SPACE_MEMBER_ROLE_ADMIN {
		return true
	}

	return false
}

// 当前用户对于此空间，是否有可读权限。
func (this *SpaceMemberService) canReadBySpaceMember(user *model.User, member *model.SpaceMember) bool {
	if user.Role == model.USER_ROLE_ADMINISTRATOR {
		return true
	}

	//only space's admin can add member.
	if member != nil {
		return true
	}

	return false
}

// 当前用户对于此空间，是否有科协权限。
func (this *SpaceMemberService) canWriteBySpaceMember(user *model.User, member *model.SpaceMember) bool {
	if user.Role == model.USER_ROLE_ADMINISTRATOR {
		return true
	}

	//only space's admin can add member.
	if member != nil && (member.Role == model.SPACE_MEMBER_ROLE_ADMIN || member.Role == model.SPACE_MEMBER_ROLE_READ_WRITE) {
		return true
	}

	return false
}
