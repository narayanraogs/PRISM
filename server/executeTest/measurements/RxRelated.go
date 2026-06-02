package measurements

import (
	"database/sql"
	"fmt"
	"prismServer/database"
	"prismServer/utils"
)

type rxRelated struct {
	rxSpec                  database.SpecRx
	scLoss                  float64
	pmLoss                  float64
	saLoss                  float64
	uplinkProfile           database.SpectrumProfile
	tmtc                    database.SpecRxTMTC
	powerLevels             []float64
	noOfCommandsAtThreshold int
	noOfCommandsNominal     int
	frequencyProfile        database.FrequencyProfile
}

func (rx *rxRelated) readRxDetails(cfg string, rxName sql.NullString, testDetails database.Test) error {
	var ok bool
	if !rxName.Valid {
		return fmt.Errorf("receiver name not specified in Configuration")
	}
	rx.rxSpec, ok = database.GetRxDetails(rxName.String)
	if !ok {
		return fmt.Errorf("unable to get the Specification of receiver %s", rxName.String)
	}
	rx.tmtc, ok = database.GetRxTMTC(rxName.String)
	if !ok {
		return fmt.Errorf("unable to get the tm and tc of receiver %s", rxName.String)
	}
	_, sa, pm, sc, ok := database.GetCurrentUplinkLoss(cfg, utils.GetTestPhase())
	if !ok {
		return fmt.Errorf("unable to get Losses for %s, %s", cfg, utils.GetTestPhase())
	}
	rx.saLoss = sa
	rx.pmLoss = pm
	rx.scLoss = sc

	if !testDetails.PowerProfileName.Valid {
		return fmt.Errorf("power Profile name is empty in test table")
	}

	powers, ok := database.GetPowerLevels(testDetails.PowerProfileName.String)
	if !ok {
		return fmt.Errorf("cannot read power Profile %s", testDetails.PowerProfileName.String)
	}
	rx.powerLevels = make([]float64, 0)
	rx.powerLevels = append(rx.powerLevels, powers...)

	th, nom, ok := database.GetNoOfCommandsInProfile(testDetails.PowerProfileName.String)
	rx.noOfCommandsAtThreshold = th
	rx.noOfCommandsNominal = nom

	if !testDetails.ULProfileName.Valid {
		return fmt.Errorf("uplink Profile name is empty in test table")
	}
	rx.uplinkProfile, ok = database.GetSpectrumProfile(testDetails.ULProfileName.String)
	if !ok {
		return fmt.Errorf("downlink Profile %s is not present in Database", testDetails.ULProfileName.String)
	}

	return nil
}

func (rx *rxRelated) getFrequencyProfile(profileName sql.NullString) error {
	if !profileName.Valid {
		return fmt.Errorf("frequency profile name not specified in test")
	}
	var ok bool
	rx.frequencyProfile, ok = database.GetFrequencyProfile(profileName.String)
	if !ok {
		return fmt.Errorf("unable to get frequency profile details details for %s", profileName.String)
	}
	return nil
}
