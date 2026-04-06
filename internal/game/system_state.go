package game

const NumGoods = 10

type SystemState struct {
	Prices   [NumGoods]int `json:"prices"`
	Event    string        `json:"event"`
	EventDay int           `json:"event_day,omitempty"`
	Visited  bool          `json:"visited"`
}
