package dao

import (
	"kenaito-conf/server/internal/config"
	"kenaito-conf/server/internal/model"
	"kenaito-conf/server/internal/util"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

var Engine *xorm.Engine

func Init() error {
	dsn := config.Cfg.Database.DSN()

	var err error
	Engine, err = xorm.NewEngine("mysql", dsn)
	if err != nil {
		return err
	}
	Engine.SetMaxOpenConns(20)
	Engine.SetMaxIdleConns(5)
	Engine.ShowSQL(true)

	// sync table structures
	if err := Engine.Sync2(
		new(model.BusinessGroup),
		new(model.Profile),
		new(model.Application),
		new(model.AppExtrAttr),
		new(model.AppConfig),
		new(model.AppConfigHistory),
		new(model.User),
		new(model.UserPermission),
		new(model.ClientInstance),
	); err != nil {
		return err
	}

	// 迁移旧数据 & 清理格式
	migrateOldData()

	// insert initial data if tables are empty
	return initSeedData()
}

func migrateOldData() {
	// 清理 owner_ids 中可能存在的空格
	Engine.Exec("UPDATE application SET owner_ids = REPLACE(owner_ids, ' ', '') WHERE owner_ids LIKE '% %'")
	// 清理 master_uids 中可能存在的空格
	Engine.Exec("UPDATE business_group SET master_uids = REPLACE(master_uids, ' ', '') WHERE master_uids LIKE '% %'")
}

func initSeedData() error {
	// check if profiles table is empty
	count, err := Engine.Count(new(model.Profile))
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// insert profiles
	profiles := []model.Profile{
		{Name: "daily", Alias: "测试环境", IsDefault: true, Enabled: true},
		{Name: "stage", Alias: "预发布环境", IsDefault: true, Enabled: true},
		{Name: "online", Alias: "线上环境", IsDefault: true, Enabled: true},
	}
	for _, p := range profiles {
		if _, err := Engine.Insert(&p); err != nil {
			return err
		}
	}

	// insert admin user
	now := time.Now()
	hashedPassword, err := util.HashPassword("admin123")
	if err != nil {
		return err
	}
	admin := model.User{
		Name:      "admin",
		Password:  hashedPassword,
		Type:      "superAdmin",
		Enabled:   true,
		CreatedAt: &now,
	}
	if _, err := Engine.Insert(&admin); err != nil {
		return err
	}

	// insert default business group
	group := model.BusinessGroup{
		Name:      "默认组",
		Enabled:   true,
		CreatedAt: &now,
		CreatedBy: "admin",
	}
	if _, err := Engine.Insert(&group); err != nil {
		return err
	}

	return nil
}
