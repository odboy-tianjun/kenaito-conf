package model

import "time"

type BusinessGroup struct {
	Id        int        `xorm:"int autoincr pk" json:"id"`
	Name      string     `xorm:"varchar(32) notnull" json:"name"`
	Master    string     `xorm:"varchar(16)" json:"master"`
	MasterUid *int       `xorm:"int" json:"masterUid"`
	Remarks   string     `xorm:"text" json:"remarks"`
	Enabled   bool       `xorm:"tinyint(1) default 1" json:"enabled"`
	CreatedAt *time.Time `xorm:"datetime" json:"createdAt"`
	CreatedBy string     `xorm:"varchar(32)" json:"createdBy"`
	UpdatedAt *time.Time `xorm:"datetime" json:"updatedAt"`
	UpdatedBy string     `xorm:"varchar(32)" json:"updatedBy"`
}
