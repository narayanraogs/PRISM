package reports

import (
	"encoding/base64"
	"fmt"
	"prismServer/utils"
	"strconv"
	"strings"
	"time"
)

type Report struct {
	Header          ReportHeader
	Results         map[string]Result
	Order           []string
	Remarks         string
	PreReqTM        []ReportTM
	LogTM           []ReportTM
	TestInformation []ReportInfo
	Screenshots     []Images
	Filenames       []string
	BatchSize       int
	OK              bool
	Message         string
}

type ReportHeader struct {
	TestType     string
	TestCategory string
	Spacecraft   string
	Config       string
	Date         string
	Time         string
	TestPhase    string
}

type Result struct {
	Header []string
	Data   [][]DataCell
}

func getColumns(length int) string {
	var tbr = "auto,"
	for i := 1; i < length; i++ {
		tbr = tbr + "1fr,"
	}
	tbr = strings.TrimRight(tbr, ",")
	return tbr
}

func getHeaders(header []string) string {
	var tbr = ""
	for _, h := range header {
		tbr = tbr + "[ *" + h + "* ],"
	}
	tbr = strings.TrimRight(tbr, ",")
	return tbr
}

func (result *Result) getString() string {
	var builder strings.Builder
	builder.WriteString(` #table(
        columns: ( `)
	builder.WriteString(getColumns(len(result.Header)))
	builder.WriteString("),\n")
	builder.WriteString("rows: ")
	builder.WriteString(strconv.Itoa(len(result.Data)))
	builder.WriteString(",\n")
	builder.WriteString("table.header(")
	builder.WriteString(getHeaders(result.Header))
	builder.WriteString("),\n")
	for _, row := range result.Data {
		for _, cell := range row {
			builder.WriteString(cell.GetString())
		}
	}
	builder.WriteString(")")
	return builder.String()
}

type DataCell struct {
	Value   string
	Error   bool
	Success bool
	Warning bool
}

type ReportTM struct {
	Mnemonic string
	Value    string
}

type ReportInfo struct {
	Parameter string
	Value     string
}

type Images struct {
	FileData string
	Caption  string
}

func (report *Report) SetHeader(config string, testType string, testCategory string, testPhase string) time.Time {
	report.Header.Spacecraft = utils.GetSatelliteName()
	report.Header.Config = config
	report.Header.TestType = testType
	report.Header.TestCategory = testCategory
	now := time.Now()
	report.Header.Date = strings.ToUpper(now.Format("02-Jan-2006"))
	report.Header.Time = now.Format("15:04:05")
	report.Header.TestPhase = testPhase
	report.Screenshots = make([]Images, 0)
	report.Results = make(map[string]Result)
	report.TestInformation = make([]ReportInfo, 0)
	return now
}

func (report *Report) SetPreRequisiteTM(mnemonic []string, value []string) {
	report.PreReqTM = make([]ReportTM, 0)
	for i := 0; i < len(mnemonic); i++ {
		var tm ReportTM
		tm.Mnemonic = mnemonic[i]
		tm.Value = value[i]
		report.PreReqTM = append(report.PreReqTM, tm)
	}
}

func (report *Report) SetPostTestTM(mnemonic []string, value []string) {
	report.LogTM = make([]ReportTM, 0)
	for i := 0; i < len(mnemonic); i++ {
		var tm ReportTM
		tm.Mnemonic = mnemonic[i]
		tm.Value = value[i]
		report.LogTM = append(report.LogTM, tm)
	}
}

func (report *Report) SetTestInformation(params []string, value []string) {
	report.TestInformation = make([]ReportInfo, 0)
	for i := 0; i < len(params); i++ {
		var parameter ReportInfo
		parameter.Parameter = params[i]
		parameter.Value = value[i]
		report.TestInformation = append(report.TestInformation, parameter)
	}
}

func (report *Report) AddTestInformation(param string, value string) {
	var parameter ReportInfo
	parameter.Parameter = param
	parameter.Value = value
	report.TestInformation = append(report.TestInformation, parameter)
}

func (report *Report) SetRemarks(remark string) {
	report.Remarks = remark
}

func (report *Report) SetScreenshots(images []Images) {
	report.Screenshots = make([]Images, 0)
	report.Screenshots = append(report.Screenshots, images...)
}

