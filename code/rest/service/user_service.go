package service

import (
	"box/code/core"
	"box/code/rest/bean"
	"box/code/rest/dao"
	"box/code/rest/model"
	"box/code/tool/cache"
	"box/code/tool/result"
	"box/code/tool/util"
	"box/code/tool/uuid"
	"net/http"
	"os"
	"time"
)

// @Service
type UserService struct {
	bean.BaseBean
	userDao    *dao.UserDao
	sessionDao *dao.SessionDao

	spaceService *SpaceService

	//file lock
	locker *cache.Table

	matterDao        *dao.MatterDao
	matterService    *MatterService
	imageCacheDao    *dao.ImageCacheDao
	spaceDao         *dao.SpaceDao
	spaceMemberDao   *dao.SpaceMemberDao
	shareDao         *dao.ShareDao
	shareService     *ShareService
	downloadTokenDao *dao.DownloadTokenDao
	uploadTokenDao   *dao.UploadTokenDao
	footprintDao     *dao.FootprintDao
}

func (this *UserService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*dao.UserDao); ok {
		this.userDao = b
	}

	b = core.CONTEXT.GetBean(this.sessionDao)
	if b, ok := b.(*dao.SessionDao); ok {
		this.sessionDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceService)
	if b, ok := b.(*SpaceService); ok {
		this.spaceService = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*dao.MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*dao.ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*dao.SpaceDao); ok {
		this.spaceDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceMemberDao)
	if b, ok := b.(*dao.SpaceMemberDao); ok {
		this.spaceMemberDao = b
	}
	b = core.CONTEXT.GetBean(this.shareService)
	if b, ok := b.(*ShareService); ok {
		this.shareService = b
	}

	b = core.CONTEXT.GetBean(this.downloadTokenDao)
	if b, ok := b.(*dao.DownloadTokenDao); ok {
		this.downloadTokenDao = b
	}

	b = core.CONTEXT.GetBean(this.uploadTokenDao)
	if b, ok := b.(*dao.UploadTokenDao); ok {
		this.uploadTokenDao = b
	}

	b = core.CONTEXT.GetBean(this.footprintDao)
	if b, ok := b.(*dao.FootprintDao); ok {
		this.footprintDao = b
	}

	//create a lock cache.
	this.locker = cache.NewTable()
}

// lock a user's operation. If lock, user cannot operate file.
func (this *UserService) MatterLock(userUuid string) {

	cacheItem, err := this.locker.Value(userUuid)
	if err != nil {
		this.Logger.Error("error while get cache" + err.Error())
	}

	if cacheItem != nil && cacheItem.Data() != nil {
		panic(result.BadRequest("file is being operating, retry later"))
	}

	duration := 12 * time.Hour
	this.locker.Add(userUuid, duration, true)
}

// unlock
func (this *UserService) MatterUnlock(userUuid string) {

	exist := this.locker.Exists(userUuid)
	if exist {
		_, err := this.locker.Delete(userUuid)
		this.PanicError(err)
	} else {
		this.Logger.Error("unlock error. %s has no matter lock ", userUuid)
	}
}

// load session to SessionCache. This method will be invoked in every request.
// authorize by 1. cookie 2. username and password in request form. 3. Basic Auth
func (this *UserService) PreHandle(writer http.ResponseWriter, request *http.Request) {

	sessionId := util.GetSessionUuidFromRequest(request, core.COOKIE_AUTH_KEY)

	if sessionId != "" {

		cacheItem, err := core.CONTEXT.GetSessionCache().Value(sessionId)
		if err != nil {
			this.Logger.Error("occur error will get session cache %s", err.Error())
		}

		//if no cache. try to find in db.
		if cacheItem == nil || cacheItem.Data() == nil {
			session := this.sessionDao.FindByUuid(sessionId)
			if session != nil {
				duration := session.ExpireTime.Sub(time.Now())
				if duration <= 0 {
					this.Logger.Error("login info has expired.")
				} else {
					user := this.userDao.FindByUuid(session.UserUuid)
					if user != nil {
						core.CONTEXT.GetSessionCache().Add(sessionId, duration, user)
					} else {
						this.Logger.Error("no user with sessionId %s", session.UserUuid)
					}
				}
			}
		}
	}

	//try to auth by USERNAME_KEY PASSWORD_KEY
	cacheItem, err := core.CONTEXT.GetSessionCache().Value(sessionId)
	if err != nil {
		this.Logger.Error("occur error will get session cache %s", err.Error())
	}

	if cacheItem == nil || cacheItem.Data() == nil {
		username := request.FormValue(core.USERNAME_KEY)
		password := request.FormValue(core.PASSWORD_KEY)

		//try to read from BasicAuth
		if username == "" || password == "" {
			username, password, _ = request.BasicAuth()
		}

		if username != "" && password != "" {

			user := this.userDao.FindByUsername(username)
			if user == nil {
				this.Logger.Error("%s no such user in db.", username)
			} else {

				if !util.MatchBcrypt(password, user.Password) {
					this.Logger.Error("%s password error", username)
				} else {

					this.Logger.Info("load a temp session by username and password.")
					timeUUID, _ := uuid.NewV4()
					uuidStr := string(timeUUID.String())
					request.Form[core.COOKIE_AUTH_KEY] = []string{uuidStr}

					core.CONTEXT.GetSessionCache().Add(uuidStr, 10*time.Second, user)
				}
			}

		}
	}

}

