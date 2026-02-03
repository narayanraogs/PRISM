package measurements

import (
	"fmt"
	"os"
	"path/filepath"
	"prismServer/database"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"strconv"
	"strings"
	"time"

	"prismServer/utils"
)

func init() {
	executeTest.Register("TRMAnalysis", "VSA", newTRMMeasurement)
	results.Register("TRMAnalysis", results.NewTRMProcessor([]string{"Results"}))
}

func newTRMMeasurement() executeTest.Tester {
	var test pulseMeasurementTRM
	return &test
}

type pulseMeasurementTRM struct {
	pulseBaseTest
	batchNo      []int
	pulseTime    []int64
	pulseNo      []string
	averagePower []string
}

func (test *pulseMeasurementTRM) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.pulseBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *pulseMeasurementTRM) DBValidate() error {
	readTRMProfile := func() error {
		return test.readTRMProfile(test.test.TRMProfileName)
	}
	return test.validateAndPrepare(true, readTRMProfile)
}

func (test *pulseMeasurementTRM) getInstruments() {
	test.ctx.Progress.Instruments = []string{"VSA", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *pulseMeasurementTRM) measure(runner *StepRunner) error {
	start := time.Now()
	vsa := test.ctx.Selected.VSA

	freq := test.pulseSpec.CenterFrequency
	span := test.downlinkProfile.Span
	rbw := float64(test.downlinkProfile.RBW)
	noOfPulses := int(test.pulseParameters.AcquisitionTime * 1000 / int64(test.pulseSpec.PulsePeriod))
	acqTime := (float64(test.pulseParameters.AcquisitionTime) * 1e-3) + (test.pulseSpec.PulseWidth * 1e-6)
	yTop := test.pulseParameters.YTop
	pdiv := 5.0
	analLength := 100.0
	refLevel := test.pulseParameters.ThresholdLevel
	hystLevel := test.pulseParameters.Hysterisis
	points := int32(50000)
	bufferLength := int32(50000)
	onTime := float64(test.trmProfile.NoOfTRMs) * test.trmProfile.TimePerTRMInSecs
	fmt.Println("OnTime", onTime)

	test.batchNo = make([]int, 0)
	test.pulseTime = make([]int64, 0)
	test.pulseNo = make([]string, 0)
	test.averagePower = make([]string, 0)

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}

	runner.Run("Setting VSA Mode", true, func() {
		runner.Exec(vsa.StartVSA)
	})

	runner.Run("Setting Pulse Mode", true, func() {
		runner.Exec(vsa.SetPulseMode)
	})

	test.setTSMPathForVSA(runner)

	runner.Run("Setting Spectrum Parameters", true, func() {
		runner.Exec(setSpectrumParameters(vsa, freq, span, rbw))
	})

	runner.Run("Setting Pulse Parameters", true, func() {
		runner.Exec(setPulseParametersForVSA(vsa, acqTime, yTop, pdiv, analLength, refLevel, hystLevel, points, bufferLength))
	})

	runner.Run("Starting Measurement", true, func() {
		runner.Exec(vsa.StartMeasurement)
	})

	runner.Run("Waiting for first pulse", true, func() {
		runner.Exec(vsa.WaitTillFirstPulse)
	})

	runner.Run("Reading pulse periodically", true, func() {

		stopTime := time.Now().Add(time.Duration(onTime) * time.Second)
		fmt.Println("stopTime", stopTime)
		batchNo := 0
		for t := time.Now(); stopTime.Sub(t) > 0; t = time.Now() {
			retVal := runner.Exec(vsa.GetPulseAveragePower)
			if retVal.ErrorMessage != "" {
				fmt.Println("Error Message", retVal.ErrorMessage)
				break
			}
			if retVal.Result["TotalNoOfPulses"].Value == 0 {
				fmt.Println("Continuing")
				continue
			}

			test.batchNo = append(test.batchNo, batchNo)
			batchNo = batchNo + 1
			test.pulseTime = append(test.pulseTime, t.UnixMilli())
			test.pulseNo = append(test.pulseNo, fmt.Sprintf("%.2f", retVal.Result["PulseNo"].Values))
			test.averagePower = append(test.averagePower, fmt.Sprintf("%.2f", retVal.Result["PulseAvgPower"].Values))

		}
	})
	fmt.Println("Final Batch No", len(test.batchNo))
	//fmt.Println("PulseNo", test.pulseNo)
	//fmt.Println("Final AverPower No", test.averagePower)

	runner.Run("Stopping Measurement", true, func() {

		runner.Exec(vsa.StopMeasurement)
	})

	if !runner.Describe && runner.Err() == nil {
		var fileData strings.Builder
		fileData.WriteString("BatchNo,TimeStamp,PulseNo,AveragePower\n")
		for i := 0; i < len(test.batchNo); i++ {
			pulseStr := strings.Trim(test.pulseNo[i], "[] ")
			averageStr := strings.Trim(test.averagePower[i], "[] ")
			pulseNos := strings.Split(pulseStr, " ")
			averages := strings.Split(averageStr, " ")
			minVal := len(pulseNos)
			if len(averages) < minVal {
				minVal = len(averages)
			}
			for j := 0; j < minVal; j++ {
				fileData.WriteString(fmt.Sprintf("%d,", test.batchNo[i]))
				fileData.WriteString(fmt.Sprintf("%d,", test.pulseTime[i]))
				fileData.WriteString(pulseNos[j])
				fileData.WriteString(",")
				fileData.WriteString(averages[j])
				fileData.WriteString("\n")
			}
		}

		csvDir := utils.GetCSVResultDirectory()
		csvDir = filepath.Join(csvDir, test.testName)
		_ = os.MkdirAll(csvDir, 0755)
		fileName := test.testName
		if strings.TrimSpace(test.testCategory) != "" {
			fileName += "-" + test.testCategory
		}

		fileName += "-" + test.configName

		fileName = utils.GetOldTimeStampedFileName(fileName, test.reportTime) + ".csv"
		fullPath := filepath.Join(csvDir, fileName)
		_ = os.WriteFile(fullPath, []byte(fileData.String()), 0666)

		test.filenames = append(test.filenames, fullPath)

		detailsHeader := []string{"Parameter", "Value"}
		tp, _ := database.GetSelectedTestPhase()
		detailRows := make([][]string, 0)
		detailRows = append(detailRows, []string{"Config", test.configName})
		detailRows = append(detailRows, []string{"NoOfTRMs", strconv.FormatInt(test.trmProfile.NoOfTRMs, 10)})
		detailRows = append(detailRows, []string{"TimePerTRM", fmt.Sprintf("%0.2f", test.trmProfile.TimePerTRMInSecs)})
		detailRows = append(detailRows, []string{"TestPhase", tp})
		detailRows = append(detailRows, []string{"BatchSize", strconv.Itoa(noOfPulses)})

		test.saveResultsToCSV("Details", detailsHeader, detailRows)
		test.addFinalTestInformation(start)
	}

	return runner.Err()
}
