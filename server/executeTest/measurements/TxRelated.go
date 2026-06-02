package measurements

import (
	"database/sql"
	"fmt"
	"prismServer/database"
	"prismServer/utils"
)

type txRelated struct {
	txSpec               database.SpecTx
	pmLoss               float64
	saLoss               float64
	downlinkProfile      database.SpectrumProfile
	txSpecHarmonic       []database.SpecTxHarmonic
	downlinkPowerProfile database.DownlinkPowerProfile
	txSpecSubCarrier     map[string]database.SpecTxSubCarrier
}

func (tx *txRelated) readTxDetails(cfg string, txName sql.NullString, dlProfileName sql.NullString) error {
	var ok bool
	if !txName.Valid {
		return fmt.Errorf("transmitter name not specified in Configuration")
	}
	tx.txSpec, ok = database.GetTxSpecs(txName.String)
	if !ok {
		return fmt.Errorf("unable to get the Specification of Transmitter %s", txName.String)
	}
	_, sa, pm, ok := database.GetCurrentDownlinkLoss(cfg, utils.GetTestPhase())
	if !ok {
		return fmt.Errorf("unable to get Losses for %s, %s", cfg, utils.GetTestPhase())
	}
	tx.saLoss = sa
	tx.pmLoss = pm

	if !dlProfileName.Valid {
		return fmt.Errorf("downlink Profile name is empty in test table")
	}
	tx.downlinkProfile, ok = database.GetSpectrumProfile(dlProfileName.String)
	if !ok {
		return fmt.Errorf("downlink Profile %s is not present in Database", dlProfileName.String)
	}

	return nil
}

func (tx *txRelated) readTxHarmonicsDetails(txName sql.NullString) error {
	if !txName.Valid {
		return fmt.Errorf("transmitter name not specified in Configuration")
	}

	harmonicDetails, ok := database.GetTxHarmonicsDetails(txName.String)
	if !ok {
		return fmt.Errorf("unable to get Harmonic details for %s", txName.String)
	}
	tx.txSpecHarmonic = make([]database.SpecTxHarmonic, 0)
	tx.txSpecHarmonic = append(tx.txSpecHarmonic, harmonicDetails...)
	return nil
}

func (tx *txRelated) readDownlinkPowerProfile(profileName sql.NullString) error {
	if !profileName.Valid {
		return fmt.Errorf("config name not specified in Configurations Table")
	}
	var ok bool
	tx.downlinkPowerProfile, ok = database.GetDownlinkPowerProfile(profileName)
	if !ok {
		return fmt.Errorf("unable to get Downlink Profile details for %s", profileName.String)
	}

	return nil
}

func (tx *txRelated) readTxSubCarrierDetails(txName sql.NullString) error {
	if !txName.Valid {
		return fmt.Errorf("Transmitter Name not proper in SpecTxSubCarriers Table")
	}
	subCarrierDetails, ok := database.GetTxSubCarriersDetails(txName.String)
	if !ok {
		return fmt.Errorf("unable to get sub Carrier Profile details for %s", txName.String)
	}
	tx.txSpecSubCarrier = make(map[string]database.SpecTxSubCarrier)
	for _, sub := range subCarrierDetails {
		tx.txSpecSubCarrier[sub.SubCarrierName] = sub
	}
	return nil
}
