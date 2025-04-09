package service

import (
	"box/code/core"
	"box/code/rest/bean"
	"box/code/rest/dao"
	"box/code/rest/model"
	"box/code/tool/i18n"
	"box/code/tool/result"
	"net/http"
	"regexp"
)

// @Service
type SpaceService struct {
	bean.BaseBean
	spaceDao           *dao.SpaceDao
	spaceMemberService *SpaceMemberService
	matterDao          *dao.MatterDao
	bridgeDao          *dao.BridgeDao
	userDao            *dao.UserDao
}

func (this *SpaceService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*dao.SpaceDao); ok {
		this.spaceDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceMemberService)
	if b, ok := b.(*SpaceMemberService); ok {
		this.spaceMemberService = b
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

func (this *SpaceService) Detail(uuid string) *model.Space {

	space := this.spaceDao.CheckByUuid(uuid)

	return space
}

// create space
func (this *SpaceService) CreateSpace(
	request *http.Request,
	name string,
	user *model.User,
	sizeLimit int64,
	totalSizeLimit int64,
	spaceType string) *model.Space {

	userUuid := ""
	//validation work.
	if m, _ := regexp.MatchString(model.USERNAME_PATTERN, name); !m {
		panic(result.BadRequestI18n(request, i18n.SpaceNameError))
	}

	if spaceType == model.SPACE_TYPE_PRIVATE {
		if user == nil {
			panic("private space requires user.")
		}

		userUuid = user.Uuid
		if this.spaceDao.CountByUserUuid(userUuid) > 0 {
			panic(result.BadRequestI18n(request, i18n.SpaceExclusive, name))
		}

	} else if spaceType == model.SPACE_TYPE_SHARED {

	} else {
		panic("Not supported spaceType:" + spaceType)
	}

	if this.spaceDao.CountByName(name) > 0 {
		panic(result.BadRequestI18n(request, i18n.SpaceNameExist, name))
	}

	space := &model.Space{
		Name:           name,
		UserUuid:       userUuid,
		SizeLimit:      sizeLimit,
		TotalSizeLimit: totalSizeLimit,
		TotalSize:      0,
		Type:           spaceType,
	}

	space = this.spaceDao.Create(space)

	return space

}

// checkout a adminAble space.
func (this *SpaceService) CheckAdminAbleByUuid(request *http.Request, user *model.User, spaceUuid string) *model.Space {
	space := this.spaceDao.CheckByUuid(spaceUuid)
	if space.Type == model.SPACE_TYPE_PRIVATE && user.Uuid == space.UserUuid {
		return space
	}

	manage := this.spaceMemberService.CanManage(user, spaceUuid)
	if !manage {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	return space
}

// checkout a writable space.
func (this *SpaceService) CheckWritableByUuid(request *http.Request, user *model.User, spaceUuid string) *model.Space {
	space := this.spaceDao.CheckByUuid(spaceUuid)
	if space.Type == model.SPACE_TYPE_PRIVATE && user.Uuid == space.UserUuid {
		return space
	}

	writable := this.spaceMemberService.canWrite(user, spaceUuid)
	if !writable {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	return space
}

// checkout a readable space.
func (this *SpaceService) CheckReadableByUuid(request *http.Request, user *model.User, spaceUuid string) *model.Space {
	space := this.spaceDao.CheckByUuid(spaceUuid)
	if space.Type == model.SPACE_TYPE_PRIVATE && user.Uuid == space.UserUuid {
		return space
	}

	manage := this.spaceMemberService.CanRead(user, spaceUuid)
	if !manage {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	return space
}

// edit space's info
func (this *SpaceService) Edit(request *http.Request, user *model.User, spaceUuid string, sizeLimit int64, totalSizeLimit int64) *model.Space {
	space := this.CheckAdminAbleByUuid(request, user, spaceUuid)

	if sizeLimit < 0 && sizeLimit != -1 {
		panic("sizeLimit cannot be negative expect -1.")
	}

	if totalSizeLimit < 0 && totalSizeLimit != -1 {
		panic("totalSizeLimit cannot be negative expect -1.")
	}

	space.SizeLimit = sizeLimit
	space.TotalSizeLimit = totalSizeLimit
	space = this.spaceDao.Save(space)

	return space
}
