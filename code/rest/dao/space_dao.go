package dao

import (
	"box/code/core"
	"box/code/rest/model"
	"box/code/tool/builder"
	"box/code/tool/result"
	"fmt"
	"gorm.io/gorm"
	"math"

	"box/code/tool/uuid"
	"time"
)

type SpaceDao struct {
	BaseDao
}

// find by uuid. if not found return nil.
func (this *SpaceDao) FindByUuid(uuid string) *model.Space {
	var entity = &model.Space{}
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
func (this *SpaceDao) CheckByUuid(uuid string) *model.Space {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity
}

func (this *SpaceDao) CountByName(name string) int {
	var count int64
	db := core.CONTEXT.GetDB().
		Model(&model.Space{}).
		Where("name = ?", name).
		Count(&count)
	this.PanicError(db.Error)
	return int(count)
}

func (this *SpaceDao) FindByName(name string) *model.Space {

	var space = &model.Space{}
	db := core.CONTEXT.GetDB().Where(&model.Space{Name: name}).First(space)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return space
}

func (this *SpaceDao) CountByUserUuid(userUuid string) int {
	var count int64
	db := core.CONTEXT.GetDB().
		Model(&model.Space{}).
		Where("user_uuid = ?", userUuid).
		Count(&count)
	this.PanicError(db.Error)
	return int(count)
}

// TODO:
func (this *SpaceDao) SelfPage(page int, pageSize int, userUuid string, spaceType string, sortArray []builder.OrderPair) *model.Pager {

	countSqlTemplate := fmt.Sprintf("SELECT COUNT(*) FROM `%sspace` WHERE uuid IN (SELECT space_uuid FROM `%sspace_member` WHERE user_uuid = ?) AND type = ?", core.TABLE_PREFIX, core.TABLE_PREFIX)
	if spaceType == model.SPACE_TYPE_PRIVATE {
		countSqlTemplate = fmt.Sprintf("SELECT COUNT(*) FROM `%sspace` WHERE user_uuid = ? AND type = ?", core.TABLE_PREFIX)
	}
	var count int
	core.CONTEXT.GetDB().Raw(countSqlTemplate, userUuid, spaceType).Scan(&count)

	orderByString := this.GetSortString(sortArray)
	if orderByString == "" {
		orderByString = "uuid"
	}
	querySqlTemplate := fmt.Sprintf("SELECT * FROM `%sspace` WHERE uuid IN (SELECT space_uuid FROM `%sspace_member` WHERE user_uuid = ?) AND type = ? ORDER BY ? LIMIT ?,?", core.TABLE_PREFIX, core.TABLE_PREFIX)
	if spaceType == model.SPACE_TYPE_PRIVATE {
		querySqlTemplate = fmt.Sprintf("SELECT * FROM `%sspace` WHERE user_uuid = ? AND type = ? ORDER BY ? LIMIT ?,?", core.TABLE_PREFIX)
	}
	var spaces []*model.Space
	core.CONTEXT.GetDB().Raw(querySqlTemplate, userUuid, spaceType, orderByString, page*pageSize, pageSize).Scan(&spaces)

	pager := model.NewPager(page, pageSize, count, spaces)

	return pager

}

func (this *SpaceDao) Page(page int, pageSize int, spaceType string, name string, sortArray []builder.OrderPair) *model.Pager {
	count, spaces := this.PlainPage(page, pageSize, spaceType, name, sortArray)
	pager := model.NewPager(page, pageSize, count, spaces)

	return pager
}

func (this *SpaceDao) PlainPage(page int, pageSize int, spaceType string, name string, sortArray []builder.OrderPair) (int, []*model.Space) {

	var wp = &builder.WherePair{}

	if spaceType != "" {
		wp = wp.And(&builder.WherePair{Query: "type = ?", Args: []interface{}{spaceType}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name LIKE ?", Args: []interface{}{"%" + name + "%"}})
	}

	var conditionDB *gorm.DB
	conditionDB = core.CONTEXT.GetDB().Model(&model.Space{}).Where(wp.Query, wp.Args...)

	var count int64 = 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var spaces []*model.Space
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&spaces)
	this.PanicError(db.Error)

	return int(count), spaces
}

func (this *SpaceDao) Create(space *model.Space) *model.Space {

	timeUUID, _ := uuid.NewV4()
	space.Uuid = string(timeUUID.String())
	space.CreateTime = time.Now()
	space.UpdateTime = time.Now()
	space.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(space)
	this.PanicError(db.Error)

	return space
}

func (this *SpaceDao) Save(space *model.Space) *model.Space {

	space.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(space)
	this.PanicError(db.Error)

	return space
}

func (this *SpaceDao) UpdateTotalSize(spaceUuid string, totalSize int64) {
	db := core.CONTEXT.GetDB().Model(&model.Space{}).Where("uuid = ?", spaceUuid).Update("total_size", totalSize)
	this.PanicError(db.Error)
}

// handle user page by page.
func (this *SpaceDao) PageHandle(fun func(space *model.Space)) {

	pageSize := 1000
	sortArray := []builder.OrderPair{
		{
			Key:   "uuid",
			Value: model.DIRECTION_ASC,
		},
	}
	count, _ := this.PlainPage(0, pageSize, "", "", sortArray)
	if count > 0 {
		var totalPages = int(math.Ceil(float64(count) / float64(pageSize)))
		var page int
		for page = 0; page < totalPages; page++ {
			_, spaces := this.PlainPage(0, pageSize, "", "", sortArray)
			for _, space := range spaces {
				fun(space)
			}
		}
	}
}

func (this *SpaceDao) DeleteByUserUuid(userUuid string) {

	db := core.CONTEXT.GetDB().Where("user_uuid = ?", userUuid).Delete(model.Space{})
	this.PanicError(db.Error)

}

func (this *SpaceDao) Delete(space *model.Space) {

	db := core.CONTEXT.GetDB().Delete(&space)
	this.PanicError(db.Error)

}

// System cleanup.
func (this *SpaceDao) Cleanup() {
	this.Logger.Info("[SpaceDao] clean up. Delete all Space")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(model.Space{})
	this.PanicError(db.Error)
}
