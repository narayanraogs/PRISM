package server

import "prismServer/utilities"

type ServerStatus struct {
	SatelliteName string
	TestPhase     string
	MemoryUsed    float64
	CPUUsed       float64
}

type RFUplinkRequest struct {
	TSMSelected string
}

type RFUplinkMetaData struct {
	UplinkConfigs           []string
	UplinkConfigInformation map[string]UplinkConfigInformation
	TSMs                    []string
	AllConfigs              []string
	ConfigPathInformation   map[string][]ConfigPathInformation
	OK                      bool
	Message                 string
}

type LinkStatus struct {
	RemoveConfigs []string
	TSMConnected  bool
	SwitchStatus  []string
	Attn1Value    float64
	Attn2Value    float64
	OK            bool
	Message       string
}

type UplinkConfigInformation struct {
	PowerAtSC float64
	SALoss    float64
	SCLoss    float64
}

type ConfigPathInformation struct {
	Path     string
	Mnemonic string
}

type TSMRouteRequest struct {
	TSMSelected string
	Mnemonic    string
}

type TSMSetAttn struct {
	TSMSelected string
	AttnNo      int
	AttnValue   float64
}

type Ack struct {
	OK      bool
	Message string
}

type AllTests struct {
	Categories     []string
	Configurations map[string][]string
	Losses         map[string]string
	Tests          map[string][]TestDescription
	OK             bool
	Message        string
}

type TestDescription struct {
	TestName        string
	TestCategory    string
	Configuration   string   `json:",omitempty"`
	Remark          string   `json:",omitempty"`
	ExtraParameters []string `json:",omitempty"`
}

type StartTestsRequest struct {
	Tests []TestDescription
}

type ClientInput struct {
	Parameters []string
}

type StabilityMetadata struct {
	InstrumentTypes []string
	Instruments     map[string][]string
	Parameters      map[string][]string
	Profiles        []SpectrumProfile
	PLConfigs       []string
	PulseProfiles   []string
	OK              bool
	Message         string
}

type SpectrumProfile struct {
	ProfileName     string
	CenterFrequency float64
	Span            float64
	RBW             float64
	VBW             float64
}

type SpectrumDumpMetadata struct {
	SpectrumDumpMode   []string
	Instruments        map[string][]string
	SpectrumProfiles   []SpectrumProfile
	ScreenshotProfiles []string
	OK                 bool
	Message            string
}

type SetSpectrumRequest struct {
	SA              string
	CenterFrequency float64
	Span            float64
	RBW             float64
	VBW             float64
	AutoReference   bool
	ReferenceLevel  float64
	Mode            string
}

type ReadSpectrumRequest struct {
	SA string
}

type DumpTraceRequest struct {
	SA          string
	TracePoints int
}

type ReadSpectrumResponse struct {
	CenterFrequency float64
	Span            float64
	RBW             float64
	VBW             float64
	ReferenceLevel  float64
	OK              bool
	Message         string
}

type SaveSpectrumRequest struct {
	Spectrum string
	Remark   string
}

type DumpScreenshotRequest struct {
	VSA  string
	Mode string
}

type MonitorMetadata struct {
	InstrumentTypes []string
	Instruments     map[string][]string
	OK              bool
	Message         string
}

type MonitorRequest struct {
	InstrumentType string
	Instrument     string
}

type MonitorResponse struct {
	Image                string
	PMChannelA           float64
	PMChannelB           float64
	PPMChannelAPeakPower float64
	PPMChannelBPeakPower float64
	PPMChannelAAvgPower  float64
	PPMChannelBAvgPower  float64
	OK                   bool
	Message              string
}

type TVACCableLossMetadata struct {
	Frequencies    []float64 `json:"frequencies"`
	DeviceProfiles []string  `json:"deviceProfiles"`
	ExistingCables []string  `json:"existingCables"`
	IsPMZeroed     bool      `json:"isPmZeroed"`
	OK             bool      `json:"ok"`
	Message        string    `json:"message"`
}

type TVACCableLossRequest struct {
	Action        string `json:"action"` // "ZeroPM" | "Measure"
	DeviceProfile string `json:"deviceProfile"`
	Channel       string `json:"channel"` // "A" | "B"
	CableName     string `json:"cableName"`
	CycleName     string `json:"cycleName"`
	Phase         string `json:"phase"`      // "Ambient" | "Hot" | "Cold"
	RequestRef    bool   `json:"requestRef"` // Manual baseline override
}

type TVACCableLossResponse struct {
	LatestRecord utilities.TVACCableLossRecord   `json:"latestRecord"`
	History      []utilities.TVACCableLossRecord `json:"history"`
	IsPMZeroed   bool                            `json:"isPmZeroed"`
	OK           bool                            `json:"ok"`
	Message      string                          `json:"message"`
}
