package tm

type parameter struct {
	PID      string `json:"pid"`
	Mnemonic string `json:"mnemonic"`
	Units    string `json:"units"`
}

type messagePayload struct {
	ScID       string   `json:"sc_id"`
	Stream     string   `json:"stream"`
	Action     string   `json:"action"`
	Parameters []string `json:"parameters"`
}

type request struct {
	UserID     string         `json:"user_id"`
	MsgType    string         `json:"msg_type"`
	MsgPayload messagePayload `json:"msg_payload"`
	OnChange   bool           `json:"on_change,omitempty"`
}

type Parameter struct {
	Param     string `json:"param"`
	TMCount   string `json:"tm_cnt"`
	FloatV    string `json:"float_v"`
	StringV   string `json:"str_v"`
	AlertFlag string `json:"ltm_flg"`
	Stream    string `json:"stream,omitempty"`
	OK        bool   `json:"ok,omitempty"`
	Error     string `json:"error,omitempty"`
}

type response struct {
	UserID     string          `json:"user_id"`
	MsgType    string          `json:"msg_type"`
	MsgPayload responsePayload `json:"msg_payload"`
}

type responsePayload struct {
	ScID           string      `json:"sc_id"`
	Stream         string      `json:"stream"`
	FrameTime      string      `json:"frame_time"`
	FrameID        int         `json:"frame_id"`
	Error          string      `json:"error"`
	ParametersInfo []Parameter `json:"parameters_info"`
}
