package model

import "time"

type ClientInstance struct {
	Id            int        `xorm:"int autoincr pk" json:"id"`
	AppConfigId   int        `xorm:"int unique(udx_aaei)" json:"appConfigId"`
	AppName       string     `xorm:"varchar(64) unique(udx_aaei)" json:"appName"`
	Env           string     `xorm:"varchar(32) unique(udx_aaei)" json:"env"`
	Ip            string     `xorm:"varchar(64) unique(udx_aaei)" json:"ip"`
	LastPingAt    *time.Time `xorm:"datetime" json:"lastPingAt"`
	LastFetchAt   *time.Time `xorm:"datetime" json:"lastFetchAt"`
	LastFetchType string     `xorm:"varchar(16)" json:"lastFetchType"`
	CreatedAt     *time.Time `xorm:"datetime" json:"createdAt"`
	UpdatedAt     *time.Time `xorm:"datetime" json:"updatedAt"`
}
