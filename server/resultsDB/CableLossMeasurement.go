package resultsDB

import (
	"context"
	"encoding/json"
	"fmt"
	"prismServer/database"
	"slices"
	"strconv"
)

type cableLossMeasured struct {
	Frequency []string
	Measured  []string
}

func CheckIfCableLossPMReferenceExists() bool {
	ctx := context.Background()
	count, err := dbObject.checkIfCableLossPMReferenceExists(ctx)
	if err != nil {
		return false
	}
	return count != 0
}

func GetCableNames() ([]string, bool) {
	ctx := context.Background()
	cableNames, err := dbObject.getCableNames(ctx)
	if err != nil {
		return nil, false
	}
	return cableNames, true
}

func UpdateCableLossPMReference(date string, time string, loss string) bool {
	ctx := context.Background()
	var arg updateCableLossPMReferenceParams
	arg.Date = date
	arg.Time = time
	arg.Loss = loss
	err := dbObject.updateCableLossPMReference(ctx, arg)
	if err != nil {
		return false
	}
	return true
}

func InsertCableLoss(date string, time string, cableName string, cableLength int, loss string) bool {
	ctx := context.Background()
	var arg insertCableLossEntryParams
	arg.Date = date
	arg.Time = time
	arg.Loss = loss
	arg.CableName = cableName
	arg.CableLength = int64(cableLength)
	err := dbObject.insertCableLossEntry(ctx, arg)
	if err != nil {
		return false
	}
	return true
}

func GetAllCableLosses() ([][]string, bool) {
	var tbr = make([][]string, 0)
	var frequencies = make([]string, 0)
	ctx := context.Background()
	values, err := dbObject.getCableLosses(ctx)
	if err != nil {
		return nil, false
	}
	frequencyNames, _ := database.GetLossMeasurementFrequencyNames()
	for _, name := range frequencyNames {
		freq, _ := database.GetFrequencyForLossMeasurement(name)
		frequencies = append(frequencies, fmt.Sprintf("%.2f", freq))
	}

	firstRow := make([]string, 0)
	firstRow = append(firstRow, "SlNo", "Cable Name", "Cable Length", "Date", "Time")
	firstRow = append(firstRow, frequencyNames...)
	secondRow := make([]string, 0)
	secondRow = append(secondRow, "", "", "", "", "")
	secondRow = append(secondRow, frequencies...)
	tbr = append(tbr, firstRow, secondRow)
	for i, value := range values {
		var row = make([]string, 0)
		row = append(row, strconv.Itoa(i+1), value.CableName)
		row = append(row, strconv.Itoa(int(value.CableLength)))
		row = append(row, value.Date, value.Time)
		var cableLoss cableLossMeasured
		err := json.Unmarshal([]byte(value.Loss), &cableLoss)
		if err != nil {
			for i := 0; i < len(frequencies); i++ {
				row = append(row, "-")
			}
		} else {
			for _, freq := range frequencies {
				index := slices.Index(cableLoss.Frequency, freq)
				if index == -1 {
					row = append(row, "-")
				} else {
					row = append(row, cableLoss.Measured[index])
				}
			}
		}
		tbr = append(tbr, row)
	}
	return tbr, true
}

func GetCableLossPMReference() (string, bool) {
	ctx := context.Background()
	reference, err := dbObject.getPMReference(ctx)
	if err != nil {
		return "", false
	}
	return reference, true
}

//------------------------------TVAC Cable Loss----------------------------------------------------------------//

func CheckIfTVACCableLossPMReferenceExists() bool {
	ctx := context.Background()
	count, err := dbObject.checkIfTVACCableLossPMReferenceExists(ctx)
	if err != nil {
		return false
	}
	return count != 0
}

func CheckIfTVACCableReferenceExists() bool {
	ctx := context.Background()
	count, err := dbObject.checkIfTVACCableAsReferenceExists(ctx)
	if err != nil {
		return false
	}
	return count != 0
}

func UpdateTVACCableLossPMReference(date string, time string, reference string, loss string) bool {
	ctx := context.Background()
	var arg updateTVACCableLossPMReferenceParams
	arg.Date = date
	arg.Time = time
	arg.Reference = reference
	arg.Loss = loss

	err := dbObject.updateTVACCableLossPMReference(ctx, arg)
	if err != nil {
		return false
	}
	return true
}

func InsertTVACCableLoss(date string, time string, cableName string, testPhase string, reference string, loss string) bool {
	ctx := context.Background()
	var arg insertTVACCableLossEntryParams
	arg.Date = date
	arg.Time = time
	arg.Reference = reference
	arg.Loss = loss
	arg.CableName = cableName
	arg.TestPhase = testPhase
	err := dbObject.insertTVACCableLossEntry(ctx, arg)
	if err != nil {
		return false
	}
	return true
}

func GetAllTVACCableLosses() ([][]string, bool) {
	var tbr = make([][]string, 0)
	var frequencies = make([]string, 0)
	ctx := context.Background()
	values, err := dbObject.getTVACCableLosses(ctx)
	if err != nil {
		return nil, false
	}
	frequencyNames, _ := database.GetLossMeasurementFrequencyNames()
	for _, name := range frequencyNames {
		freq, _ := database.GetFrequencyForLossMeasurement(name)
		frequencies = append(frequencies, fmt.Sprintf("%.2f", freq))
	}

	firstRow := make([]string, 0)
	firstRow = append(firstRow, "SlNo", "Cable Name", "Test Phase", "Reference", "Date", "Time")
	firstRow = append(firstRow, frequencyNames...)
	secondRow := make([]string, 0)
	secondRow = append(secondRow, "", "", "", "", "", "")
	secondRow = append(secondRow, frequencies...)
	tbr = append(tbr, firstRow, secondRow)
	for i, value := range values {
		var row = make([]string, 0)
		row = append(row, strconv.Itoa(i+1), value.CableName)
		row = append(row, value.TestPhase)
		row = append(row, value.Reference, value.Date, value.Time)
		var cableLoss cableLossMeasured
		err = json.Unmarshal([]byte(value.Loss), &cableLoss)

		for _, freq := range frequencies {
			index := slices.Index(cableLoss.Frequency, freq)
			if index == -1 {
				row = append(row, "-")
			} else {
				row = append(row, cableLoss.Measured[index])
			}
		}

		tbr = append(tbr, row)
	}
	return tbr, true
}

func GetTVACCableLossPMReference() (string, bool) {
	ctx := context.Background()
	reference, err := dbObject.getTVACPMReference(ctx)
	if err != nil {
		return "", false
	}
	return reference, true
}

func GetTVACCableReference(cableName string) (string, bool) {
	ctx := context.Background()
	cableReference, err := dbObject.getTVACReferenceCableLosses(ctx, cableName)
	if err != nil {
		return "", false
	}
	return cableReference, true
}
