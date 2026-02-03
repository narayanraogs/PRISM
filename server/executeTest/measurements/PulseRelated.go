package measurements

import (
	"database/sql"
	"fmt"
	"prismServer/database"
	"prismServer/utils"
)

type pulseRelated struct {
	pulseSpec       database.SpecPL
	pmLoss          float64
	saLoss          float64
	scLoss          float64
	downlinkProfile database.SpectrumProfile
	pulseParameters database.PulseProfile
	pulseProfile    string
	trmProfile      database.TRMProfile
	uplinkPowers    []float64
	uplinkFrequency database.FrequencyProfile
}

func (pl *pulseRelated) readPulseDetails(cfgName string, testType string, testCategory string, dlProfileName sql.NullString,
	pulseProfileName string) error {
	var ok bool
	if cfgName == "" {
		return fmt.Errorf("configuration name not specified in SpecPL")
	}
	if testType == "HighResolutionPulse" {
		pl.pulseSpec, ok = database.GetSpecHRMode(cfgName)
	} else {
		pl.pulseSpec, ok = database.GetFullSpec(cfgName)
	}

	_, sa, pm, ok := database.GetCurrentDownlinkLoss(cfgName, utils.GetTestPhase())
	if !ok {
		return fmt.Errorf("unable to get Losses for %s, %s", cfgName, utils.GetTestPhase())
	}
	pl.saLoss = sa
	pl.pmLoss = pm

	if !dlProfileName.Valid {
		return fmt.Errorf("downlink Profile name is empty in test table")
	}
	pl.downlinkProfile, ok = database.GetSpectrumProfile(dlProfileName.String)
	if !ok {
		return fmt.Errorf("downlink Profile %s is not present in Database", dlProfileName.String)
	}

	pl.pulseProfile, ok = database.GetPulsePowerProfile(cfgName, testType, testCategory)
	if !ok {
		return fmt.Errorf("pulse profile for %s is not present in Database", cfgName)
	}
	fmt.Println("PulseProfileName", pulseProfileName)

	pl.pulseParameters, ok = database.GetPPMRelatedParameters(pulseProfileName)

	return nil
}

func (pl *pulseRelated) readUplinkDetails(cfgName string, testType string, testCategory string, ulSpectrumProfileName sql.NullString,
	frequencyProfileName sql.NullString, powerProfileName sql.NullString) error {
	var ok bool
	if cfgName == "" {
		return fmt.Errorf("configuration name not specified in SpecPL")
	}
	pl.pulseSpec, ok = database.GetFullSpec(cfgName)

	_, sa, _, sc, ok := database.GetCurrentUplinkLoss(cfgName, utils.GetTestPhase())
	if !ok {
		return fmt.Errorf("unable to get Losses for %s, %s", cfgName, utils.GetTestPhase())
	}
	pl.saLoss = sa
	pl.scLoss = sc

	if !ulSpectrumProfileName.Valid {
		return fmt.Errorf("uplink Profile name is empty in test table")
	}
	pl.downlinkProfile, ok = database.GetSpectrumProfile(ulSpectrumProfileName.String)
	if !ok {
		return fmt.Errorf("uplink Profile %s is not present in Database", ulSpectrumProfileName.String)
	}

	if !powerProfileName.Valid {
		return fmt.Errorf("power Profile name is empty in test table")
	}
	pl.uplinkPowers, ok = database.GetPowerLevels(powerProfileName.String)
	if !ok {
		return fmt.Errorf("power Profile %s is not present in Database", ulSpectrumProfileName.String)
	}

	if !frequencyProfileName.Valid {
		return fmt.Errorf("frequency Profile name is empty in test table")
	}
	pl.uplinkFrequency, ok = database.GetFrequencyProfile(frequencyProfileName.String)
	if !ok {
		return fmt.Errorf("power Profile %s is not present in Database", ulSpectrumProfileName.String)
	}

	return nil
}

func (pl *pulseRelated) readTRMProfile(profileName sql.NullString) error {
	if !profileName.Valid {
		return fmt.Errorf("trm profile name not specified in Tests Table")
	}
	var ok bool
	pl.trmProfile, ok = database.GetTRMParameters(profileName)
	if !ok {
		return fmt.Errorf("unable to get TRM Profile details for %s", profileName.String)
	}

	return nil
}
