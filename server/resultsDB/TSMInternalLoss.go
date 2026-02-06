package resultsDB

import "context"

func CreateNewTSMInternalLoss(table [][]string) bool {
	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		return false
	}
	defer tx.Rollback()
	q := dbObject.WithTx(tx)
	err = q.clearTSMInternalLoss(ctx)
	if err != nil {
		return false
	}
	for i, row := range table {
		var arg insertTSMInternalLossParams
		arg.LossID = int64(i + 1)
		arg.InputPort = row[0]
		arg.OutputPort = row[1]
		arg.PathMnemonic.String = row[2]
		arg.PathMnemonic.Valid = true
		arg.MeasuredLosses = row[3]
		err = q.insertTSMInternalLoss(ctx, arg)
		if err != nil {
			return false
		}
	}
	err = tx.Commit()
	return err == nil
}

func GetTSMInternalLossStructure() ([]TSMInternalLoss, error) {
	ctx := context.Background()
	tsm, err := dbObject.getAllTSMInternalLoss(ctx)
	return tsm, err
}

func GetTSMInternalLossTable() ([][]string, bool) {
	ctx := context.Background()
	tsm, err := dbObject.getAllTSMInternalLoss(ctx)
	if err != nil {
		return nil, false
	}
	var tbr = make([][]string, 0)
	for _, t := range tsm {
		row := make([]string, 0)
		row = append(row, t.InputPort, t.OutputPort, t.PathMnemonic.String, t.MeasuredLosses, t.MeasurementCompleted)
		tbr = append(tbr, row)
	}
	return tbr, true
}

func GetTSMInternalMeasuredLoss(inputPort string, outputPort string) (string, string, bool) {
	ctx := context.Background()
	var arg getMeasuredTSMInternalLossParams
	arg.InputPort = inputPort
	arg.OutputPort = outputPort
	tsm, err := dbObject.getMeasuredTSMInternalLoss(ctx, arg)
	if err != nil {
		return "", "", false
	}
	return tsm.PathMnemonic.String, tsm.MeasuredLosses, true
}

func UpdateTSMInternalLoss(inputPort string, outputPort string, measured string) bool {
	ctx := context.Background()
	var arg updateTSMInternalLossParams
	arg.InputPort = inputPort
	arg.OutputPort = outputPort
	arg.MeasuredLosses = measured
	err := dbObject.updateTSMInternalLoss(ctx, arg)
	return err == nil
}

func GetTSMInternalLossPMOffset() (string, bool) {
	ctx := context.Background()
	loss, err := dbObject.getPMOffsetForTSMInternalLoss(ctx)
	return loss, err == nil
}

func GetTSMInternalLossCableLoss() (string, bool) {
	ctx := context.Background()
	loss, err := dbObject.getCableLossForTSMInternalLoss(ctx)
	return loss, err == nil
}

func UpdateTSMInternalLossPMOffset(measured string) bool {
	ctx := context.Background()
	err := dbObject.updatePMOffsetTSMInternalLoss(ctx, measured)
	return err == nil
}

func UpdateTSMInternalLossCableLoss(measured string) bool {
	ctx := context.Background()
	err := dbObject.updateCableLossTSMInternalLoss(ctx, measured)
	return err == nil
}