func (report *Report) SetResults(reportName string, header []string, data [][]DataCell) {
	var result Result
	result.Header = make([]string, 0)
	result.Header = append(result.Header, header...)
	result.Data = make([][]DataCell, 0)
	for i := 0; i < len(data); i++ {
		var row = make([]DataCell, 0)
		row = append(row, data[i]...)
		result.Data = append(result.Data, row)
	}
	report.Results[reportName] = result
}

func (report *Report) SetOrder(order []string) {
	report.Order = make([]string, 0)
	report.Order = append(report.Order, order...)
}

func GetDataCell(value string) DataCell {
	var data DataCell
	data.Value = value
	data.Success = false
	data.Error = false
	data.Warning = false
	return data
}

func (data *DataCell) SetError() {
	data.Error = true
}

func (data *DataCell) SetSuccess() {
	data.Success = true
}

func (data *DataCell) SetWarning() {
	data.Warning = true
}

func (data *DataCell) GetString() string {
	var tbr string
	if data.Success {
		tbr = "[#text(blue)"
	} else if data.Warning {
		tbr = "[#text(fuchsia)"
	} else if data.Error {
		tbr = "[#text(red)"
	} else {
		tbr = "[#text(black)"
	}
	tbr = tbr + "[" + data.Value + "]],\n"
	return tbr
}

func (header *ReportHeader) getString() string {
	var builder strings.Builder
	builder.WriteString(`
	#set page(
	margin: (
	top: 4cm,
	x: 1.5cm,
	bottom:2cm
	),

	header:
	box(
	stroke: black,
	table(
	stroke: none,
	columns: (auto,1fr,auto),
	align:center,
	rows: 2,
	table.cell(align: left, [`)
	builder.WriteString(header.Spacecraft)
	builder.WriteString("]),\n")
	builder.WriteString("text(purple,14pt)[")
	builder.WriteString(header.TestType)
	if strings.TrimSpace(header.TestCategory) != "" {
		builder.WriteString(": _")
		builder.WriteString(header.TestCategory)
		builder.WriteString("_")
	}
	builder.WriteString("],\n")
	builder.WriteString("table.cell(align: right, [")
	builder.WriteString(header.Config)
	builder.WriteString("]),\n")
	builder.WriteString("table.cell(align: left, [")
	builder.WriteString(header.TestPhase)
	builder.WriteString("]),\n")
	builder.WriteString("[")
	builder.WriteString(header.Date)
	builder.WriteString(" ")
	builder.WriteString(header.Time)
	builder.WriteString("],\n")
	builder.WriteString(`table.cell(align: right, [Page #context(counter(page).display("1/1",both:true,))]),
		)
	),
	footer: table(
			stroke: none,
			columns: (1fr, auto, 1fr),
			align: (left, center, right),
			rows: 1,
			table.hline(),
			[PRISM],
	  text(size:8pt)[_Committed to total Quality and Zero defect in Space Systems and Services through Continual Improvement_],
			[SCG],
		)
	)
`)
	builder.WriteString("\n")
	return builder.String()
}

func (report *Report) GetPreReqTMTable() string {
	var content strings.Builder
	content.WriteString(`
	#text(purple,14pt)[*Pre-Requisite TM*]
	#linebreak()
	#table(
	columns: (1fr, 1fr, 1fr, 1fr),
	table.header(
	[*TM Mnemonic*],[*Value*],[*TM Mnemonic*],[*Value*]
	),
	`)
	noOfRows := len(report.PreReqTM) / 2
	if len(report.PreReqTM)%2 == 1 {
		noOfRows = noOfRows + 1
		report.PreReqTM = append(report.PreReqTM, ReportTM{})
	}
	if noOfRows == 0 {
		noOfRows = 4
	}
	rows := "rows: " + strconv.Itoa(noOfRows) + ",\n"
	content.WriteString(rows)
	content.WriteString(getTableFromTM(report.PreReqTM))
	content.WriteString("\n")

	return content.String()
}

