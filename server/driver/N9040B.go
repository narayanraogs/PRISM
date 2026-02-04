package driver

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"prismServer/utils"
	"strconv"
	"strings"
	"time"
)

type n9040b struct {
	n9030
}

func (device *n9040b) setPulseMode() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, "setReset", "setPulseMeasurementMode")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9040b) setSpectrumParameters(centerFrequency float64, span float64, rbw float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	centerFreqString := fmt.Sprintf("%.2f", centerFrequency)
	spanString := fmt.Sprintf("%.2f", span)
	rbwString := fmt.Sprintf("%.2f", rbw)

	mnemonics = append(mnemonics, "setCenterFrequencyVSAMode", "setSpanVSAMode", "setRBWVSAMode")
	arguments = append(arguments, centerFreqString, spanString, rbwString)
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9040b) setPulseParameters(acqtime float64, YTop float64, pdiv float64, analLength float64, reflevel float64,
	hystlevel float64, points int32, bufferlength int32) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 14)

	mnemonics = append(mnemonics, "setPulseAcquisitionTime", "setYScaleTop", "setDivisionYAxis",
		"setPulseAnalysisLength", "setPulseReferencePoint", "setPulseReferenceLevel", "setPulseHysteresisLevel",
		"setPulseModulation", "setDisplayPoints", "setBufferLength", "setMeasurementContinuous", "setDisplayLayout",
		"setTraceName", "setRange")
	replacements[12] = "4"
	replacements[13] = "1"
	arguments = append(arguments, fmt.Sprintf("%.6f", acqtime),
		fmt.Sprintf("%.2f", YTop),
		fmt.Sprintf("%.2f", pdiv),
		fmt.Sprintf("%.6f", analLength),
		"\"RelativeToBase\"",
		fmt.Sprintf("%.2f", reflevel),
		fmt.Sprintf("%.2f", hystlevel),
		"",
		fmt.Sprintf("%d", points),
		fmt.Sprintf("%d", bufferlength),
		"ON",
		"3,2", "\"Pulse Inst Phase Meas Time1\"",
		"-2")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9040b) startMeasurement() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, "setMeasurementStart")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9040b) stopMeasurement() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setMeasurementStop", "setTraceName")
	arguments = append(arguments, "", "\"Pulse Cumulative Results Table1\"")
	replacements = append(replacements, "", "6")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9040b) setSpectrogramMode(mode string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setSpectrogramMode")
	arguments = append(arguments, mode)
	replacements = append(replacements, "2")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9040b) getPulseParameters() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "getTableHeader", "getTableData")
	replacements = append(replacements, "6", "6")
	arguments = append(arguments, "", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)
	values := device.communicate(cmds, "Alternate")
	if values == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	header := values[0]
	headerCols := strings.Split(header, ",")
	temp, _ := base64.StdEncoding.DecodeString(values[1])
	tempStr := string(temp)
	cols := strings.Split(tempStr, ";")
	if len(headerCols) != len(cols) {

		return getErrorResponse("Column length is not same")
	}
	var data = make([][]string, 0)
	for _, col := range cols {
		column := strings.Split(col, ",")
		data = append(data, column)
	}

	data = utils.Transpose(data)

	var fileData strings.Builder
	headerString := strings.Join(headerCols, ",")
	fileData.WriteString(headerString)
	fileData.WriteString("\n")

	for _, line := range data {
		lineString := strings.Join(line, ",")
		fileData.WriteString(lineString)
		fileData.WriteString("\n")
	}

	err := os.WriteFile(utils.Config.BaseFolder+"/temp/temp.csv", []byte(fileData.String()), os.ModePerm)
	if err != nil {
		return utils.CommandResponse{}
	}

	return getSuccessResponse()
}

