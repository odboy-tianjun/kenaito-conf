package model

import "time"

type AppConfig struct {
	Id        int        `xorm:"int autoincr pk" json:"id"`
	GroupId   *int       `xorm:"int" json:"groupId"`
	AppId     *int       `xorm:"int" json:"appId"`
	AppCode   string     `xorm:"varchar(32)" json:"appCode"`
	Env       string     `xorm:"varchar(16)" json:"env"`
	Version   string     `xorm:"varchar(32)" json:"version"`
	Name      string     `xorm:"varchar(32)" json:"name"`
	Type      int        `xorm:"smallint(1)" json:"type"` // 1:file, 2:kv, 3:json
	Contents  string     `xorm:"text" json:"contents"`
	Enabled   bool       `xorm:"tinyint(1) default 1" json:"enabled"`
	CreatedAt *time.Time `xorm:"datetime" json:"createdAt"`
	CreatedBy string     `xorm:"varchar(32)" json:"createdBy"`
	UpdatedAt *time.Time `xorm:"datetime" json:"updatedAt"`
	UpdatedBy string     `xorm:"varchar(32)" json:"updatedBy"`
	Global    bool       `xorm:"-" json:"global"`
	AppNames  string     `xorm:"-" json:"appNames,omitempty"`
	ClientCount int      `xorm:"-" json:"clientCount,omitempty"`
}

type AppConfigHistory struct {
	Id            int        `xorm:"int autoincr pk" json:"id"`
	OriginId      int        `xorm:"int" json:"originId"`
	AppId         *int       `xorm:"int" json:"appId"`
	AppName       string     `xorm:"varchar(32)" json:"appName"`
	AppNames      string     `xorm:"varchar(64)" json:"appNames"`
	Env           string     `xorm:"varchar(16)" json:"env"`
	Version       string     `xorm:"varchar(32)" json:"version"`
	Name          string     `xorm:"varchar(32)" json:"name"`
	Type          int        `xorm:"smallint(1)" json:"type"`
	Contents      string     `xorm:"text" json:"contents"`
	ActiveContents string   `xorm:"-" json:"activeContents,omitempty"`
	CreatedAt     *time.Time `xorm:"datetime" json:"createdAt"`
	CreatedBy     string     `xorm:"varchar(32)" json:"createdBy"`
}
