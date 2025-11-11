package web

// Webix table column definitions
type WebixTableColumn struct {
	Id         string      `json:"id,omitempty"`
	Header     interface{} `json:"header,omitempty"`
	Headermenu interface{} `json:"headermenu,omitempty"`
	Editor     string      `json:"editor,omitempty"`
	Options    interface{} `json:"options,omitempty"`
	Adjust     interface{} `json:"adjust,omitempty"`
	Hidden     interface{} `json:"hidden,omitempty"`
	Sort       string      `json:"sort,omitempty"`
	Fillspace  interface{} `json:"fillspace,omitempty"`
	Css        string      `json:"css,omitempty"`
	Template   string      `json:"template,omitempty"`
	Width      int         `json:"width,omitempty"`
	Height     int         `json:"height,omitempty"`
}
