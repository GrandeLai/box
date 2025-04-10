package controller

import (
	"box/code/core"
	"box/code/rest/dao"
	"box/code/rest/model"
	"box/code/rest/service"
	"box/code/tool/i18n"
	"box/code/tool/result"
	"box/code/tool/util"
	"net/http"
	"sort"
	"strings"
)

type MatterController struct {
	BaseController
	matterDao         *dao.MatterDao
	matterService     *service.MatterService
	preferenceService *service.PreferenceService
	downloadTokenDao  *dao.DownloadTokenDao
	imageCacheDao     *dao.ImageCacheDao
	shareDao          *dao.ShareDao
	spaceDao          *dao.SpaceDao
	shareService      *service.ShareService
	bridgeDao         *dao.BridgeDao
	imageCacheService *service.ImageCacheService
}

func (this *MatterController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*dao.MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*service.MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.downloadTokenDao)
	if b, ok := b.(*dao.DownloadTokenDao); ok {
		this.downloadTokenDao = b
	}

	b = core.CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*dao.ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = core.CONTEXT.GetBean(this.shareDao)
	if b, ok := b.(*dao.ShareDao); ok {
		this.shareDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*dao.SpaceDao); ok {
		this.spaceDao = b
	}

	b = core.CONTEXT.GetBean(this.shareService)
	if b, ok := b.(*service.ShareService); ok {
		this.shareService = b
	}

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*dao.BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.imageCacheService)
	if b, ok := b.(*service.ImageCacheService); ok {
		this.imageCacheService = b
	}
}

func (this *MatterController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))
	routeMap["/api/matter/detail"] = this.Wrap(this.Detail, model.USER_ROLE_USER)
	routeMap["/api/matter/page"] = this.Wrap(this.Page, model.USER_ROLE_USER)
	routeMap["/api/matter/search"] = this.Wrap(this.Search, model.USER_ROLE_USER)

	routeMap["/api/matter/create/directory"] = this.Wrap(this.CreateDirectory, model.USER_ROLE_USER)
	routeMap["/api/matter/upload"] = this.Wrap(this.Upload, model.USER_ROLE_USER)
	routeMap["/api/matter/crawl"] = this.Wrap(this.Crawl, model.USER_ROLE_USER)
	routeMap["/api/matter/soft/delete"] = this.Wrap(this.SoftDelete, model.USER_ROLE_USER)
	routeMap["/api/matter/soft/delete/batch"] = this.Wrap(this.SoftDeleteBatch, model.USER_ROLE_USER)
	routeMap["/api/matter/recovery"] = this.Wrap(this.Recovery, model.USER_ROLE_USER)
	routeMap["/api/matter/recovery/batch"] = this.Wrap(this.RecoveryBatch, model.USER_ROLE_USER)
	routeMap["/api/matter/delete"] = this.Wrap(this.Delete, model.USER_ROLE_USER)
	routeMap["/api/matter/delete/batch"] = this.Wrap(this.DeleteBatch, model.USER_ROLE_USER)
	routeMap["/api/matter/clean/expired/deleted/matters"] = this.Wrap(this.CleanExpiredDeletedMatters, model.USER_ROLE_ADMINISTRATOR)
	routeMap["/api/matter/rename"] = this.Wrap(this.Rename, model.USER_ROLE_USER)
	routeMap["/api/matter/change/privacy"] = this.Wrap(this.ChangePrivacy, model.USER_ROLE_USER)
	routeMap["/api/matter/move"] = this.Wrap(this.Move, model.USER_ROLE_USER)

	//mirror local files.
	routeMap["/api/matter/mirror"] = this.Wrap(this.Mirror, model.USER_ROLE_USER)
	routeMap["/api/matter/zip"] = this.Wrap(this.Zip, model.USER_ROLE_USER)

	return routeMap
}

func (this *MatterController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := util.ExtractRequestString(request, "uuid")
	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckReadableByUuid(request, user, spaceUuid)

	matter := this.matterService.Detail(request, uuid)

	if matter.SpaceUuid != space.Uuid {
		panic(result.UNAUTHORIZED)
	}

	//add the user's info.
	if space.Uuid == user.SpaceUuid {
		matter.User = user
	} else {
		matter.User = this.userDao.FindByUuid(matter.UserUuid)
	}

	return this.Success(matter)

}

