package measurements

import (
	"prismServer/database"
	"prismServer/driver"
	"prismServer/utils"
	"strconv"
	"strings"
)

func setTSMPath(tsm driver.TSM, path string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return tsm.SetDriverStatus(path)
	}
}

func GetAllActiveUplinkConfigs(tsmName string) []string {
	var tbr = make([]string, 0)
	var tsm driver.TSM
	ok := tsm.LoadDevice(tsmName)
	if !ok {
		return tbr
	}
	resp := tsm.GetDriverPath()
	if !resp.Success {
		return tbr
	}
	currentStatus, err := getStatusMap(resp.Result["DriverPath"].String)
	if err != nil {
		return tbr
	}
	tsmPaths, ok := database.GetTSMUplinkPathForAllConfigs()
	if !ok {
		return tbr
	}
	for k, v := range tsmPaths {
		if strings.TrimSpace(v) == "" {
			continue
		}
		pathMap, err := getStatusMap(v)
		if err != nil {
			continue
		}
		if isPathSubset(currentStatus, pathMap) {
			tbr = append(tbr, k)
		}
	}
	return tbr
}

func getStatusMap(status string) (map[int]map[int]bool, error) {
	var statusMap = make(map[int]map[int]bool)
	drSts := strings.Split(status, "!")
	var indStatus bool
	for _, d := range drSts {
		if strings.TrimSpace(d) == "" {
			continue
		}
		driverNo := d[1:2]
		drNoInt, err := strconv.Atoi(driverNo)
		if err != nil {
			return statusMap, err
		}
		statusMap[drNoInt] = make(map[int]bool)
		d = d[2:]
		for i := 0; i < len(d); i++ {
			str := d[i : i+1]
			if strings.EqualFold(str, "A") {
				indStatus = true
				continue
			}
			if strings.EqualFold(str, "B") {
				indStatus = false
				continue
			}
			switchNo, err := strconv.Atoi(str)
			if err != nil {
				return statusMap, err
			}
			statusMap[drNoInt][switchNo] = indStatus
		}
	}
	return statusMap, nil
}

func isPathSubset(overallPath map[int]map[int]bool, smallerPath map[int]map[int]bool) bool {
	var subset = true
	for dr := range smallerPath {
		biggerPath := overallPath[dr]
		subsetPath := smallerPath[dr]
		for sw := range subsetPath {
			match := biggerPath[sw] == subsetPath[sw]
			subset = subset && match
		}
	}
	return subset
}
