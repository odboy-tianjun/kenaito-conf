package model

type WrapperResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

type PageResult struct {
	PageNo   int         `json:"pageNo"`
	PageSize int         `json:"pageSize"`
	Total    int64       `json:"total"`
	Data     interface{} `json:"data"`
}

type SelectOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type Menu struct {
	Title    string    `json:"title"`
	Icon     string    `json:"icon"`
	Active   bool      `json:"active"`
	Children []SubMenu `json:"children"`
}

type SubMenu struct {
	Title string `json:"title"`
	Icon  string `json:"icon"`
	Url   string `json:"url"`
}

func NewMenu(title, icon string, active bool) Menu {
	return Menu{Title: title, Icon: icon, Active: active, Children: make([]SubMenu, 0)}
}

func (m *Menu) AddSubMenu(title, icon, url string) {
	m.Children = append(m.Children, SubMenu{Title: title, Icon: icon, Url: url})
}
