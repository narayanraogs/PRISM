package server

import (
	"prismServer/tne"
	"prismServer/utilities"
)

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
	PPMChannels     []string
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

type CableLossMetadata struct {
	OK             bool     `json:"ok"`
	Message        string   `json:"message"`
	Frequencies    []string `json:"frequencies"`    // e.g. ["L-Band;1.5", "S-Band;2.2"]
	DeviceProfiles []string `json:"deviceProfiles"` // From LossMeasurementFrequencies/DeviceProfiles
	ExistingCables []string `json:"existingCables"` // Distinct cable names from cable_loss table
	IsPMZeroed     bool     `json:"isPmZeroed"`     // Check if PM ref exists in resultsDB
}

type CableLossHistoryResponse struct {
	OK      bool                  `json:"ok"`
	Message string                `json:"message"`
	History []tne.CableLossRecord `json:"history"`
}

type CableLossRequest struct {
	Action              string   `json:"action"` // "pmreference" or "measure"
	DeviceProfile       string   `json:"deviceProfile"`
	Channel             string   `json:"channel"` // "A" or "B"
	CableName           string   `json:"cableName"`
	CableLength         float64  `json:"cableLength"`
	SelectedFrequencies []string `json:"selectedFrequencies"` // Names only, e.g. ["L-Band", "Ku1"]
}

type MeasurementStatus struct {
	Message   string `json:"message"`
	Completed bool   `json:"completed"`
	Success   bool   `json:"success"`
	Error     bool   `json:"error"`
}

type AttnRange struct {
	Max      float64
	Min      float64
	StepSize float64
}

type AttnMetaData struct {
	DeviceProfile    []string
	Receiver         []string
	SprectrumProfile []string
	TSMConfig        []string
	GTxComponents    []string
	AttnRanges       map[string]AttnRange
	OK               bool
	Message          string
}

type AttnRequest struct {
	Type            string
	DeviceProfile   string
	Receiver        string
	SpectrumProfile string
	TSMConfig       string
	Component       string
	Min             float64
	Max             float64
	Step            float64
}

type AttnProgressResponse struct {
	MeasurementStatus tne.AttnMeasurementStatus `json:"measurementStatus,omitempty"`
	Deviations        []tne.CorrectedDeviation  `json:"deviations,omitempty"`
	OK                bool                      `json:"ok"`
	Message           string                    `json:"message"`
}

type DatabaseMetadata struct {
	TestPhases []string
	OK         bool
	Message    string
}

type ConfigsForLossRequest struct {
	TestPhase string
}

type ConfigsForLossResponse struct {
	Configs []string
	OK      bool
	Message string
}

type LossProfileRequest struct {
	TestPhase string
	Config    string
}

type LossProfileResponse struct {
	Profile string
	OK      bool
	Message string
}

type SaveLossProfileRequest struct {
	TestPhase string
	Config    string
	Profile   string
}

type SelectTestPhaseRequest struct {
	TestPhase string
}

type AddNewTestPhaseRequest struct {
	NewPhase string
	CopyFrom string
}

// StabilityRequest is sent by the client immediately after connection
type StabilityRequest struct {
	ProfileName string                        `json:"ProfileName"`
	Parameters  []StabilityParameterSelection `json:"Parameters"`
}

// StabilityParameterSelection matches the configuration from the client
type StabilityParameterSelection struct {
	Description    string                 `json:"description"`
	InstrumentType string                 `json:"instrumentType"`
	Instrument     string                 `json:"instrument"`
	Parameter      string                 `json:"parameter"`
	Details        string                 `json:"details"`
	ExtraDetails   map[string]interface{} `json:"extraDetails"`
}

// StabilityUpdate is an individual data point sent from server to client

// StabilityResponse is the wrapper sent to the client periodically
type StabilityResponse struct {
	Updates []utilities.StabilityUpdate `json:"Updates"`
	OK      bool                        `json:"OK"`
	Message string                      `json:"Message"`
}

// StabilityAction is for control messages (like "abort")
type StabilityAction struct {
	Action string `json:"Action"`
}
