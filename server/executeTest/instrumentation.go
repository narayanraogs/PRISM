package executeTest

import (
	"context"
	"fmt"
	"prismServer/utils"

	"golang.org/x/sync/errgroup"
)

type connectionChecker interface {
	CheckConnection() utils.CommandResponse
}

func (e *Engine) checkInstrumentConnection(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)
	e.context.Progress.CurrentStep = "InstrumentValidation"
	e.context.UpdateChannel <- *e.context.Progress

	instrumentMap := map[string]connectionChecker{
		"PM":  &e.context.Selected.PM,
		"SA":  &e.context.Selected.SA,
		"SG":  &e.context.Selected.SG,
		"TSM": &e.context.Selected.TSM,
		"GTX": &e.context.Selected.GTx,
		"VSA": &e.context.Selected.VSA,
		"PPM": &e.context.Selected.PPM,
	}

	for i, instName := range e.context.Progress.Instruments {
		i := i
		instName := instName

		instrument, ok := instrumentMap[instName]
		if !ok {
			return fmt.Errorf("unknown instrument configured for test: %s", instName)
		}

		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
				resp := instrument.CheckConnection()
				if resp.Success {
					e.context.Progress.InstrumentStatus[i] = "Connected"
				} else {
					e.context.Progress.InstrumentStatus[i] = "NotConnected"
				}
				e.context.UpdateChannel <- *e.context.Progress

				if !resp.Success {
					return fmt.Errorf("%s not connected", instName)
				}
				return nil
			}
		})
	}
	err := g.Wait()
	return err
}
