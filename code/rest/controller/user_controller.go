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
	"regexp"
	"strconv"
	"time"
)

type UserController struct {
	BaseController
	preferenceService *service.PreferenceService
	userService       *service.UserService
	spaceDao          *dao.SpaceDao
	spaceService      *service.SpaceService
	matterService     *service.MatterService
}

func (this *UserController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.preferenceService)
	if b, ok := b.(*service.PreferenceService); ok {
		this.preferenceService = b
	}

	b = core.CONTEXT.GetBean(this.userService)
	if b, ok := b.(*service.UserService); ok {
		this.userService = b
	}

	b = core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*dao.SpaceDao); ok {
		this.spaceDao = b
	}
	b = core.CONTEXT.GetBean(this.spaceService)
	if b, ok := b.(*service.SpaceService); ok {
		this.spaceService = b
	}
	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*service.MatterService); ok {
		this.matterService = b
	}

}

func (this *UserController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/user/info"] = this.Wrap(this.Info, model.USER_ROLE_GUEST)
	routeMap["/api/user/login"] = this.Wrap(this.Login, model.USER_ROLE_GUEST)
	routeMap["/api/user/authentication/login"] = this.Wrap(this.AuthenticationLogin, model.USER_ROLE_GUEST)
	routeMap["/api/user/register"] = this.Wrap(this.Register, model.USER_ROLE_GUEST)
	routeMap["/api/user/create"] = this.Wrap(this.Create, model.USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/edit"] = this.Wrap(this.Edit, model.USER_ROLE_USER)
	routeMap["/api/user/detail"] = this.Wrap(this.Detail, model.USER_ROLE_USER)
	routeMap["/api/user/logout"] = this.Wrap(this.Logout, model.USER_ROLE_GUEST)
	routeMap["/api/user/change/password"] = this.Wrap(this.ChangePassword, model.USER_ROLE_USER)
	routeMap["/api/user/reset/password"] = this.Wrap(this.ResetPassword, model.USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/page"] = this.Wrap(this.Page, model.USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/search"] = this.Wrap(this.Search, model.USER_ROLE_USER)
	routeMap["/api/user/toggle/status"] = this.Wrap(this.ToggleStatus, model.USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/transfiguration"] = this.Wrap(this.Transfiguration, model.USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/scan"] = this.Wrap(this.Scan, model.USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/delete"] = this.Wrap(this.Delete, model.USER_ROLE_ADMINISTRATOR)

	return routeMap
}

func (this *UserController) innerLogin(writer http.ResponseWriter, request *http.Request, user *model.User) {

	if user.Status == model.USER_STATUS_DISABLED {
		panic(result.BadRequestI18n(request, i18n.UserDisabled))
	}

	//set cookie. expire after 30 days.
	expiration := time.Now()
	expiration = expiration.AddDate(0, 0, 30)

	//save session to db.
	session := &model.Session{
		UserUuid:   user.Uuid,
		Ip:         util.GetIpAddress(request),
		ExpireTime: expiration,
	}
	session.UpdateTime = time.Now()
	session.CreateTime = time.Now()
	session = this.sessionDao.Create(session)

	//set cookie
	cookie := http.Cookie{
		Name:    core.COOKIE_AUTH_KEY,
		Path:    "/",
		Value:   session.Uuid,
		Expires: expiration}
	http.SetCookie(writer, &cookie)

	//update lastTime and lastIp
	user.LastTime = time.Now()
	user.LastIp = util.GetIpAddress(request)
	this.userDao.Save(user)
}

// login by username and password
func (this *UserController) Login(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	username := request.FormValue("username")
	password := request.FormValue("password")

	if "" == username || "" == password {
		panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordCannotNull))
	}

	user := this.userDao.FindByUsername(username)
	if user == nil {
		panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordError))
	}

	if !util.MatchBcrypt(password, user.Password) {
		panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordError))
	}
	this.innerLogin(writer, request, user)

	//append the space info.
	space := this.spaceDao.FindByUuid(user.SpaceUuid)
	user.Space = space

	return this.Success(user)
}

// login by authentication.
func (this *UserController) AuthenticationLogin(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	authentication := request.FormValue("authentication")
	if authentication == "" {
		panic(result.BadRequest("authentication cannot be null"))
	}
	session := this.sessionDao.FindByUuid(authentication)
	if session == nil {
		panic(result.BadRequest("authentication error"))
	}
	duration := session.ExpireTime.Sub(time.Now())
	if duration <= 0 {
		panic(result.BadRequest("login info has expired"))
	}

	user := this.userDao.CheckByUuid(session.UserUuid)
	this.innerLogin(writer, request, user)
	return this.Success(user)
}

// fetch current user's info.
func (this *UserController) Info(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	user := this.CheckUser(request)

	//append the space info.
	space := this.spaceDao.FindByUuid(user.SpaceUuid)
	user.Space = space

	return this.Success(user)
}

