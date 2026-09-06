package dao

import (
	"errors"
	"kenaito-conf/server/internal/config"
	"kenaito-conf/server/internal/model"
	"kenaito-conf/server/internal/util"
	"time"

	"github.com/gin-gonic/gin"
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

	if config.Cfg.Server.Mode == gin.DebugMode {
		Engine.ShowSQL(true)
	}

	// sync table structures
	if err := Engine.Sync2(
		new(model.BusinessGroup),
		new(model.Profile),
		new(model.Application),
		new(model.AppExtrAttr),
		new(model.AppConfig),
		new(model.AppConfigHistory),
		new(model.User),
		new(model.ClientInstance),
	); err != nil {
		return err
	}

	// insert initial data if tables are empty
	return initSeedData()
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

// Now 返回数据库当前时间（毫秒精度），用于增量同步游标和 updated_at 写入。
// 统一使用数据库时间可以避免多实例应用服务器之间的时钟偏差。
func Now() (time.Time, error) {
	rows, err := Engine.DB().Query("SELECT NOW(3)")
	if err != nil {
		return time.Time{}, err
	}
	defer rows.Close()

	var now time.Time
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return time.Time{}, err
		}
		return time.Time{}, errors.New("no database time result")
	}
	if err := rows.Scan(&now); err != nil {
		return time.Time{}, err
	}
	return now, nil
}
