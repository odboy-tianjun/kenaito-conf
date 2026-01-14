package model

import "time"

type User struct {
	Id           int        `xorm:"int autoincr pk" json:"id"`
	GroupId      *int       `xorm:"int" json:"groupId"`
	Name         string     `xorm:"varchar(16)" json:"name"`
	Password     string     `xorm:"varchar(64)" json:"-"`
	Mobile       string     `xorm:"char(11)" json:"mobile"`
	Email        string     `xorm:"varchar(32)" json:"email"`
	Type         string     `xorm:"varchar(16) default 'user'" json:"type"` // user, groupAdmin, superAdmin
	Enabled      bool       `xorm:"tinyint(1) default 1" json:"enabled"`
	CreatedAt    *time.Time `xorm:"datetime" json:"createdAt"`
	UpdatedAt    *time.Time `xorm:"datetime" json:"updatedAt"`
	GroupName    string     `xorm:"-" json:"groupName,omitempty"`
	GroupMaster  bool       `xorm:"-" json:"groupMaster,omitempty"`
}

type UserPermission struct {
	Id      int    `xorm:"int autoincr pk" json:"id"`
	UserId  int    `xorm:"int" json:"userId"`
	Env     string `xorm:"varchar(16)" json:"env"`
	AppId   int    `xorm:"int" json:"appId"`
	Operate string `xorm:"varchar(8) default 'RW'" json:"operate"` // RO, RW
}

type LoginUserInfo struct {
	Id              string   `json:"id"`
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	GroupId         *int     `json:"groupId"`
	GrantAppIds     []int    `json:"grantAppIds"`
	GrantPermissons []string `json:"grantPermissons"`
	GroupMaster     bool     `json:"groupMaster"`
}

func (u *LoginUserInfo) IsSuperAdmin() bool {
	return u.Type == "superAdmin"
}

func (u *LoginUserInfo) IsGroupAdmin() bool {
	return u.Type == "groupAdmin"
}