func (this *MatterController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	page := util.ExtractRequestOptionalInt(request, "page", 0)
	pageSize := util.ExtractRequestOptionalInt(request, "pageSize", 200)
	orderCreateTime := util.ExtractRequestOptionalString(request, "orderCreateTime", "")
	orderUpdateTime := util.ExtractRequestOptionalString(request, "orderUpdateTime", "")
	orderDeleteTime := util.ExtractRequestOptionalString(request, "orderDeleteTime", "")
	orderSort := util.ExtractRequestOptionalString(request, "orderSort", "")
	orderTimes := util.ExtractRequestOptionalString(request, "orderTimes", "")
	orderDir := util.ExtractRequestOptionalString(request, "orderDir", "")
	orderSize := util.ExtractRequestOptionalString(request, "orderSize", "")
	orderName := util.ExtractRequestOptionalString(request, "orderName", "")

	puuid := util.ExtractRequestOptionalString(request, "puuid", "")
	name := util.ExtractRequestOptionalString(request, "name", "")
	dir := util.ExtractRequestOptionalString(request, "dir", "")
	deleted := util.ExtractRequestOptionalString(request, "deleted", "")
	extensionsStr := util.ExtractRequestOptionalString(request, "extensions", "")

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	this.spaceService.CheckReadableByUuid(request, user, spaceUuid)

	var extensions []string
	if extensionsStr != "" {
		extensions = strings.Split(extensionsStr, ",")
	}

	pager := this.matterService.Page(
		request,
		page,
		pageSize,
		orderCreateTime,
		orderUpdateTime,
		orderDeleteTime,
		orderSort,
		orderTimes,
		orderDir,
		orderSize,
		orderName,
		puuid,
		name,
		dir,
		deleted,
		extensions,
		spaceUuid,
	)

	return this.Success(pager)
}

// DFS search.
func (this *MatterController) Search(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	limit := util.ExtractRequestOptionalInt(request, "limit", 200)
	puuid := util.ExtractRequestString(request, "puuid")
	keyword := util.ExtractRequestString(request, "keyword")
	deletedStr := util.ExtractRequestOptionalString(request, "deleted", "")
	var deleted = false
	if deletedStr == model.TRUE {
		deleted = true
	}

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	this.spaceService.CheckReadableByUuid(request, user, spaceUuid)

	matters := this.matterService.DfsSearch(
		request,
		limit,
		puuid,
		keyword,
		spaceUuid,
		deleted,
	)

	//sort.
	//从小到大自定义排序
	sort.Slice(matters, func(i, j int) bool {
		matter1 := matters[i]
		matter2 := matters[j]
		if matter1.Dir {
			return true
		} else if matter2.Dir {
			return false
		} else {
			cmp := strings.Compare(matter1.Name, matter2.Name)
			if cmp < 0 {
				return true
			} else if cmp == 0 {
				return true
			} else {
				return false
			}
		}
	})

	return this.Success(matters)
}

func (this *MatterController) CreateDirectory(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	puuid := util.ExtractRequestString(request, "puuid")
	name := util.ExtractRequestString(request, "name")
	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	var dirMatter = this.matterDao.CheckWithRootByUuid(puuid, space)

	matter := this.matterService.AtomicCreateDirectory(request, dirMatter, name, user, space)
	return this.Success(matter)
}

func (this *MatterController) Upload(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	puuid := util.ExtractRequestString(request, "puuid")
	privacy := util.ExtractRequestOptionalBool(request, "privacy", true)

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	file, handler, err := request.FormFile("file")
	this.PanicError(err)
	defer func() {
		err := file.Close()
		this.PanicError(err)
	}()

	err = request.ParseMultipartForm(32 << 20)
	this.PanicError(err)

	//for IE browser. filename may contain filepath.
	fileName := handler.Filename
	pos := strings.LastIndex(fileName, "\\")
	if pos != -1 {
		fileName = fileName[pos+1:]
	}
	pos = strings.LastIndex(fileName, "/")
	if pos != -1 {
		fileName = fileName[pos+1:]
	}

	dirMatter := this.matterDao.CheckWithRootByUuid(puuid, space)

	//support upload simultaneously
	matter := this.matterService.Upload(request, file, handler, user, space, dirMatter, fileName, privacy)

	return this.Success(matter)
}

