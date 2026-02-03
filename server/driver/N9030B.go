package driver

import (
	"bytes"
	"encoding/base64"
	"os"
	"prismServer/utils"
)

type n9030b struct {
	n9030
}

func (device *n9030b) getSpectrumDump() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setFullScreen", "setMonochromeBackground", "setScreenDump", "getScreenDump")
	arguments = append(arguments, "", "", "\"D:\\temp.png\"", "\"D:\\temp.png\"")
	replacements = append(replacements, "", "", "", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
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

	crop := utils.CropImage(0, 55, 0, 55, data)
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