func (report *Report) GetTestInformationTable() string {
	if len(report.TestInformation) == 0 {
		return "\n"
	}
	var content strings.Builder
	content.WriteString(`
	#text(purple,14pt)[*Information*]
	#linebreak()
	#table(
	columns: (1fr, 1fr, 1fr, 1fr),
	table.header(
	[*Parameter*],[*Value*],[*Parameter*],[*Value*]
	),
	`)
	noOfRows := len(report.TestInformation) / 2
	if len(report.TestInformation)%2 == 1 {
		noOfRows = noOfRows + 1
		report.TestInformation = append(report.TestInformation, ReportInfo{})
	}
	if noOfRows == 0 {
		noOfRows = 4
	}
	rows := "rows: " + strconv.Itoa(noOfRows) + ",\n"
	content.WriteString(rows)
	content.WriteString(getTableFromTestInformation(report.TestInformation))
	content.WriteString("\n")

	return content.String()
}

func (report *Report) GetLogTMTable() string {
	var content strings.Builder
	content.WriteString(`
	#text(purple,14pt)[*Log TM*]
	#linebreak()
	#table(
	columns: (1fr, 1fr, 1fr, 1fr),
	table.header(
	[*TM Mnemonic*],[*Value*],[*TM Mnemonic*],[*Value*]
	),
	`)
	noOfRows := len(report.LogTM) / 2
	if len(report.LogTM)%2 == 1 {
		noOfRows = noOfRows + 1
		report.LogTM = append(report.LogTM, ReportTM{})
	}
	if noOfRows == 0 {
		noOfRows = 4
	}
	rows := "rows: " + strconv.Itoa(noOfRows) + ",\n"
	content.WriteString(rows)
	content.WriteString(getTableFromTM(report.LogTM))
	content.WriteString("\n")

	return content.String()
}

func (report *Report) GetHeader() string {
	return report.Header.getString()
}

func (report *Report) GetResults() string {
	var content strings.Builder
	for _, reportName := range report.Order {
		result := report.Results[reportName]
		if len(result.Header) == 0 || len(result.Data) == 0 {
			continue
		}
		content.WriteString("#text(purple,14pt)[*")
		content.WriteString(reportName)
		content.WriteString("*]\n#linebreak()\n")
		content.WriteString(result.getString())
	}
	content.WriteString("\n")
	return content.String()
}

func (report *Report) GetRemarks() string {
	var content strings.Builder
	content.WriteString("#text(purple,14pt)[*Remark:*] ")
	content.WriteString(report.Remarks)
	content.WriteString("\n#linebreak()\n")
	return content.String()
}

func (report *Report) GetScreenshots() string {
	if len(report.Screenshots) == 0 {
		return ""
	}
	var content strings.Builder
	content.WriteString("#text(purple,14pt)[*Screenshots*]\n")
	for _, image := range report.Screenshots {
		bytes, _ := base64.StdEncoding.DecodeString(image.FileData)
		content.WriteString(`
		#page(flipped: true)[
		#grid(
			columns: (1fr),
			align: center,
		`)
		content.WriteString("image(\n")
		content.WriteString("bytes((")
		content.WriteString(getStringOfBytes(bytes))
		content.WriteString(")),\nwidth: 90%,\nfit:\"contain\"\n")
		content.WriteString("),\n[#linebreak()],\n[")
		content.WriteString(image.Caption)
		content.WriteString("]\n")
		content.WriteString(")\n]\n")
	}
	return content.String()
}

func getStringOfBytes(array []byte) string {
	var content strings.Builder
	for _, b := range array {
		content.WriteString(fmt.Sprintf("%d, ", b))
	}
	str := content.String()
	return strings.TrimRight(str, ", ")
}

func getTableFromTM(tm []ReportTM) string {
	var content strings.Builder
	for i := 0; i < len(tm); i = i + 2 {
		content.WriteString("[" + tm[i].Mnemonic + "],\n")
		content.WriteString("[" + tm[i].Value + "],\n")
		content.WriteString("[" + tm[i+1].Mnemonic + "],\n")
		content.WriteString("[" + tm[i+1].Value + "],\n")
	}

	content.WriteString("\n)")
	return content.String()
}

func getTableFromTestInformation(tm []ReportInfo) string {
	var content strings.Builder
	for i := 0; i < len(tm); i = i + 2 {
		content.WriteString("[" + tm[i].Parameter + "],\n")
		content.WriteString("[" + tm[i].Value + "],\n")
		content.WriteString("[" + tm[i+1].Parameter + "],\n")
		content.WriteString("[" + tm[i+1].Value + "],\n")
	}

	content.WriteString("\n)")
	return content.String()
}
