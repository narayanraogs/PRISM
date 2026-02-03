package measurements

import (
	"prismServer/driver"
	"prismServer/utils"
)

func setSpectrum(sa driver.SA, center float64, span float64, rbw float64, vbw float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sa.SetSpectrum(center, span, rbw, vbw)
	}
}

func waitForSweeps(sa driver.SA, sweep int) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sa.WaitForSweeps(sweep)
	}
}

func getFrequencyInCounterMode(sa driver.SA, marker int) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sa.GetFrequencyInCounterMode(marker)
	}
}

func peakSearch(sa driver.SA, maxHold bool, marker int) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sa.PeakSearch(maxHold, marker)
	}

}

func getMarkerValue(sa driver.SA, marker int) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sa.GetMarkerValue(marker)
	}
}

func setOccupiedBW(sa driver.SA, bwPercent float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sa.SetOccupiedBW(bwPercent)
	}
}

func setOBWSpectrum(sa driver.SA, freq float64, span float64, rbw float64, vbw float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sa.SetOBWSpectrum(freq, span, rbw, vbw)
	}
}

func getOccupiedBW(sa driver.SA) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sa.GetOccupiedBW()
	}
}

func getTraceDump(sa driver.SA, points int) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sa.GetTraceDump(points)
	}
}

func getAllPeaksAbove(sa driver.SA, excursion float64, markerNo int) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sa.GetAllPeaksAbove(excursion, markerNo)
	}
}

func getModIndex(sa driver.SA, offsetFreq float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sa.GetModIndex(offsetFreq)
	}
}

func getSAFrequencyDeviation(sa driver.SA, frequency float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sa.GetFrequencyDeviationFM(frequency)
	}
}
