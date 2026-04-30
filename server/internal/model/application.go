package model

import "time"

type Application struct {
	Id         int        `xorm:"int autoincr pk" json:"id"`
	GroupId    *int       `xorm:"int" json:"groupId"`
	Name       string     `xorm:"varchar(32)" json:"name"`
	Code       string     `xorm:"varchar(32)" json:"code"`
	ServiceId  string     `xorm:"varchar(64)" json:"serviceId"`
	Token      string     `xorm:"varchar(64)" json:"token"`
	OwnerId    *int       `xorm:"int" json:"ownerId"`
	OwnerName  string     `xorm:"varchar(16)" json:"ownerName"`
	Remarks    string     `xorm:"text" json:"remarks"`
	Enabled    bool       `xorm:"tinyint(1) default 1" json:"enabled"`
	CreatedAt  *time.Time `xorm:"datetime" json:"createdAt"`
	CreatedBy  string     `xorm:"varchar(32)" json:"createdBy"`
	UpdatedAt  *time.Time `xorm:"datetime" json:"updatedAt"`
	UpdatedBy  string     `xorm:"varchar(32)" json:"updatedBy"`
	GroupName  string     `xorm:"-" json:"groupName,omitempty"`
}
