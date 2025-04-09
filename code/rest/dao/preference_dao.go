package dao

import (
	"box/code/core"
	"box/code/rest/model"
	"box/code/tool/result"
	"box/code/tool/uuid"
	"time"
)

type PreferenceDao struct {
	BaseDao
}

// find by uuid. if not found return nil.
func (this *PreferenceDao) Fetch() *model.Preference {

	// Read
	var preference = &model.Preference{}
	db := core.CONTEXT.GetDB().First(preference)
	if db.Error != nil {

		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			preference.Name = "EyeblueTank"
			preference.Version = core.VERSION
			preference.PreviewConfig = "{}"
			preference.ScanConfig = "{}"
			this.Create(preference)
			return preference
		} else {
			return nil
		}
	}

	preference.Version = core.VERSION
	return preference
}

func (this *PreferenceDao) Create(preference *model.Preference) *model.Preference {

	timeUUID, _ := uuid.NewV4()
	preference.Uuid = string(timeUUID.String())
	preference.CreateTime = time.Now()
	preference.UpdateTime = time.Now()
	preference.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(preference)
	this.PanicError(db.Error)

	return preference
}

func (this *PreferenceDao) Save(preference *model.Preference) *model.Preference {

	preference.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(preference)
	this.PanicError(db.Error)

	return preference
}

// System cleanup.
func (this *PreferenceDao) Cleanup() {

	this.Logger.Info("[PreferenceDao] clean up. Delete all Preference")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(model.Preference{})
	this.PanicError(db.Error)
}
