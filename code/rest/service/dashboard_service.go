package service

import (
	"box/code/core"
	"box/code/rest/bean"
	"box/code/rest/dao"
	"box/code/rest/model"
	"box/code/tool/util"
	"time"
)

// @Service
type DashboardService struct {
	bean.BaseBean
	dashboardDao  *dao.DashboardDao
	footprintDao  *dao.FootprintDao
	matterDao     *dao.MatterDao
	imageCacheDao *dao.ImageCacheDao
	userDao       *dao.UserDao
}

func (this *DashboardService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.dashboardDao)
	if b, ok := b.(*dao.DashboardDao); ok {
		this.dashboardDao = b
	}

	b = core.CONTEXT.GetBean(this.footprintDao)
	if b, ok := b.(*dao.FootprintDao); ok {
		this.footprintDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*dao.MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*dao.ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*dao.UserDao); ok {
		this.userDao = b
	}

}

func (this *DashboardService) Bootstrap() {

	this.Logger.Info("Immediately ETL dashboard data.")

	//do the etl method now.
	go core.RunWithRecovery(this.Etl)
}

func (this *DashboardService) Etl() {
	this.etlOneDay(util.Yesterday())
	this.etlOneDay(time.Now())
}

// handle the dashboard data.
func (this *DashboardService) etlOneDay(thenTime time.Time) {

	startTime := util.FirstSecondOfDay(thenTime)
	endTime := util.LastSecondOfDay(thenTime)
	dt := util.ConvertTimeToDateString(startTime)
	now := time.Now()
	longTimeAgo := time.Now()
	longTimeAgo = longTimeAgo.AddDate(-20, 0, 0)

	this.Logger.Info("ETL dashboard data from %s to %s", util.ConvertTimeToDateTimeString(startTime), util.ConvertTimeToDateTimeString(endTime))

	//check whether the record has created.
	dbDashboard := this.dashboardDao.FindByDt(dt)
	if dbDashboard != nil {
		this.Logger.Info(" %s already exits. delete it and insert new one.", dt)
		this.dashboardDao.Delete(dbDashboard)
	}

	invokeNum := this.footprintDao.CountBetweenTime(startTime, endTime)
	totalInvokeNum := this.footprintDao.CountBetweenTime(longTimeAgo, now)
	uv := this.footprintDao.UvBetweenTime(startTime, endTime)
	totalUv := this.footprintDao.UvBetweenTime(longTimeAgo, now)
	matterNum := this.matterDao.CountBetweenTime(startTime, endTime)
	totalMatterNum := this.matterDao.CountBetweenTime(longTimeAgo, now)

	matterSize := this.matterDao.SizeBetweenTime(startTime, endTime)

	totalMatterSize := this.matterDao.SizeBetweenTime(longTimeAgo, now)

	cacheSize := this.imageCacheDao.SizeBetweenTime(startTime, endTime)

	totalCacheSize := this.imageCacheDao.SizeBetweenTime(longTimeAgo, now)

	avgCost := this.footprintDao.AvgCostBetweenTime(startTime, endTime)

	this.Logger.Info("Dashboard Summery 1. invokeNum = %d, totalInvokeNum = %d, UV = %d, totalUV = %d, matterNum = %d, totalMatterNum = %d",
		invokeNum, totalInvokeNum, uv, totalUv, matterNum, totalMatterNum)

	this.Logger.Info("Dashboard Summery 2. matterSize = %d, totalMatterSize = %d, cacheSize = %d, totalCacheSize = %d, avgCost = %d",
		matterSize, totalMatterSize, cacheSize, totalCacheSize, avgCost)

	dashboard := &model.Dashboard{
		InvokeNum:      invokeNum,
		TotalInvokeNum: totalInvokeNum,
		Uv:             uv,
		TotalUv:        totalUv,
		MatterNum:      matterNum,
		TotalMatterNum: totalMatterNum,
		FileSize:       matterSize + cacheSize,
		TotalFileSize:  totalMatterSize + totalCacheSize,
		AvgCost:        avgCost,
		Dt:             dt,
	}

	this.dashboardDao.Create(dashboard)
}
