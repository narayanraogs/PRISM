package measurements

import (
	"fmt"
	"os"
	"path/filepath"
	"prismServer/database"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/reports"
	"prismServer/utils"
	"time"
)

func init() {
	executeTest.Register("HighResolutionPulse", "VSA", newHRMeasurementVSA)
	results.Register("HighResolutionPulse", results.NewHRProcessor([]string{"HR Pulse", "Normal Pulse"}))
}

func newHRMeasurementVSA() executeTest.Tester {
	var test hrMeasurementVSA
	return &test
}

type hrMeasurementVSA struct {
	pulseBaseTest
	lrSpec database.SpecPL
}

func (test *hrMeasurementVSA) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.pulseBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *hrMeasurementVSA) DBValidate() error {
	var ok bool
	test.lrSpec, ok = database.GetFullSpec(test.configName)
	if !ok {
		return fmt.Errorf("Unable to read LR Spec for %s", test.configName)
	}
	return test.validateAndPrepare(true, nil)
}

func (test *hrMeasurementVSA) getInstruments() {
	test.ctx.Progress.Instruments = []string{"VSA", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *hrMeasurementVSA) measure(runner *StepRunner) error {
	start := time.Now()
	vsa := test.ctx.Selected.VSA

	freq := test.pulseSpec.CenterFrequency
	span := test.downlinkProfile.Span
	rbw := float64(test.downlinkProfile.RBW)
	//noOfPulses := int(test.pulseParameters.AcquisitionTime * 1000 / int64(test.pulseSpec.PulsePeriod))
	acqTime := (float64(test.pulseParameters.AcquisitionTime) * 1e-3) + (test.pulseSpec.PulseWidth * 1e-6)
	yTop := test.pulseParameters.YTop
	pdiv := 5.0
	analLength := 100.0
	refLevel := test.pulseParameters.ThresholdLevel
	hystLevel := test.pulseParameters.Hysterisis
	points := int32(50000)
	bufferLength := int32(50000)
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

	runner.Run("Stopping Measurement", true, func() {
		runner.Wait(int(test.pulseSpec.OnTime))
		runner.Exec(vsa.StopMeasurement)
	})

	runner.Run("Fetching Spectrum", true, func() {
		resp := runner.Exec(getScreenshot(vsa, ""))
		if runner.execErr != nil {
			return
		}
		caption := fmt.Sprintf("VSA Downlink Spectrum for %s", test.configName)
		test.spectra = append(test.spectra, reports.Images{Caption: caption, FileData: resp.Result["Screenshot"].String})
	})

	runner.Run("Getting Pulse Parameters", true, func() {
		runner.Exec(vsa.GetPulseParamaeters)
	})

	if !runner.Describe && runner.Err() == nil {
		detailsHeader := []string{"Parameter", "Value"}
		tp, _ := database.GetSelectedTestPhase()
		detailRows := make([][]string, 0)
		detailRows = append(detailRows, []string{"Config", test.configName})
		detailRows = append(detailRows, []string{"TestPhase", tp})
		detailRows = append(detailRows, []string{"BatchSize", "-1"})
		files, err := test.segregateHRPulses()
		if err == nil {
			_ = os.WriteFile(utils.Config.BaseFolder+"/temp/tempHR.csv", []byte(files[0]), 0666)
			_ = os.WriteFile(utils.Config.BaseFolder+"/temp/tempLR.csv", []byte(files[1]), 0666)
			test.saveResultsAndCSV("HRPulse", utils.Config.BaseFolder+"/temp/tempHR.csv")
			test.saveResultsAndCSV("LRPulse", utils.Config.BaseFolder+"/temp/tempLR.csv")
			test.saveResultsToCSV("Details", detailsHeader, detailRows)
			test.addFinalTestInformation(start)
		}
	}

	return runner.Err()
}

func (test *hrMeasurementVSA) segregateHRPulses() ([]string, error) {
	var pulse utils.PulseParameters
	fileName := filepath.Join(utils.Config.BaseFolder, "/temp/temp.csv")
	err := pulse.LoadFile(fileName, -1)
	if err != nil {
		return nil, err
	}
	//todo: Column name from Configuration
	return pulse.SegregateHighResolution("Pulse_Off_Time1", test.pulseSpec.PulseSeperation.Float64, test.lrSpec.PulseSeperation.Float64)
}