// register by username and password. After registering, will auto login.
func (this *UserController) Register(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	username := request.FormValue("username")
	password := request.FormValue("password")

	preference := this.preferenceService.Fetch()
	if !preference.AllowRegister {
		panic(result.BadRequestI18n(request, i18n.UserRegisterNotAllowd))
	}

	if m, _ := regexp.MatchString(model.USERNAME_PATTERN, username); !m {
		panic(result.BadRequestI18n(request, i18n.UsernameError))
	}

	if len(password) < 6 {
		panic(result.BadRequestI18n(request, i18n.UserPasswordLengthError))
	}

	if this.userDao.CountByUsername(username) > 0 {
		panic(result.BadRequestI18n(request, i18n.UsernameExist, username))
	}

	user := this.userService.CreateUser(request, username, -1, preference.DefaultTotalSizeLimit, password, model.USER_ROLE_USER)

	//auto login
	this.innerLogin(writer, request, user)

	return this.Success(user)
}

func (this *UserController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	username := request.FormValue("username")
	password := request.FormValue("password")
	role := request.FormValue("role")

	sizeLimit := util.ExtractRequestInt64(request, "sizeLimit")
	totalSizeLimit := util.ExtractRequestInt64(request, "totalSizeLimit")

	//validation work.
	if m, _ := regexp.MatchString(model.USERNAME_PATTERN, username); !m {
		panic(result.BadRequestI18n(request, i18n.UsernameError))
	}

	if len(password) < 6 {
		panic(result.BadRequestI18n(request, i18n.UserPasswordLengthError))
	}

	if this.userDao.CountByUsername(username) > 0 {
		panic(result.BadRequestI18n(request, i18n.UsernameExist, username))
	}
	if this.spaceDao.CountByName(username) > 0 {
		panic(result.BadRequestI18n(request, i18n.SpaceNameExist, username))
	}

	//check user role.
	if role != model.USER_ROLE_USER && role != model.USER_ROLE_ADMINISTRATOR {
		panic(result.BadRequestI18n(request, i18n.UserRoleError))
	}

	user := this.userService.CreateUser(request, username, sizeLimit, totalSizeLimit, password, role)

	return this.Success(user)
}

func (this *UserController) Edit(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	avatarUrl := request.FormValue("avatarUrl")
	role := request.FormValue("role")

	sizeLimit := util.ExtractRequestInt64(request, "sizeLimit")
	totalSizeLimit := util.ExtractRequestInt64(request, "totalSizeLimit")

	operator := this.CheckUser(request)
	currentUser := this.userDao.CheckByUuid(uuid)

	currentUser.AvatarUrl = avatarUrl

	if operator.Role == model.USER_ROLE_ADMINISTRATOR {
		//only admin can edit user's role and sizeLimit

		if role == model.USER_ROLE_USER || role == model.USER_ROLE_ADMINISTRATOR {
			currentUser.Role = role
		}

	} else if operator.Uuid == uuid {
		//cannot edit sizeLimit, totalSizeLimit
		space := this.spaceDao.CheckByUuid(currentUser.SpaceUuid)
		if space.SizeLimit != sizeLimit {
			this.Logger.Error(" %s try to modify sizeLimit from %d to %d.", operator.Uuid, space.SizeLimit, sizeLimit)
			panic(result.BadRequestI18n(request, i18n.PermissionDenied))
		}
		if space.TotalSizeLimit != totalSizeLimit {
			this.Logger.Error(" %s try to modify TotalSizeLimit from %d to %d.", operator.Uuid, space.TotalSizeLimit, totalSizeLimit)
			panic(result.BadRequestI18n(request, i18n.PermissionDenied))
		}

	} else {
		panic(result.UNAUTHORIZED)
	}

	//edit user's info
	currentUser = this.userDao.Save(currentUser)

	//edit user's private space info.
	space := this.spaceService.Edit(request, operator, currentUser.SpaceUuid, sizeLimit, totalSizeLimit)

	//remove cache user.
	this.userService.RemoveCacheUserByUuid(currentUser.Uuid)

	currentUser.Space = space

	return this.Success(currentUser)
}

func (this *UserController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")

	user := this.userDao.CheckByUuid(uuid)

	//append the space info.
	space := this.spaceDao.FindByUuid(user.SpaceUuid)
	user.Space = space

	return this.Success(user)

}

func (this *UserController) Logout(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//try to find from SessionCache.
	sessionId := util.GetSessionUuidFromRequest(request, core.COOKIE_AUTH_KEY)
	if sessionId == "" {
		return nil
	}

	user := this.FindUser(request)
	if user != nil {
		session := this.sessionDao.FindByUuid(sessionId)
		session.ExpireTime = time.Now()
		this.sessionDao.Save(session)
	}

	//delete session.
	_, err := core.CONTEXT.GetSessionCache().Delete(sessionId)
	if err != nil {
		this.Logger.Error("error while deleting session.")
	}

	//clear cookie.
	expiration := time.Now()
	expiration = expiration.AddDate(-1, 0, 0)
	cookie := http.Cookie{
		Name:    core.COOKIE_AUTH_KEY,
		Path:    "/",
		Value:   sessionId,
		Expires: expiration}
	http.SetCookie(writer, &cookie)

	return this.Success("OK")
}

