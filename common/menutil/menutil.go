package menutil

type Menus struct {
	Id    string `json:"id"`
	Value string `json:"value"`
	Icon  string `json:"icon"`
	Url   string `json:"url,omitempty"`
	Data  []struct {
		Id    string `json:"id"`
		Value string `json:"value"`
		Icon  string `json:"icon"`
		Url   string `json:"url"`
	} `json:"data,omitempty"`
}
