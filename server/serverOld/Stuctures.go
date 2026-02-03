package server

import (
	"prismServer/executeTest"
	"prismServer/global"
	"prismServer/tne"
	"strings"
	"time"
)

type client struct {
	global         global.ClientGlobal
	orchestrator   *executeTest.Orchestrator
	gtxMeasurement *tne.GroundTransmitterMeasurement
	lastSeen       time.Time
}

type emptyRequest struct {
	ID string
}

type ackResponse struct {
	Message string
	OK      bool
}

type parameterValue struct {
	Name   string
	Values []string
}

type getRequest struct {
	ID         string
	Parameters []string
}

type getResponse struct {
	Values  []parameterValue
	OK      bool
	Message string
}

type setRequest struct {
	ID     string
	Values []parameterValue
}

type actionRequest struct {
	ID         string
	Action     string
	Parameters []parameterValue
}

func (act *actionRequest) getParam(name string) []string {
	for _, p := range act.Parameters {
		if strings.EqualFold(p.Name, name) {
			return p.Values
		}
	}
	return nil
}

type valueRequest struct {
	ID            string
	ParameterName string
}

type tableValueResponse struct {
	Values  [][]string
	Message string
	OK      bool
}