func (this *UserController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	orderCreateTime := request.FormValue("orderCreateTime")
	orderUpdateTime := request.FormValue("orderUpdateTime")
	orderSort := request.FormValue("orderSort")

	username := request.FormValue("username")
	status := request.FormValue("status")
	orderLastTime := request.FormValue("orderLastTime")

	var page int
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	pageSize := 200
	if pageSizeStr != "" {
		tmp, err := strconv.Atoi(pageSizeStr)
		if err == nil {
			pageSize = tmp
		}
	}

	sortArray := []builder.OrderPair{
		{
			Key:   "create_time",
			Value: orderCreateTime,
		},
		{
			Key:   "update_time",
			Value: orderUpdateTime,
		},
		{
			Key:   "sort",
			Value: orderSort,
		},
		{
			Key:   "last_time",
			Value: orderLastTime,
		},
	}

	pager := this.userDao.Page(page, pageSize, username, status, sortArray)

	//append the space info. FIXME: user better way to get Space.
	for _, u := range pager.Data.([]*model.User) {
		space := this.spaceDao.FindByUuid(u.SpaceUuid)
		u.Space = space
	}

	return this.Success(pager)
}

func (this *UserController) Search(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	keyword := request.FormValue("keyword")

	pager := this.userDao.Page(0, 10, keyword, "", nil)

	var resultList []*model.User = make([]*model.User, 0)
	for _, u := range pager.Data.([]*model.User) {
		resultList = append(resultList, &model.User{
			Uuid:     u.Uuid,
			Username: u.Username,
		})
	}

	return this.Success(resultList)
}

func (this *UserController) ToggleStatus(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	currentUser := this.userDao.CheckByUuid(uuid)
	user := this.CheckUser(request)
	if uuid == user.Uuid {
		panic(result.BadRequest("You cannot disable yourself."))
	}

	if currentUser.Status == model.USER_STATUS_OK {
		currentUser.Status = model.USER_STATUS_DISABLED
	} else if currentUser.Status == model.USER_STATUS_DISABLED {
		currentUser.Status = model.USER_STATUS_OK
	}

	currentUser = this.userDao.Save(currentUser)

	//remove cache user.
	this.userService.RemoveCacheUserByUuid(currentUser.Uuid)

	return this.Success(currentUser)

}

func (this *UserController) Transfiguration(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	currentUser := this.userDao.CheckByUuid(uuid)

	//expire after 10 minutes.
	expiration := time.Now()
	expiration = expiration.Add(10 * time.Minute)

	session := &model.Session{
		UserUuid:   currentUser.Uuid,
		Ip:         util.GetIpAddress(request),
		ExpireTime: expiration,
	}
	session.UpdateTime = time.Now()
	session.CreateTime = time.Now()
	session = this.sessionDao.Create(session)

	return this.Success(session.Uuid)
}

// scan user's physics files. create index into EyeblueTank
func (this *UserController) Scan(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	currentUser := this.userDao.CheckByUuid(uuid)
	space := this.spaceDao.CheckByUuid(currentUser.SpaceUuid)
	this.matterService.DeleteByPhysics(request, currentUser, space)
	this.matterService.ScanPhysics(request, currentUser, space)

	return this.Success("OK")
}

func (this *UserController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	currentUser := this.userDao.CheckByUuid(uuid)
	user := this.CheckUser(request)

	if currentUser.Status != model.USER_STATUS_DISABLED {
		panic(result.BadRequest("Only disabled user can be deleted."))
	}
	if currentUser.Uuid == user.Uuid {
		panic(result.BadRequest("You cannot delete yourself."))
	}

	this.userService.DeleteUser(request, currentUser)

	return this.Success("OK")
}

func (this *UserController) ChangePassword(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	oldPassword := request.FormValue("oldPassword")
	newPassword := request.FormValue("newPassword")
	if oldPassword == "" || newPassword == "" {
		panic(result.BadRequest("oldPassword and newPassword cannot be null"))
	}

	user := this.CheckUser(request)

	//if username is demo, cannot change password.
	if user.Username == model.USERNAME_DEMO {
		return this.Success(user)
	}

	if !util.MatchBcrypt(oldPassword, user.Password) {
		panic(result.BadRequestI18n(request, i18n.UserOldPasswordError))
	}

	user.Password = util.GetBcrypt(newPassword)

	user = this.userDao.Save(user)

	return this.Success(user)
}

// admin reset password.
func (this *UserController) ResetPassword(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	userUuid := request.FormValue("userUuid")
	password := request.FormValue("password")
	if userUuid == "" {
		panic(result.BadRequest("userUuid cannot be null"))
	}
	if password == "" {
		panic(result.BadRequest("password cannot be null"))
	}

	currentUser := this.CheckUser(request)

	if currentUser.Role != model.USER_ROLE_ADMINISTRATOR {
		panic(result.UNAUTHORIZED)
	}

	user := this.userDao.CheckByUuid(userUuid)

	user.Password = util.GetBcrypt(password)

	user = this.userDao.Save(user)

	return this.Success(currentUser)
}
