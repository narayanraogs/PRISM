package measurements

import (
	"fmt"
	"prismServer/database"
)

type tpRelated struct {
	tpSpec        database.SpecTp
	tpRangingSpec database.SpecTpRanging
}

func (tp *tpRelated) readTpSpecTransponder(tpName string) error {
	if tpName == "" {
		return fmt.Errorf("transponder Name not proper in specTpRanging Table")
	}
	var ok bool

	tp.tpRangingSpec, ok = database.GetTpSpecs(tpName)
	if !ok {
		return fmt.Errorf("unable to get the Ranging Specification of Transponder %s", tpName)
	}
	return nil
}
