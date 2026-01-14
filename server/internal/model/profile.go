package model

type Profile struct {
	Id        int    `xorm:"int autoincr pk" json:"id"`
	Name      string `xorm:"varchar(32)" json:"name"`
	Alias     string `xorm:"varchar(32)" json:"alias"`
	IsDefault bool   `xorm:"tinyint(1) default 0" json:"isDefault"`
	Enabled   bool   `xorm:"tinyint(1) default 1" json:"enabled"`
}

type AppExtrAttr struct {
	AppId    int    `xorm:"int pk" json:"appId"`
	Env      string `xorm:"varchar(16) pk" json:"env"`
	AttrName string `xorm:"varchar(32) pk" json:"attrName"`
	AttrValue string `xorm:"varchar(100)" json:"attrValue"`
}
