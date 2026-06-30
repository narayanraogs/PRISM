package remote

type Acknowledgement struct {
	Status bool   `json:"status"`
	Msg    string `json:"msg"`
}

type SoftwareResponse struct {
	Ack     Acknowledgement `json:"ack"`
	Setters []Setter        `json:"setters"`
	Getters []string        `json:"getters"`
	Actions []Action        `json:"actions"`
}

type Setter struct {
	ParamName string   `json:"paramName"`
	Values    []string `json:"values"`
}

type Action struct {
	Type       string   `json:"type"`
	ParamNames []string `json:"paramNames"`
}

type GetRequest struct {
	Params []string `json:"params"`
}

type GetResponse struct {
	Ack    Acknowledgement `json:"ack"`
	Params []string        `json:"params"`
	Values []string        `json:"values"`
}

type SetRequest struct {
	Params []string `json:"params"`
	Values []string `json:"values"`
}

type ActionRequest struct {
	Type string `json:"type"`
}
