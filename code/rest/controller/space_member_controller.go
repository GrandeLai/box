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
	"strings"
)

type SpaceMemberController struct {
	BaseController
	spaceMemberDao     *dao.SpaceMemberDao
	spaceDao           *dao.SpaceDao
	bridgeDao          *dao.BridgeDao
	matterDao          *dao.MatterDao
	matterService      *service.MatterService
	spaceMemberService *service.SpaceMemberService
}

func (this *SpaceMemberController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.spaceMemberDao)
	if b, ok := b.(*dao.SpaceMemberDao); ok {
		this.spaceMemberDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*dao.SpaceDao); ok {
		this.spaceDao = b
	}

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*dao.BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*dao.MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*service.MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.spaceMemberService)
	if b, ok := b.(*service.SpaceMemberService); ok {
		this.spaceMemberService = b
	}

}

func (this *SpaceMemberController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//admin user can create/edit/delete
	routeMap["/api/space/member/create"] = this.Wrap(this.Create, model.USER_ROLE_USER)
	routeMap["/api/space/member/edit"] = this.Wrap(this.Edit, model.USER_ROLE_USER)
	routeMap["/api/space/member/delete"] = this.Wrap(this.Delete, model.USER_ROLE_USER)

	routeMap["/api/space/member/detail"] = this.Wrap(this.Detail, model.USER_ROLE_USER)
	routeMap["/api/space/member/mine"] = this.Wrap(this.Mine, model.USER_ROLE_USER)
	routeMap["/api/space/member/page"] = this.Wrap(this.Page, model.USER_ROLE_USER)

	return routeMap
}

func (this *SpaceMemberController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	spaceUuid := util.ExtractRequestString(request, "spaceUuid")
	userUuidsStr := util.ExtractRequestString(request, "userUuids")
	spaceRole := util.ExtractRequestString(request, "role")

	if spaceRole != model.SPACE_MEMBER_ROLE_READ_ONLY && spaceRole != model.SPACE_MEMBER_ROLE_READ_WRITE && spaceRole != model.SPACE_MEMBER_ROLE_ADMIN {
		panic("spaceRole is not correct")
	}

	//validate userUuids
	if userUuidsStr == "" {
		panic("userUuids is required")
	}
	userUuids := strings.Split(userUuidsStr, ",")

	// check operator's permission
	currentUser := this.CheckUser(request)
	canManage := this.spaceMemberService.CanManage(currentUser, spaceUuid)
	if !canManage {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	//check whether exists.
	for _, userUuid := range userUuids {
		spaceMember := this.spaceMemberDao.FindBySpaceUuidAndUserUuid(spaceUuid, userUuid)
		user := this.userDao.CheckByUuid(userUuid)
		if spaceMember != nil {
			panic(result.BadRequestI18n(request, i18n.SpaceMemberExist, user.Username))
		}
	}

	//check whether space exists.
	space := this.spaceDao.CheckByUuid(spaceUuid)

	//check whether exists.
	for _, userUuid := range userUuids {
		user := this.userDao.CheckByUuid(userUuid)
		this.spaceMemberService.CreateMember(space, user, spaceRole)
	}

	return this.Success("OK")
}

func (this *SpaceMemberController) Edit(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	uuid := util.ExtractRequestString(request, "uuid")
	spaceRole := util.ExtractRequestString(request, "role")

	if spaceRole != model.SPACE_MEMBER_ROLE_READ_ONLY && spaceRole != model.SPACE_MEMBER_ROLE_READ_WRITE && spaceRole != model.SPACE_MEMBER_ROLE_ADMIN {
		panic("spaceRole is not correct")
	}

	spaceMember := this.spaceMemberDao.CheckByUuid(uuid)

	currentUser := this.CheckUser(request)
	canManage := this.spaceMemberService.CanManage(currentUser, spaceMember.SpaceUuid)
	if !canManage {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	spaceMember.Role = spaceRole
	spaceMember = this.spaceMemberDao.Save(spaceMember)

	return this.Success(spaceMember)
}

func (this *SpaceMemberController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	uuid := util.ExtractRequestString(request, "uuid")

	spaceMember := this.spaceMemberDao.CheckByUuid(uuid)
	user := this.CheckUser(request)
	canManage := this.spaceMemberService.CanManage(user, spaceMember.SpaceUuid)
	if !canManage {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	this.spaceMemberDao.Delete(spaceMember)

	return this.Success("OK")
}

func (this *SpaceMemberController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := util.ExtractRequestString(request, "uuid")

	spaceMember := this.spaceMemberDao.CheckByUuid(uuid)

	user := this.CheckUser(request)

	if spaceMember.UserUuid != user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	return this.Success(spaceMember)

}

// find my role in the space.
func (this *SpaceMemberController) Mine(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	spaceUuid := util.ExtractRequestString(request, "spaceUuid")

	user := this.CheckUser(request)
	spaceMember := this.spaceMemberDao.FindBySpaceUuidAndUserUuid(spaceUuid, user.Uuid)
	if spaceMember == nil {
		spaceMember = &model.SpaceMember{SpaceUuid: spaceUuid, Role: model.SPACE_MEMBER_GUEST}
	}
	if user.Role == model.USER_ROLE_ADMINISTRATOR {
		spaceMember.Role = model.SPACE_MEMBER_ROLE_ADMIN
	}

	return this.Success(spaceMember)

}

func (this *SpaceMemberController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	page := util.ExtractRequestOptionalInt(request, "page", 0)
	pageSize := util.ExtractRequestOptionalInt(request, "pageSize", 20)
	orderCreateTime := util.ExtractRequestOptionalString(request, "orderCreateTime", "")
	spaceUuid := util.ExtractRequestString(request, "spaceUuid")

	user := this.CheckUser(request)
	canRead := this.spaceMemberService.CanRead(user, spaceUuid)
	if !canRead {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	sortArray := []builder.OrderPair{
		{
			Key:   "create_time",
			Value: orderCreateTime,
		},
	}

	pager := this.spaceMemberDao.Page(page, pageSize, spaceUuid, sortArray)

	//fill the space's user. FIXME: user better way to get User.
	if pager != nil {
		for _, spaceMember := range pager.Data.([]*model.SpaceMember) {
			spaceMember.User = this.userDao.FindByUuid(spaceMember.UserUuid)
		}
	}

	return this.Success(pager)
}