func (device *n9040b) getPulseAveragePower() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "getTableDataAsString", "getTableDataAsString")
	arguments = append(arguments, "\"Pulse_Pulse_Number1\"", "\"Pulse_Top_Level1\"")
	replacements = append(replacements, "6", "6")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with VSA")
	}

	pulseNumbersStr := strings.Split(retVal[0], ",")
	averagePowersStr := strings.Split(retVal[1], ",")
	pulseNumbers := utils.GetFloatArray(pulseNumbersStr)
	averagePowers := utils.GetFloatArray(averagePowersStr)

	ret := getSuccessResponse()
	ret.Result["PulseNo"] = utils.CommandResult{
		ResultType: "Values",
		Values:     pulseNumbers,
	}
	ret.Result["PulseAvgPower"] = utils.CommandResult{
		ResultType: "Values",
		Values:     averagePowers,
	}
	ret.Result["TotalNoOfPulses"] = utils.CommandResult{
		ResultType: "Value",
		Value:      float64(len(pulseNumbersStr)),
	}
	return ret
}

func (device *n9040b) waitTillFirstPulse() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "getTableDataAsString")
	arguments = append(arguments, "\"Pulse_Pulse_Number1\"")
	replacements = append(replacements, "6")
	var cmds = device.getCommands(mnemonics, arguments, replacements)
	for {
		retVal := device.communicate(cmds, "Alternate")
		if retVal == nil {
			return getErrorResponse("Cannot Communicate with SA")
		}
		if len(retVal[0]) == 0 {
			continue
		}
		return getSuccessResponse()
	}
}