// crawl a file by url.
func (this *MatterController) Crawl(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	url := util.ExtractRequestString(request, "url")
	destPath := util.ExtractRequestOptionalString(request, "destPath", "")
	puuid := util.ExtractRequestOptionalString(request, "puuid", "")
	filename := util.ExtractRequestString(request, "filename")

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	var dirMatter *model.Matter
	if puuid != "" {
		dirMatter = this.matterDao.CheckWithRootByUuid(puuid, space)
		if dirMatter.SpaceUuid != space.Uuid {
			panic(result.UNAUTHORIZED)
		}
		if !dirMatter.Dir {
			panic(" puuid is not a dir.")
		}
	} else {
		if destPath == "" {
			panic(" puuid or destPath cannot be null")
		}
		dirMatter = this.matterService.CreateDirectories(request, user, space, destPath)
	}

	if url == "" || (!strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")) {
		panic(" url must start with  http:// or https://")
	}

	matter := this.matterService.AtomicCrawl(request, url, filename, user, space, dirMatter, true)

	return this.Success(matter)
}

// soft delete.
func (this *MatterController) SoftDelete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := util.ExtractRequestString(request, "uuid")

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	matter := this.matterDao.CheckByUuid(uuid)
	if matter.SpaceUuid != space.Uuid {
		panic(result.UNAUTHORIZED)
	}

	this.matterService.AtomicSoftDelete(request, matter, user, space)

	return this.Success("OK")
}

func (this *MatterController) SoftDeleteBatch(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuids := util.ExtractRequestString(request, "uuids")

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	uuidArray := strings.Split(uuids, ",")

	matters := make([]*model.Matter, 0)
	for _, uuid := range uuidArray {

		matter := this.matterDao.FindByUuid(uuid)

		if matter == nil {
			this.Logger.Warn("%s not exist anymore", uuid)
			continue
		}

		if matter.SpaceUuid != spaceUuid {
			panic(result.UNAUTHORIZED)
		}

		matters = append(matters, matter)
	}

	for _, matter := range matters {
		this.matterService.AtomicSoftDelete(request, matter, user, space)
	}

	return this.Success("OK")
}

// recovery delete.
func (this *MatterController) Recovery(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := util.ExtractRequestString(request, "uuid")

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	matter := this.matterDao.CheckByUuid(uuid)

	if matter.SpaceUuid != space.Uuid {
		panic(result.UNAUTHORIZED)
	}

	this.matterService.AtomicRecovery(request, matter, user)

	return this.Success("OK")
}

// recovery batch.
func (this *MatterController) RecoveryBatch(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	uuids := util.ExtractRequestString(request, "uuids")

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	uuidArray := strings.Split(uuids, ",")

	for _, uuid := range uuidArray {

		matter := this.matterDao.FindByUuid(uuid)

		if matter == nil {
			this.Logger.Warn("%s not exist anymore", uuid)
			continue
		}

		if matter.SpaceUuid != space.Uuid {
			panic(result.UNAUTHORIZED)
		}

		this.matterService.AtomicRecovery(request, matter, user)

	}

	return this.Success("OK")
}

// complete delete.
func (this *MatterController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := util.ExtractRequestString(request, "uuid")

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	matter := this.matterDao.CheckByUuid(uuid)
	if matter.SpaceUuid != space.Uuid {
		panic(result.UNAUTHORIZED)
	}

	this.matterService.AtomicDelete(request, matter, user, space)

	return this.Success("OK")
}

func (this *MatterController) DeleteBatch(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuids := util.ExtractRequestString(request, "uuids")

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	uuidArray := strings.Split(uuids, ",")
	matters := make([]*model.Matter, 0)
	for _, uuid := range uuidArray {

		matter := this.matterDao.FindByUuid(uuid)

		if matter == nil {
			this.Logger.Warn("%s not exist anymore", uuid)
			continue
		}

		if matter.SpaceUuid != space.Uuid {
			panic(result.UNAUTHORIZED)
		}

		matters = append(matters, matter)
	}

	for _, matter := range matters {

		this.matterService.AtomicDelete(request, matter, user, space)
	}

	return this.Success("OK")
}

