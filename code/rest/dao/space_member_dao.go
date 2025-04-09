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

type SpaceMemberDao struct {
	BaseDao
}

// find by uuid. if not found return nil.
func (this *SpaceMemberDao) FindByUuid(uuid string) *model.SpaceMember {
	var entity = &model.SpaceMember{}
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
func (this *SpaceMemberDao) CheckByUuid(uuid string) *model.SpaceMember {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity
}

// find by spaceUuid and userUuid. if not found return nil.
func (this *SpaceMemberDao) FindBySpaceUuidAndUserUuid(spaceUuid string, userUuid string) *model.SpaceMember {
	var entity = &model.SpaceMember{}
	db := core.CONTEXT.GetDB().Where("space_uuid = ? AND user_uuid = ?", spaceUuid, userUuid).First(entity)

	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return entity
}

func (this *SpaceMemberDao) Page(page int, pageSize int, spaceUuid string, sortArray []builder.OrderPair) *model.Pager {

	count, spaceMembers := this.PlainPage(page, pageSize, spaceUuid, sortArray)
	pager := model.NewPager(page, pageSize, count, spaceMembers)

	return pager
}

func (this *SpaceMemberDao) PlainPage(page int, pageSize int, spaceUuid string, sortArray []builder.OrderPair) (int, []*model.SpaceMember) {

	var wp = &builder.WherePair{}

	if spaceUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "space_uuid = ?", Args: []interface{}{spaceUuid}})
	}

	var conditionDB *gorm.DB
	conditionDB = core.CONTEXT.GetDB().Model(&model.SpaceMember{}).Where(wp.Query, wp.Args...)

	var count int64 = 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var spaceMembers []*model.SpaceMember
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&spaceMembers)
	this.PanicError(db.Error)

	return int(count), spaceMembers
}

func (this *SpaceMemberDao) Create(spaceMember *model.SpaceMember) *model.SpaceMember {

	timeUUID, _ := uuid.NewV4()
	spaceMember.Uuid = string(timeUUID.String())
	spaceMember.CreateTime = time.Now()
	spaceMember.UpdateTime = time.Now()
	spaceMember.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(spaceMember)
	this.PanicError(db.Error)

	return spaceMember
}

func (this *SpaceMemberDao) Save(spaceMember *model.SpaceMember) *model.SpaceMember {

	spaceMember.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(spaceMember)
	this.PanicError(db.Error)

	return spaceMember
}

func (this *SpaceMemberDao) Delete(spaceMember *model.SpaceMember) {

	db := core.CONTEXT.GetDB().Delete(&spaceMember)
	this.PanicError(db.Error)

}

func (this *SpaceMemberDao) DeleteBySpaceUuid(spaceUuid string) {

	var wp = &builder.WherePair{}

	wp = wp.And(&builder.WherePair{Query: "space_uuid = ?", Args: []interface{}{spaceUuid}})

	db := core.CONTEXT.GetDB().Where(wp.Query, wp.Args).Delete(model.SpaceMember{})
	this.PanicError(db.Error)
}

func (this *SpaceMemberDao) CountBySpaceUuid(spaceUuid string) int {
	var count int64
	db := core.CONTEXT.GetDB().
		Model(&model.SpaceMember{}).
		Where("space_uuid = ?", spaceUuid).
		Count(&count)
	this.PanicError(db.Error)
	return int(count)
}

// System cleanup.
func (this *SpaceMemberDao) Cleanup() {
	this.Logger.Info("[SpaceMemberDao] clean up. Delete all SpaceMember")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(model.SpaceMember{})
	this.PanicError(db.Error)
}
