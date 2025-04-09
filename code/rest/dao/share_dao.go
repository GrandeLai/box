package dao

import (
	"box/code/core"
	"box/code/rest/model"
	"box/code/tool/builder"
	"box/code/tool/result"
	"gorm.io/gorm"

	"box/code/tool/uuid"
	"time"
)

type ShareDao struct {
	BaseDao
}

// find by uuid. if not found return nil.
func (this *ShareDao) FindByUuid(uuid string) *model.Share {
	var entity = &model.Share{}
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
func (this *ShareDao) CheckByUuid(uuid string) *model.Share {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity
}

func (this *ShareDao) Page(page int, pageSize int, userUuid string, sortArray []builder.OrderPair) *model.Pager {

	count, shares := this.PlainPage(page, pageSize, userUuid, sortArray)
	pager := model.NewPager(page, pageSize, count, shares)

	return pager
}

func (this *ShareDao) PlainPage(page int, pageSize int, userUuid string, sortArray []builder.OrderPair) (int, []*model.Share) {

	var wp = &builder.WherePair{}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	var conditionDB *gorm.DB
	conditionDB = core.CONTEXT.GetDB().Model(&model.Share{}).Where(wp.Query, wp.Args...)

	var count int64 = 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var shares []*model.Share
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&shares)
	this.PanicError(db.Error)

	return int(count), shares
}

func (this *ShareDao) Create(share *model.Share) *model.Share {

	timeUUID, _ := uuid.NewV4()
	share.Uuid = string(timeUUID.String())
	share.CreateTime = time.Now()
	share.UpdateTime = time.Now()
	share.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(share)
	this.PanicError(db.Error)

	return share
}

func (this *ShareDao) Save(share *model.Share) *model.Share {

	share.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(share)
	this.PanicError(db.Error)

	return share
}

func (this *ShareDao) Delete(share *model.Share) {

	db := core.CONTEXT.GetDB().Delete(&share)
	this.PanicError(db.Error)

}

// System cleanup.
func (this *ShareDao) Cleanup() {
	this.Logger.Info("[ShareDao] clean up. Delete all Share")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(model.Share{})
	this.PanicError(db.Error)
}
