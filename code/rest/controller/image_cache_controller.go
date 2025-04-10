package controller

import (
	"box/code/core"
	"box/code/rest/dao"
	"box/code/rest/model"
	"box/code/rest/service"
	"box/code/tool/builder"
	"box/code/tool/result"
	"net/http"
	"strconv"
	"strings"
)

type ImageCacheController struct {
	BaseController
	imageCacheDao     *dao.ImageCacheDao
	imageCacheService *service.ImageCacheService
}

func (this *ImageCacheController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*dao.ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = core.CONTEXT.GetBean(this.imageCacheService)
	if b, ok := b.(*service.ImageCacheService); ok {
		this.imageCacheService = b
	}

}

func (this *ImageCacheController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/image/cache/delete"] = this.Wrap(this.Delete, model.USER_ROLE_USER)
	routeMap["/api/image/cache/delete/batch"] = this.Wrap(this.DeleteBatch, model.USER_ROLE_USER)
	routeMap["/api/image/cache/detail"] = this.Wrap(this.Detail, model.USER_ROLE_USER)
	routeMap["/api/image/cache/page"] = this.Wrap(this.Page, model.USER_ROLE_USER)

	return routeMap
}

func (this *ImageCacheController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("uuid cannot be null"))
	}

	imageCache := this.imageCacheService.Detail(uuid)

	user := this.CheckUser(request)
	if imageCache.UserUuid != user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	return this.Success(imageCache)

}

func (this *ImageCacheController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	orderCreateTime := request.FormValue("orderCreateTime")
	orderUpdateTime := request.FormValue("orderUpdateTime")
	orderSort := request.FormValue("orderSort")

	userUuid := request.FormValue("userUuid")
	matterUuid := request.FormValue("matterUuid")
	orderSize := request.FormValue("orderSize")

	user := this.CheckUser(request)
	userUuid = user.Uuid

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
			Key:   "size",
			Value: orderSize,
		},
	}

	pager := this.imageCacheDao.Page(page, pageSize, userUuid, matterUuid, sortArray)

	return this.Success(pager)
}

func (this *ImageCacheController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("uuid cannot be null"))
	}

	imageCache := this.imageCacheDao.FindByUuid(uuid)

	user := this.CheckUser(request)
	if imageCache.UserUuid != user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	this.imageCacheDao.Delete(imageCache)

	return this.Success("OK")
}

func (this *ImageCacheController) DeleteBatch(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuids := request.FormValue("uuids")
	if uuids == "" {
		panic(result.BadRequest("uuids cannot be null"))
	}

	uuidArray := strings.Split(uuids, ",")

	for _, uuid := range uuidArray {

		imageCache := this.imageCacheDao.FindByUuid(uuid)

		user := this.CheckUser(request)
		if imageCache.UserUuid != user.Uuid {
			panic(result.UNAUTHORIZED)
		}

		this.imageCacheDao.Delete(imageCache)

	}

	return this.Success("OK")
}
