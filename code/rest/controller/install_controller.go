package controller

import (
	"box/code/core"
	"box/code/rest/dao"
	"box/code/rest/model"
	"box/code/rest/service"
	"box/code/tool/builder"
	"box/code/tool/i18n"
	"box/code/tool/result"
	"box/code/tool/third"
	"box/code/tool/util"
	"box/code/tool/uuid"
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// install apis. Only when installing period can be visited.
type InstallController struct {
	BaseController
	uploadTokenDao    *dao.UploadTokenDao
	downloadTokenDao  *dao.DownloadTokenDao
	matterDao         *dao.MatterDao
	matterService     *service.MatterService
	imageCacheDao     *dao.ImageCacheDao
	imageCacheService *service.ImageCacheService
	tableNames        []interface{}
}

func (this *InstallController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.uploadTokenDao)
	if c, ok := b.(*dao.UploadTokenDao); ok {
		this.uploadTokenDao = c
	}

	b = core.CONTEXT.GetBean(this.downloadTokenDao)
	if c, ok := b.(*dao.DownloadTokenDao); ok {
		this.downloadTokenDao = c
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if c, ok := b.(*dao.MatterDao); ok {
		this.matterDao = c
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if c, ok := b.(*service.MatterService); ok {
		this.matterService = c
	}

	b = core.CONTEXT.GetBean(this.imageCacheDao)
	if c, ok := b.(*dao.ImageCacheDao); ok {
		this.imageCacheDao = c
	}

	b = core.CONTEXT.GetBean(this.imageCacheService)
	if c, ok := b.(*service.ImageCacheService); ok {
		this.imageCacheService = c
	}

	this.tableNames = []interface{}{
		&model.Dashboard{},
		&model.Bridge{},
		&model.DownloadToken{},
		&model.Footprint{},
		&model.ImageCache{},
		&model.Matter{},
		&model.Preference{},
		&model.Session{},
		&model.Share{},
		&model.Space{},
		&model.SpaceMember{},
		&model.UploadToken{},
		&model.User{},
	}

}

func (this *InstallController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/install/verify"] = this.Wrap(this.Verify, model.USER_ROLE_GUEST)
	routeMap["/api/install/table/info/list"] = this.Wrap(this.TableInfoList, model.USER_ROLE_GUEST)
	routeMap["/api/install/create/table"] = this.Wrap(this.CreateTable, model.USER_ROLE_GUEST)
	routeMap["/api/install/admin/list"] = this.Wrap(this.AdminList, model.USER_ROLE_GUEST)
	routeMap["/api/install/create/admin"] = this.Wrap(this.CreateAdmin, model.USER_ROLE_GUEST)
	routeMap["/api/install/validate/admin"] = this.Wrap(this.ValidateAdmin, model.USER_ROLE_GUEST)
	routeMap["/api/install/finish"] = this.Wrap(this.Finish, model.USER_ROLE_GUEST)

	return routeMap
}

func (this *InstallController) openDbConnection(writer http.ResponseWriter, request *http.Request) *gorm.DB {
	dbType := request.FormValue("dbType")
	mysqlPortStr := request.FormValue("mysqlPort")
	mysqlHost := request.FormValue("mysqlHost")
	mysqlSchema := request.FormValue("mysqlSchema")
	mysqlUsername := request.FormValue("mysqlUsername")
	mysqlPassword := request.FormValue("mysqlPassword")
	mysqlCharset := request.FormValue("mysqlCharset")

	var mysqlPort int
	if mysqlPortStr != "" {
		tmp, err := strconv.Atoi(mysqlPortStr)
		this.PanicError(err)
		mysqlPort = tmp
	}

	//log config
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // slow SQL 1s
			LogLevel:                  logger.Silent, // log level. open when debug.
			IgnoreRecordNotFoundError: true,          // ignore ErrRecordNotFound
			Colorful:                  false,         // colorful print
		},
	)
	//table name strategy
	namingStrategy := core.CONFIG.NamingStrategy()

	if dbType == "sqlite" {

		var err error = nil
		sqliteFolder := core.CONFIG.SqliteFolder() + "/tank.sqlite"
		this.Logger.Info("Connect Sqlite %s", sqliteFolder)
		db, err := gorm.Open(sqlite.Open(sqliteFolder), &gorm.Config{Logger: dbLogger, NamingStrategy: namingStrategy})

		if err != nil {
			core.LOGGER.Panic("failed to connect sqlite database")
		}

		//sqlite lock issue. https://gist.github.com/mrnugget/0eda3b2b53a70fa4a894
		phyDb, err := db.DB()
		phyDb.SetMaxOpenConns(1)

		return db

	} else {
		mysqlUrl := util.GetMysqlUrl(mysqlPort, mysqlHost, mysqlSchema, mysqlUsername, mysqlPassword, mysqlCharset)

		this.Logger.Info("Connect MySQL %s", mysqlUrl)

		var err error = nil
		db, err := gorm.Open(mysql.Open(mysqlUrl), &gorm.Config{Logger: dbLogger, NamingStrategy: namingStrategy})
		this.PanicError(err)

		return db
	}

}

