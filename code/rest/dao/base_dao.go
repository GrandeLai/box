package dao

import (
	"box/code/rest/bean"
	"box/code/rest/model"
	"box/code/tool/builder"
)

type BaseDao struct {
	bean.BaseBean
}

// get an order string by sortMap
func (this *BaseDao) GetSortString(sortArray []builder.OrderPair) string {

	if sortArray == nil || len(sortArray) == 0 {
		return ""
	}
	str := ""
	for _, pair := range sortArray {
		if pair.Value == model.DIRECTION_DESC || pair.Value == model.DIRECTION_ASC {
			if str != "" {
				str = str + ","
			}
			str = str + " " + pair.Key + " " + pair.Value
		}
	}

	return str
}
