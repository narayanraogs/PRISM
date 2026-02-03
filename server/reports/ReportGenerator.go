package reports

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"prismServer/utils"
	"strings"
)

func GenerateResult(report Report) (string, error) {
	var builder strings.Builder
	builder.WriteString(report.GetHeader())
	builder.WriteString(report.GetResults())
	builder.WriteString(report.GetRemarks())
	builder.WriteString(report.GetPreReqTMTable())
	builder.WriteString(report.GetLogTMTable())
	builder.WriteString(report.GetTestInformationTable())
	builder.WriteString(report.GetScreenshots())

	withoutExtension := filepath.Join(utils.Config.BaseFolder, "temp", "report")
	withoutExtension = utils.GetTimeStampedFileName(withoutExtension)
	typstName := withoutExtension + ".typ"
	err := os.WriteFile(typstName, []byte(builder.String()), 0666)
	if err != nil {
		return "", err
	}
	pdfName := withoutExtension + ".pdf"
	err = compile(typstName, pdfName)
	return pdfName, err
}

func compile(typstname string, pdfname string) error {
	cmd := "typst"

	options := make([]string, 0)
	options = append(options, "compile")
	options = append(options, typstname)
	options = append(options, pdfname)
	command := exec.Command(cmd, options...)
	log, err := command.CombinedOutput()
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(log)
		return err
	}
	return nil
}