func (this *InstallController) closeDbConnection(db *gorm.DB) {

	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			core.LOGGER.Error("occur error when get *sql.DB %s", err.Error())
		}
		err = sqlDB.Close()
		if err != nil {
			core.LOGGER.Error("occur error when closing db %s", err.Error())
		}
	}
}

// (tableName, exists, allFields, missingFields)
func (this *InstallController) getTableMeta(gormDb *gorm.DB, entity interface{}) (string, bool, []*model.InstallFieldInfo, []*model.InstallFieldInfo) {

	//get all useful fields from model.
	entitySchema, err := schema.Parse(entity, &sync.Map{}, core.CONFIG.NamingStrategy())
	this.PanicError(err)

	tableName := entitySchema.Table

	var allFields = make([]*model.InstallFieldInfo, 0)

	for _, field := range entitySchema.Fields {
		allFields = append(allFields, &model.InstallFieldInfo{
			Name:     field.DBName,
			DataType: string(field.DataType),
		})
	}

	var missingFields = make([]*model.InstallFieldInfo, 0)

	if !gormDb.Migrator().HasTable(tableName) {
		missingFields = append(missingFields, allFields...)

		return tableName, false, allFields, missingFields
	} else {

		for _, field := range allFields {
			//tag with `gorm:"-"` will be ""
			if field.Name != "" {

				database := gormDb.Migrator().CurrentDatabase()

				//if sqlite
				if _, ok := gormDb.Dialector.(*sqlite.Dialector); ok {

					if !gormDb.Migrator().HasColumn(tableName, field.Name) {
						missingFields = append(missingFields, field)
					}
				} else {

					if !third.MysqlMigratorHasColumn(gormDb, database, tableName, field.Name) {
						missingFields = append(missingFields, field)
					}

				}

			}
		}

		return tableName, true, allFields, missingFields
	}

}

func (this *InstallController) getTableMetaList(db *gorm.DB) []*model.InstallTableInfo {

	var installTableInfos []*model.InstallTableInfo

	for _, iBase := range this.tableNames {
		tableName, exist, allFields, missingFields := this.getTableMeta(db, iBase)

		installTableInfos = append(installTableInfos, &model.InstallTableInfo{
			Name:          tableName,
			TableExist:    exist,
			AllFields:     allFields,
			MissingFields: missingFields,
		})
	}

	return installTableInfos
}

// validate table whether integrity. if not panic err.
func (this *InstallController) validateTableMetaList(tableInfoList []*model.InstallTableInfo) {

	for _, tableInfo := range tableInfoList {
		if tableInfo.TableExist {
			if len(tableInfo.MissingFields) != 0 {

				var strs []string
				for _, v := range tableInfo.MissingFields {
					strs = append(strs, v.Name)
				}

				panic(result.BadRequest(fmt.Sprintf("table %s miss the following fields %v", tableInfo.Name, strs)))
			}
		} else {
			panic(result.BadRequest(tableInfo.Name + " table not exist"))
		}
	}

}

// Ping db.
func (this *InstallController) Verify(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	this.Logger.Info("Ping DB")
	phyDb, err := db.DB()
	this.PanicError(err)
	err = phyDb.Ping()
	this.PanicError(err)

	return this.Success("OK")
}

func (this *InstallController) TableInfoList(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	return this.Success(this.getTableMetaList(db))
}

func (this *InstallController) CreateTable(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	var installTableInfos []*model.InstallTableInfo

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	for _, iBase := range this.tableNames {

		if _, ok := db.Dialector.(*sqlite.Dialector); ok {
			//if sqlite. no need to set CHARSET.
			err := db.AutoMigrate(iBase)
			this.PanicError(err)
		} else {
			//use utf8 charset
			mysqlCharset := request.FormValue("mysqlCharset")
			err := db.Set("gorm:table_options", fmt.Sprintf("CHARSET=%s", mysqlCharset)).AutoMigrate(iBase)
			this.PanicError(err)
		}

		//complete the missing fields or create table.
		tableName, exist, allFields, missingFields := this.getTableMeta(db, iBase)
		installTableInfos = append(installTableInfos, &model.InstallTableInfo{
			Name:          tableName,
			TableExist:    exist,
			AllFields:     allFields,
			MissingFields: missingFields,
		})

	}

	return this.Success(installTableInfos)

}

