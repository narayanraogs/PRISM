package resultsDB

import (
	"context"
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

func GetCableNamesForCableLoss() ([]string, bool) {
	ctx := context.Background()
	cableNames, err := dbObject.getCableNamesForCableLoss(ctx)
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

func InsertCableLoss(date string, time string, cableName string, cableLength float64, loss string) bool {
	ctx := context.Background()
	var arg insertCableLossEntryParams
	arg.Date = date
	arg.Time = time
	arg.Loss = loss
	arg.CableName = cableName
	arg.CableLength = cableLength
	err := dbObject.insertCableLossEntry(ctx, arg)
	if err != nil {
		return false
	}
	return true
}

func GetAllCableLosses() ([]CableLoss, error) {
	ctx := context.Background()
	values, err := dbObject.getCableLosses(ctx)
	return values, err
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

func GetAllTVACCableLosses() ([]TVACCableLoss, bool) {
	ctx := context.Background()
	values, err := dbObject.getTVACCableLosses(ctx)
	if err != nil {
		return nil, false
	}
	return values, true
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
