package v1

type Error struct {
	Code int    `json:"code"`
	Err  string `json:"err"`
	Data any    `json:"data,omitempty"`
}

func (e Error) Error() string {
	return e.Err
}

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func (e Response) Error() string {
	return e.Msg
}