// get the list of admin.
func (this *InstallController) AdminList(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	var wp = &builder.WherePair{}

	wp = wp.And(&builder.WherePair{Query: "role = ?", Args: []interface{}{model.USER_ROLE_ADMINISTRATOR}})

	var users []*model.User
	db = db.Where(wp.Query, wp.Args...).Offset(0).Limit(10).Find(&users)

	this.PanicError(db.Error)

	return this.Success(users)
}

// create admin
func (this *InstallController) CreateAdmin(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	adminUsername := request.FormValue("adminUsername")
	adminPassword := request.FormValue("adminPassword")

	//validate admin's username
	if m, _ := regexp.MatchString(model.USERNAME_PATTERN, adminUsername); !m {
		panic(result.BadRequestI18n(request, i18n.UsernameError))
	}

	if len(adminPassword) < 6 {
		panic(result.BadRequest(`admin's password at least 6 chars'`))
	}

	//check whether duplicate
	var count2 int64
	db2 := db.Model(&model.User{}).Where("username = ?", adminUsername).Count(&count2)
	this.PanicError(db2.Error)
	if count2 > 0 {
		panic(result.BadRequestI18n(request, i18n.UsernameExist, adminUsername))
	}

	space := &model.Space{}
	timeUUID, _ := uuid.NewV4()
	space.Uuid = timeUUID.String()
	space.CreateTime = time.Now()
	space.UpdateTime = time.Now()
	space.Name = adminUsername
	space.UserUuid = ""
	space.SizeLimit = -1
	space.Type = model.SPACE_TYPE_PRIVATE
	db3 := db.Create(space)
	this.PanicError(db3.Error)

	user := &model.User{}
	timeUUID, _ = uuid.NewV4()
	user.Uuid = timeUUID.String()
	user.CreateTime = time.Now()
	user.UpdateTime = time.Now()
	user.LastTime = time.Now()
	user.Sort = time.Now().UnixNano() / 1e6
	user.Role = model.USER_ROLE_ADMINISTRATOR
	user.Username = adminUsername
	user.Password = util.GetBcrypt(adminPassword)
	user.SpaceUuid = space.Uuid
	user.Status = model.USER_STATUS_OK

	db4 := db.Create(user)
	this.PanicError(db4.Error)

	space.UserUuid = user.Uuid
	db5 := db.Save(space)
	this.PanicError(db5.Error)

	return this.Success("OK")

}

// (if there is admin in db)Validate admin.
func (this *InstallController) ValidateAdmin(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	adminUsername := request.FormValue("adminUsername")
	adminPassword := request.FormValue("adminPassword")

	if adminUsername == "" {
		panic(result.BadRequest(`admin's username cannot be null'`))
	}
	if len(adminPassword) < 6 {
		panic(result.BadRequest(`admin's password at least 6 chars'`))
	}

	var existUsernameUser = &model.User{}
	db = db.Where(&model.User{Username: adminUsername}).First(existUsernameUser)
	if db.Error != nil {
		panic(result.BadRequestI18n(request, i18n.UsernameNotExist, adminUsername))
	}

	if !util.MatchBcrypt(adminPassword, existUsernameUser.Password) {
		panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordError, adminUsername))
	}

	if existUsernameUser.Role != model.USER_ROLE_ADMINISTRATOR {
		panic(result.BadRequestI18n(request, i18n.UsernameIsNotAdmin, adminUsername))
	}

	return this.Success("OK")

}

// Finish the installation
func (this *InstallController) Finish(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	dbType := request.FormValue("dbType")
	mysqlPortStr := request.FormValue("mysqlPort")
	mysqlHost := request.FormValue("mysqlHost")
	mysqlSchema := request.FormValue("mysqlSchema")
	mysqlUsername := request.FormValue("mysqlUsername")
	mysqlPassword := request.FormValue("mysqlPassword")
	mysqlCharset := request.FormValue("mysqlCharset")

	var mysqlPort int
	if mysqlPortStr != "" {
		tmp, err := strconv.Atoi(mysqlPortStr)
		this.PanicError(err)
		mysqlPort = tmp
	}

	//Recheck the db connection
	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	//Recheck the integrity of tables.
	tableMetaList := this.getTableMetaList(db)
	this.validateTableMetaList(tableMetaList)

	//At least one admin
	var count1 int64
	db1 := db.Model(&model.User{}).Where("role = ?", model.USER_ROLE_ADMINISTRATOR).Count(&count1)
	this.PanicError(db1.Error)
	if count1 == 0 {
		panic(result.BadRequest(`please config at least one admin user`))
	}

	//announce the config to write config to tank.json
	core.CONFIG.FinishInstall(dbType, mysqlPort, mysqlHost, mysqlSchema, mysqlUsername, mysqlPassword, mysqlCharset)

	//announce the context to broadcast the installation news to bean.
	core.CONTEXT.InstallOk()

	return this.Success("OK")
}
