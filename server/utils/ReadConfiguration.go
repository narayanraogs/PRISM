package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type simulator struct {
	SA  bool
	PM  bool
	GTx bool
	TSM bool
	SG  bool
	VSA bool
	PPM bool
}

type db struct {
	DBPath        string
	ResultsDBPath string
}

type scc struct {
	IP      string
	PortNo  int
	Path    string
	Timeout int
}

type tsm struct {
	NoOfDrivers         int
	MaxProgrammableAttn float64
	StepSize            float64
}

type testRelated struct {
	SweepsForSustainedLock int
	MinProgAttnValueForSG  float64
	ZerodBmTolerence       float64
	RxPowerLevelTolerance  float64
	ChipDopplerRate        float64
	MasterFrameTimeSecs    int
}

type mainConfiguration struct {
	SatelliteName      string
	UMACSSatelliteName string
	BaseFolder         string
	Database           db
	TMServer           scc
	ProcedureServer    scc
	TSM                tsm
	TestRelated        testRelated
	Simulator          simulator `json:"Simulator,omitempty"`
}

var Config mainConfiguration

func WriteConfig() {
	Config = mainConfiguration{
		SatelliteName: "Oceansat-3A",
		BaseFolder:    "~/prism",
		Database: db{
			DBPath:        "~/prism/db/prism.db",
			ResultsDBPath: "~/prism/db/prism-results.db",
		},
		TMServer: scc{
			IP:      "172.20.10.1",
			PortNo:  9050,
			Path:    "ws",
			Timeout: 32,
		},
		ProcedureServer: scc{
			IP:      "172.20.10.1",
			PortNo:  10050,
			Path:    "proc",
			Timeout: 5,
		},
		TSM: tsm{
			NoOfDrivers:         2,
			MaxProgrammableAttn: 63.5,
			StepSize:            0.125,
		},
		TestRelated: testRelated{
			SweepsForSustainedLock: 4,
			MinProgAttnValueForSG:  6,
			ZerodBmTolerence:       0.5,
			RxPowerLevelTolerance:  0.5,
			ChipDopplerRate:        0.001461161387,
			MasterFrameTimeSecs:    16,
		},
	}
	data, err := json.MarshalIndent(Config, "", " ")
	if err != nil {
		fmt.Println("Cannot convert struct to JSON")
		return
	}
	filePath := "~/prism/config/config.json"
	filePath, _ = expandTilde(filePath)
	err = os.WriteFile(filePath, data, 0666)
	if err != nil {
		fmt.Println("Cannot wirte file", filePath)
	} else {
		fmt.Println("Config File Created")
	}
}

func ReadConfiguration(path string) bool {
	path, _ = expandTilde(path)
	fmt.Println("Reading Configuration from", path)
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Cannot read Configuration from", path)
		fmt.Println("Error is", err.Error())
		return false
	}
	err = json.Unmarshal(data, &Config)
	if err != nil {
		fmt.Println("Cannot read Configuration from", path)
		fmt.Println("Error is", err.Error())
		return false
	}
	Config.BaseFolder, _ = expandTilde(Config.BaseFolder)
	Config.Database.DBPath, _ = expandTilde(Config.Database.DBPath)
	Config.Database.ResultsDBPath, _ = expandTilde(Config.Database.ResultsDBPath)
	fmt.Println("Configuration Read")
	return true
}

func expandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return usr.HomeDir, nil
	}

	return filepath.Join(usr.HomeDir, path[2:]), nil
}