// find a cache user by its userUuid
func (this *UserService) FindCacheUsersByUuid(userUuid string) []*model.User {

	var users []*model.User
	//let session user work.
	core.CONTEXT.GetSessionCache().Foreach(func(key interface{}, cacheItem *cache.Item) {
		if cacheItem == nil || cacheItem.Data() == nil {
			return
		}
		if value, ok := cacheItem.Data().(*model.User); ok {
			var user = value
			if user.Uuid == userUuid {
				users = append(users, user)
			}
		} else {
			this.Logger.Error("cache item not store the *User")
		}
	})

	return users
}

// remove cache user by its userUuid
func (this *UserService) RemoveCacheUserByUuid(userUuid string) {

	var sessionId interface{}
	//let session user work.
	core.CONTEXT.GetSessionCache().Foreach(func(key interface{}, cacheItem *cache.Item) {
		if cacheItem == nil || cacheItem.Data() == nil {
			return
		}
		if value, ok := cacheItem.Data().(*model.User); ok {
			var user = value
			if user.Uuid == userUuid {
				sessionId = key
				this.Logger.Info("sessionId %v", key)
			}
		} else {
			this.Logger.Error("cache item not store the *User")
		}
	})

	exists := core.CONTEXT.GetSessionCache().Exists(sessionId)
	if exists {
		_, err := core.CONTEXT.GetSessionCache().Delete(sessionId)
		if err != nil {
			this.Logger.Error("occur error when deleting cache user.")
		}
	}
}

// create user
func (this *UserService) CreateUser(request *http.Request, username string, sizeLimit int64, totalSizeLimit int64, password string, role string) *model.User {

	user := &model.User{
		Username: username,
		Password: util.GetBcrypt(password),
		Role:     role,
		Status:   model.USER_STATUS_OK,
	}

	user = this.userDao.Create(user)

	//create space.
	space := this.spaceService.CreateSpace(request, username, user, sizeLimit, totalSizeLimit, model.SPACE_TYPE_PRIVATE)

	//update user's space.
	user.SpaceUuid = space.Uuid
	this.userDao.Save(user)

	user.Space = space

	return user

}

// delete user
func (this *UserService) DeleteUser(request *http.Request, currentUser *model.User) {

	//delete from cache
	this.Logger.Info("delete from cache userUuid = %s", currentUser.Uuid)
	this.RemoveCacheUserByUuid(currentUser.Uuid)

	//delete download tokens
	this.Logger.Info("delete download tokens")
	this.downloadTokenDao.DeleteByUserUuid(currentUser.Uuid)

	//delete upload tokens
	this.Logger.Info("delete upload tokens")
	this.uploadTokenDao.DeleteByUserUuid(currentUser.Uuid)

	//delete footprints
	this.Logger.Info("delete footprints")
	this.footprintDao.DeleteByUserUuid(currentUser.Uuid)

	//delete session
	this.Logger.Info("delete session")
	this.sessionDao.DeleteByUserUuid(currentUser.Uuid)

	//delete shares and bridges
	this.Logger.Info("elete shares and bridges")
	this.shareService.DeleteSharesByUser(request, currentUser)

	//delete caches
	this.Logger.Info("delete caches")
	this.imageCacheDao.DeleteByUserUuid(currentUser.Uuid)

	//delete matters
	this.Logger.Info("delete matters")
	this.matterDao.DeleteByUserUuid(currentUser.Uuid)

	//delete space members
	space := this.spaceDao.CheckByUuid(currentUser.SpaceUuid)
	this.Logger.Info("delete space members")
	this.spaceMemberDao.DeleteBySpaceUuid(space.Uuid)

	//delete spaces
	this.Logger.Info("delete spaces")
	this.spaceDao.DeleteByUserUuid(currentUser.Uuid)

	//delete this user
	this.Logger.Info("delete this user.")
	this.userDao.Delete(currentUser)

	//delete files from disk.
	this.Logger.Info("delete files from disk. %s", model.GetUserSpaceRootDir(currentUser.Username))
	err := os.RemoveAll(model.GetUserSpaceRootDir(currentUser.Username))
	this.PanicError(err)

}
