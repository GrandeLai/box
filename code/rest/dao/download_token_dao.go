package dao

import (
	"box/code/core"
	"box/code/rest/model"
	"box/code/tool/result"
	"box/code/tool/uuid"
	"time"
)

type DownloadTokenDao struct {
	BaseDao
}

// find by uuid. if not found return nil.
func (this *DownloadTokenDao) FindByUuid(uuid string) *model.DownloadToken {
	var entity = &model.DownloadToken{}
	db := core.CONTEXT.GetDB().Where("uuid = ?", uuid).First(entity)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return entity

}

// find by uuid. if not found panic NotFound error
func (this *DownloadTokenDao) CheckByUuid(uuid string) *model.DownloadToken {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity
}

func (this *DownloadTokenDao) Create(downloadToken *model.DownloadToken) *model.DownloadToken {

	timeUUID, _ := uuid.NewV4()
	downloadToken.Uuid = string(timeUUID.String())

	downloadToken.CreateTime = time.Now()
	downloadToken.UpdateTime = time.Now()
	downloadToken.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(downloadToken)
	this.PanicError(db.Error)

	return downloadToken
}

func (this *DownloadTokenDao) Save(downloadToken *model.DownloadToken) *model.DownloadToken {

	downloadToken.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(downloadToken)
	this.PanicError(db.Error)

	return downloadToken
}

func (this *DownloadTokenDao) DeleteByUserUuid(userUuid string) {

	db := core.CONTEXT.GetDB().Where("user_uuid = ?", userUuid).Delete(model.DownloadToken{})
	this.PanicError(db.Error)

}

func (this *DownloadTokenDao) Cleanup() {
	this.Logger.Info("[DownloadTokenDao] clean up. Delete all DownloadToken")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(model.DownloadToken{})
	this.PanicError(db.Error)
}
