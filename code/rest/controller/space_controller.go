package controller

import (
	"box/code/core"
	"box/code/rest/dao"
	"box/code/rest/model"
	"box/code/rest/service"
	"box/code/tool/builder"
	"box/code/tool/i18n"
	"box/code/tool/result"
	"box/code/tool/util"
	"net/http"
)

type SpaceController struct {
	BaseController
	spaceDao           *dao.SpaceDao
	spaceMemberDao     *dao.SpaceMemberDao
	spaceMemberService *service.SpaceMemberService
	matterDao          *dao.MatterDao
	matterService      *service.MatterService
	spaceService       *service.SpaceService
	userService        *service.UserService
}

func (this *SpaceController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*dao.SpaceDao); ok {
		this.spaceDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceMemberDao)
	if b, ok := b.(*dao.SpaceMemberDao); ok {
		this.spaceMemberDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceMemberService)
	if b, ok := b.(*service.SpaceMemberService); ok {
		this.spaceMemberService = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*dao.MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*service.MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.spaceService)
	if b, ok := b.(*service.SpaceService); ok {
		this.spaceService = b
	}

	b = core.CONTEXT.GetBean(this.userService)
	if b, ok := b.(*service.UserService); ok {
		this.userService = b
	}

}

func (this *SpaceController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/space/create"] = this.Wrap(this.Create, model.USER_ROLE_ADMINISTRATOR)
	routeMap["/api/space/edit"] = this.Wrap(this.Edit, model.USER_ROLE_ADMINISTRATOR)
	routeMap["/api/space/delete"] = this.Wrap(this.Delete, model.USER_ROLE_ADMINISTRATOR)
	routeMap["/api/space/detail"] = this.Wrap(this.Detail, model.USER_ROLE_USER)
	routeMap["/api/space/page"] = this.Wrap(this.Page, model.USER_ROLE_USER)
	return routeMap
}

func (this *SpaceController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	//space's name
	name := util.ExtractRequestString(request, "name")
	sizeLimit := util.ExtractRequestInt64(request, "sizeLimit")
	totalSizeLimit := util.ExtractRequestInt64(request, "totalSizeLimit")

	//create related space.
	space := this.spaceService.CreateSpace(request, name, nil, sizeLimit, totalSizeLimit, model.SPACE_TYPE_SHARED)

	return this.Success(space)
}

func (this *SpaceController) Edit(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//space's uuid
	uuid := util.ExtractRequestString(request, "uuid")
	sizeLimit := util.ExtractRequestInt64(request, "sizeLimit")
	totalSizeLimit := util.ExtractRequestInt64(request, "totalSizeLimit")

	user := this.CheckUser(request)
	space := this.spaceService.Edit(request, user, uuid, sizeLimit, totalSizeLimit)

	return this.Success(space)
}

func (this *SpaceController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//space's name
	uuid := util.ExtractRequestString(request, "uuid")
	space := this.spaceDao.CheckByUuid(uuid)

	//when space has members, cannot delete.
	memberCount := this.spaceMemberDao.CountBySpaceUuid(uuid)
	if memberCount > 0 {
		panic(result.BadRequest("space has members, cannot be deleted."))
	}

	//TODO: when space has files, cannot delete.

	//delete the space.
	this.spaceDao.Delete(space)

	return this.Success(nil)
}

func (this *SpaceController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := util.ExtractRequestString(request, "uuid")

	user := this.CheckUser(request)
	space := this.spaceDao.CheckByUuid(uuid)
	canRead := this.spaceMemberService.CanRead(user, space.Uuid)
	if !canRead {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	return this.Success(space)

}

func (this *SpaceController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	page := util.ExtractRequestOptionalInt(request, "page", 0)
	pageSize := util.ExtractRequestOptionalInt(request, "pageSize", 20)
	orderCreateTime := util.ExtractRequestOptionalString(request, "orderCreateTime", "")
	spaceType := util.ExtractRequestOptionalString(request, "type", "")
	name := util.ExtractRequestOptionalString(request, "name", "")

	user := this.CheckUser(request)

	sortArray := []builder.OrderPair{
		{
			Key:   "create_time",
			Value: orderCreateTime,
		},
	}

	var pager *model.Pager
	if user.Role == model.USER_ROLE_USER {
		if spaceType != model.SPACE_TYPE_SHARED {
			panic(result.BadRequest("user can only query shared space type."))
		}
		pager = this.spaceDao.SelfPage(page, pageSize, user.Uuid, spaceType, sortArray)
	} else if user.Role == model.USER_ROLE_ADMINISTRATOR {
		pager = this.spaceDao.Page(page, pageSize, spaceType, name, sortArray)
	}

	return this.Success(pager)
}
