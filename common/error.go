package common

type Error struct {
	Code int    `json:"state_code"`
	Msg  string `json:"msg"`
}
