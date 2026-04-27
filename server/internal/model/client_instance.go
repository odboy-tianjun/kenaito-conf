package model

import "time"

type ClientInstance struct {
	Id            int        `xorm:"int autoincr pk" json:"id"`
	AppConfigId   int        `xorm:"int index" json:"appConfigId"`
	AppName       string     `xorm:"varchar(64) index" json:"appName"`
	Env           string     `xorm:"varchar(32)" json:"env"`
	IP            string     `xorm:"varchar(64)" json:"ip"`
	LastFetchAt   *time.Time `xorm:"datetime" json:"lastFetchAt"`
	LastFetchType string     `xorm:"varchar(16)" json:"lastFetchType"`
	CreatedAt     *time.Time `xorm:"datetime" json:"createdAt"`
	UpdatedAt     *time.Time `xorm:"datetime" json:"updatedAt"`
}
