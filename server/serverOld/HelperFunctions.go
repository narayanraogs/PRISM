package server

import (
	"fmt"
	"prismServer/driver"
	"prismServer/utils"
	"strconv"
	"strings"
	"time"
)

func getSpectrumDump(c *client, sa string) (string, bool) {
	var dev driver.SA
	ok := dev.LoadDevice(sa)
	if !ok {
		return "Unable to Load Device [Database Error]", false
	}
	response := dev.GetSpectrumDump()
	if response.Success {
		return response.Result["SpectrumDump"].String, true
	}
	return response.ErrorMessage, false
}

func getTraceDump(c *client, sa string, noOfPoints int) (string, bool) {
	var dev driver.SA
	ok := dev.LoadDevice(sa)
	if !ok {
		return "Unable to Load Device [Database Error]", false
	}
	response := dev.GetTraceDump(noOfPoints)
	if !response.Success {
		return response.ErrorMessage, false
	}
	fileData := response.Result["TraceDump"].String
	response = dev.GetNoOfRowsToSkipInTrace()
	if !response.Success {
		return response.ErrorMessage, false
	}
	noOfRows := response.Result["NoOfRows"].Integer
	return utils.GetTracePlot(fileData, noOfRows)
}

func getScreenshot(c *client, vsa string, mode string) (string, bool) {
	var dev driver.VSA
	ok := dev.LoadDevice(vsa)
	if !ok {
		return "Unable to Load Device [Database Error]", false
	}
	response := dev.GetScreenshot(mode)
	if response.Success {
		return response.Result["Screenshot"].String, true
	}
	return response.ErrorMessage, false
}

func getPowerFromPM(c *client, pm string) (string, bool) {
	var dev driver.PM
	ok := dev.LoadDevice(pm)
	if !ok {
		return "Unable to Load Device [Database Error]", false
	}
	power := ""
	response := dev.GetPowerChannelA(true)
	if !response.Success {
		return response.ErrorMessage, false
	}

	power = power + fmt.Sprintf("%.2f", response.Result["Power"].Value) + ","
	time.Sleep(100 * time.Millisecond)

	response = dev.GetPowerChannelB(true)
	if !response.Success {
		return response.ErrorMessage, false
	}
	power = power + fmt.Sprintf("%.2f", response.Result["Power"].Value)
	return power, true
}

func setSpectrum(c *client, sa string, details []string) (string, bool) {
	var dev driver.SA
	freq, err := strconv.ParseFloat(details[0], 64)
	if err != nil {
		return "Frequency not a floating point number", false
	}
	span, err := strconv.ParseFloat(details[1], 64)
	if err != nil {
		return "Span not a floating point number", false
	}
	rbw, err := strconv.ParseFloat(details[2], 64)
	if err != nil {
		return "RBW not a floating point number", false
	}
	vbw, err := strconv.ParseFloat(details[3], 64)
	if err != nil {
		return "VBW not a floating point number", false
	}
	ref, err := strconv.ParseFloat(details[5], 64)
	if err != nil {
		return "Reference Level not a floating point number", false
	}
	ok := dev.LoadDevice(sa)
	if !ok {
		return "Unable to Load Device [Database Error]", false
	}
	resp := dev.SetSpectrum(freq, span, rbw, vbw)
	if !resp.Success {
		return "Unable to set spectrum", false
	}
	if strings.EqualFold(details[4], "Auto") {
		dev.SetReferenceNominal()
	} else {
		dev.SetReferenceLevel(ref)
	}
	if strings.EqualFold(details[6], "ClearWrite") {
		dev.SetNormalMode()
	}
	if strings.EqualFold(details[6], "MaxHold") {
		dev.SetMaxHold()
	}
	return "", true
}

func getSpectrum(c *client, sa string) ([]string, bool) {
	var dev driver.SA
	ok := dev.LoadDevice(sa)
	if !ok {
		return nil, false
	}
	resp := dev.GetSpectrum()
	if !resp.Success {
		return nil, false
	}
	var tbr = make([]string, 0)
	tbr = append(tbr, fmt.Sprintf("%.2f", resp.Result["CenterFrequency"].Value))
	tbr = append(tbr, fmt.Sprintf("%.2f", resp.Result["Span"].Value))
	tbr = append(tbr, fmt.Sprintf("%.2f", resp.Result["RBW"].Value))
	tbr = append(tbr, fmt.Sprintf("%.2f", resp.Result["VBW"].Value))
	tbr = append(tbr, fmt.Sprintf("%.2f", resp.Result["ReferenceLevel"].Value))
	return tbr, true
}