func (device *n9040b) getScreenshot(mode string) utils.CommandResponse {
	var restore bool = false
	switch strings.ToLower(mode) {
	case "magnitude":
		device.setSelectedTrace(1)
		restore = true
	case "spectrogram":
		device.setSelectedTrace(2)
		restore = true
	case "pulse magnitude":
		device.setSelectedTrace(3)
		restore = true
	case "pulse phase":
		device.setSelectedTrace(4)
		restore = true
	case "pulse frequency":
		device.setSelectedTrace(5)
		restore = true
	}
	time.Sleep(2 * time.Second)
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setScreenDumpVSA")
	arguments = append(arguments, "\"D:\\temp.png\", \"Png\"")
	replacements = append(replacements, "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with VSA")
	}

	mnemonics = make([]string, 0)
	arguments = make([]string, 0)
	replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getScreenDump")
	arguments = append(arguments, "\"D:\\temp.png\"")
	//arguments = append(arguments, "\"D:\\temp.png\", \"Png\"","\"D:\\temp.png\"")
	replacements = append(replacements, "")
	cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal = device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with VSA")
	}
	if restore {
		device.restoreTrace()
	}
	if len(retVal) <= 0 {
		return getErrorResponse("Unable to get screenshot as length is zero")
	}

	data, _ := base64.StdEncoding.DecodeString(retVal[0])

	sep := []byte{0x50, 0x4E, 0x47}
	var index = bytes.Index(data, sep)
	if index == -1 {
		return getErrorResponse("Unable to get screenshot as byte index is -1")
	}
	index = index - 1
	data = data[index:]

	crop := utils.CropImage(10, 60, 0, 25, data)
	if crop == nil {
		return getErrorResponse("Unable to crop image")
	}

	filename := utils.GetTimeStampedFileName("screenshot")
	filename = utils.Config.BaseFolder + "/screenshots/" + filename + ".png"
	_ = os.WriteFile(filename, crop, os.ModePerm)

	var encodedImage = base64.StdEncoding.EncodeToString(crop)
	ret := getSuccessResponse()
	ret.Result["Screenshot"] = utils.CommandResult{
		ResultType: "Image",
		String:     encodedImage,
	}
	return ret
}
func (device *n9040b) setSelectedTrace(number int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setDisplayLayout")
	arguments = append(arguments, "1,1")
	replacements = append(replacements, "")
	var cmds1 = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds1, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	mnemonics = make([]string, 0)
	arguments = make([]string, 0)
	replacements = make([]string, 0)

	for i := 1; i < 7; i++ {
		mnemonics = append(mnemonics, "setTraceVisible")
		replacements = append(replacements, strconv.Itoa(i))
		arguments = append(arguments, "0")
	}
	mnemonics = append(mnemonics, "setTraceVisible")
	replacements = append(replacements, "7")
	arguments = append(arguments, "1")

	var cmds2 = device.getCommands(mnemonics, arguments, replacements)

	retVal = device.communicate(cmds2, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}
func (device *n9040b) restoreTrace() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setDisplayLayout")
	arguments = append(arguments, "3,2")
	replacements = append(replacements, "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	mnemonics = make([]string, 0)
	arguments = make([]string, 0)
	replacements = make([]string, 0)
	for i := 1; i < 7; i++ {
		mnemonics = append(mnemonics, "setTraceVisible")
		replacements = append(replacements, strconv.Itoa(i))
		arguments = append(arguments, "1")
	}
	var cmds2 = device.getCommands(mnemonics, arguments, replacements)

	retVal = device.communicate(cmds2, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9040b) getSpectrumDumpSA() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setFullScreen", "setScreenDump", "getScreenDump")
	arguments = append(arguments, "", "\"D:\\temp.png\"", "\"D:\\temp.png\"")
	replacements = append(replacements, "", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Alternate")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	data, _ := base64.StdEncoding.DecodeString(retVal[0])
	sep := []byte{0x50, 0x4E, 0x47}
	var index = bytes.Index(data, sep)
	if index == -1 {
		return getErrorResponse("Sprectum Dump is Not PNG")
	}
	index = index - 1
	data = data[index:]

	crop := utils.CropImage(0, 45, 0, 45, data)
	if crop == nil {
		return getErrorResponse("Unable to Crop Image")
	}

	filename := utils.GetTimeStampedFileName("screenshot")
	filename = utils.Config.BaseFolder + "/screenshots/" + filename + ".png"
	_ = os.WriteFile(filename, crop, os.ModePerm)

	var encodedImage = base64.StdEncoding.EncodeToString(crop)
	ret := getSuccessResponse()
	ret.Result["SpectrumDump"] = utils.CommandResult{
		ResultType: "Image",
		String:     encodedImage,
	}
	return ret
}

func (device *n9040b) getTraceDumpSA(points int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setNoOfTracePoints", "setSaveTrace", "getScreenDump")
	arguments = append(arguments, strconv.Itoa(points), "TRACE1, 'D:\\VALUES.csv'", "'D:\\VALUES.csv'")
	replacements = append(replacements, "", "", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}

	data, _ := base64.StdEncoding.DecodeString(retVal[0])
	filename := utils.GetTimeStampedFileName("trace")
	filename = utils.Config.BaseFolder + "/screenshots/" + filename + ".csv"
	_ = os.WriteFile(filename, data, os.ModePerm)

	fileData := string(data)
	ret := getSuccessResponse()
	ret.Result["TraceDump"] = utils.CommandResult{
		ResultType: "String",
		String:     fileData,
	}
	return ret
}
func (device *n9040b) startVSA() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, "getSelectedInstrument")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	//if !strings.EqualFold(retVal[0], "VSA89601") {
	if strings.EqualFold(retVal[0], "SA") {
		mnemonics = make([]string, 0)
		mnemonics = append(mnemonics, "setVSAMode")
		var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

		retVal = device.communicate(cmds, "Control")
		time.Sleep(30 * time.Second)
	}
	return getSuccessResponse()
}

func (device *n9040b) startSA() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, "getSelectedInstrument")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	if !strings.EqualFold(retVal[0], "SA") {
		mnemonics = make([]string, 0)
		mnemonics = append(mnemonics, "setSAMode")
		var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))
		retVal = device.communicate(cmds, "Control")
		time.Sleep(30 * time.Second)
	}
	return getSuccessResponse()
}

func (device *n9040b) checkSA() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, "getSelectedInstrument")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	if !strings.EqualFold(retVal[0], "SA") {
		return getErrorResponse("SA not selected")
	}
	return getSuccessResponse()
}

func (device *n9040b) checkVSA() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, "getSelectedInstrument")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with VSA")
	}
	if !strings.EqualFold(retVal[0], "VSA89601") {
		device.startVSA()
	}

	return getSuccessResponse()
}

func (device *n9040b) waitForSweeps(noOfSweeps int) utils.CommandResponse {
	var sweepTime float64
	temp := device.getSweepTime()
	sweepTime = temp.Result["SweepTime"].Value
	sleepTime := sweepTime * float64(noOfSweeps) * 1000
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	return getSuccessResponse()
}
