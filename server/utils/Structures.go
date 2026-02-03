package utils

type Command struct {
	SlNo          int
	Mnemonic      string
	Command       string
	Delay         float64
	Argument      bool
	Read          bool
	ReadBinary    bool
	DataType      string
	Component     string
	Port          string
	Packet        []byte
	ArgumentValue string
}

type Component struct {
	SlNo          int
	ComponentName string
	ComponentType string
	ComponentCode int64
}

type CommandResponse struct {
	Success      bool
	ErrorMessage string
	Result       map[string]CommandResult
}

type CommandResult struct {
	ResultType string
	String     string
	Value      float64
	Binary     []byte
	Bool       bool
	Integer    int
	Values     []float64
}
