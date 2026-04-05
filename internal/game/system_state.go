package game

const NumGoods = 10

type SystemState struct {
	Prices  [NumGoods]int `json:"prices"`
	Event   string        `json:"event"`
	Visited bool          `json:"visited"`
}
