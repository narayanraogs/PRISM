package remote

type Acknowledgement struct {
	Status bool
	Msg    string
}

type SoftwareResponse struct {
	Ack     Acknowledgement
	Setters []Setter
	Getters []string
	Actions []Action
}

type Setter struct {
	ParamName string
	Values    []string
}

type Action struct {
	Type       string
	ParamNames []string
}

type GetRequest struct {
	Params []string
}

type GetResponse struct {
	Ack    Acknowledgement
	Params []string
	Values []string
}

type SetRequest struct {
	Params []string
	Values []string
}

type ActionRequest struct {
	Type string
}