// manual clean expired deleted matters.
func (this *MatterController) CleanExpiredDeletedMatters(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	this.matterService.CleanExpiredDeletedMatters()

	return this.Success("OK")
}

func (this *MatterController) Rename(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := util.ExtractRequestString(request, "uuid")
	name := util.ExtractRequestString(request, "name")

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	matter := this.matterDao.CheckByUuid(uuid)

	if matter.SpaceUuid != space.Uuid {
		panic(result.UNAUTHORIZED)
	}

	this.matterService.AtomicRename(request, matter, name, false, user, space)

	return this.Success(matter)
}

func (this *MatterController) ChangePrivacy(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := util.ExtractRequestString(request, "uuid")
	privacy := util.ExtractRequestOptionalBool(request, "privacy", false)

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	matter := this.matterDao.CheckByUuid(uuid)

	if matter.Deleted {
		panic(result.BadRequest("matter has been deleted. Cannot change privacy."))
	}

	if matter.Privacy == privacy {
		panic(result.BadRequest("not changed. Invalid operation."))
	}

	if matter.SpaceUuid != space.Uuid {
		panic(result.UNAUTHORIZED)
	}

	matter.Privacy = privacy
	this.matterDao.Save(matter)

	return this.Success("OK")
}

func (this *MatterController) Move(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	srcUuidsStr := util.ExtractRequestString(request, "srcUuids")
	destUuid := util.ExtractRequestString(request, "destUuid")
	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	var srcUuids []string
	srcUuids = strings.Split(srcUuidsStr, ",")

	var destMatter = this.matterDao.CheckWithRootByUuid(destUuid, space)
	if !destMatter.Dir {
		panic(result.BadRequest("destination is not a directory"))
	}

	if destMatter.SpaceUuid != space.Uuid {
		panic(result.UNAUTHORIZED)
	}

	if destMatter.Deleted {
		panic(result.BadRequest("dest matter has been deleted. Cannot move."))
	}

	var srcMatters []*model.Matter
	for _, uuid := range srcUuids {
		srcMatter := this.matterDao.CheckByUuid(uuid)

		if srcMatter.Puuid == destMatter.Uuid {
			panic(result.BadRequest("no move, invalid operation"))
		}

		if srcMatter.Deleted {
			panic(result.BadRequest("src matter has been deleted. Cannot move."))
		}

		//check whether there are files with the same name.
		count := this.matterDao.CountByUserUuidAndPuuidAndDirAndName(user.Uuid, destMatter.Uuid, srcMatter.Dir, srcMatter.Name)

		if count > 0 {
			panic(result.BadRequestI18n(request, i18n.MatterExist, srcMatter.Name))
		}

		if srcMatter.SpaceUuid != destMatter.SpaceUuid {
			panic("space not the same")
		}

		srcMatters = append(srcMatters, srcMatter)
	}

	this.matterService.AtomicMoveBatch(request, srcMatters, destMatter, user, space)

	return this.Success(nil)
}

// mirror local files to EyeblueTank
func (this *MatterController) Mirror(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	srcPath := util.ExtractRequestString(request, "srcPath")
	destPath := util.ExtractRequestString(request, "destPath")
	overwrite := util.ExtractRequestOptionalBool(request, "overwrite", false)

	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	this.matterService.AtomicMirror(request, srcPath, destPath, overwrite, user, space)

	return this.Success(nil)

}

// download zip.
func (this *MatterController) Zip(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	uuids := util.ExtractRequestString(request, "uuids")
	user := this.CheckUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckReadableByUuid(request, user, spaceUuid)

	uuidArray := strings.Split(uuids, ",")

	matters := this.matterDao.FindByUuids(uuidArray, nil)

	if matters == nil || len(matters) == 0 {
		panic(result.BadRequest("matters cannot be nil."))
	}

	for _, matter := range matters {
		if matter.Deleted {
			panic(result.BadRequest("matter has been deleted. Cannot download batch."))
		}
	}

	puuid := matters[0].Puuid

	for _, m := range matters {
		if m.SpaceUuid != space.Uuid {
			panic(result.UNAUTHORIZED)
		} else if m.Puuid != puuid {
			panic(result.BadRequest("puuid not same"))
		}
	}

	this.matterService.DownloadZip(writer, request, matters)

	return nil
}
