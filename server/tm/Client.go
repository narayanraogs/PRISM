package tm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"prismServer/logger"
	"prismServer/utils"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	paramCache     map[string]parameter
	paramCacheOnce sync.Once
	paramCacheErr  error
)

func getParamCache() (map[string]parameter, error) {
	paramCacheOnce.Do(func() {
		paramCache, paramCacheErr = fetchFullParamsFromServer()
	})
	return paramCache, paramCacheErr
}

func fetchFullParamsFromServer() (map[string]parameter, error) {
	u := "http://" + utils.Config.TMServer.IP + fmt.Sprintf(":%d", utils.Config.TMServer.PortNo) + "/pid_info?sc_id=" + utils.Config.UMACSSatelliteName
	resp, err := http.Get(u)
	if err != nil {
		logger.Log.Error("Failed to fetch TM parameters", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Log.Error("Failed to fetch TM parameters", "status", resp.Status)
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("Failed to read TM parameters response", "error", err)
		return nil, err
	}

	var params []parameter
	if err := json.Unmarshal(body, &params); err != nil {
		logger.Log.Error("Failed to unmarshal TM parameters", "error", err)
		return nil, err
	}

	paramCache := make(map[string]parameter)
	for _, p := range params {
		paramCache[utils.GetComparableMnemonic(p.Mnemonic)] = p
	}
	return paramCache, nil
}

func Subscribe(mnemonics []string, output chan<- Parameter, onlyOnChange bool) {
	go processRequest(mnemonics, output, false, onlyOnChange)
}

func Fetch(mnemonics []string, output chan<- Parameter) {
	go processRequest(mnemonics, output, true, true)
}

func processRequest(mnemonics []string, output chan<- Parameter, isFetch bool, onlyOnChange bool) {
	if isFetch {
		defer close(output)
	}

	// Get the cached parameter metadata.
	paramCache, err := getParamCache()
	if err != nil {
		logger.Log.Error("Failed to get TM parameter list", "error", err)
		for _, m := range mnemonics {
			output <- Parameter{Param: m, OK: false, Error: "Failed to get TM parameter list: " + err.Error()}
		}
		return
	}

	paramsByStream := make(map[string][]string)
	originalMnemonicMap := make(map[string]string)

	for _, m := range mnemonics {
		parts := strings.SplitN(m, ":", 2)
		stream := "TM1" // Default stream
		paramName := m
		if len(parts) == 2 {
			stream = parts[0]
			paramName = parts[1]
		}
		paramsByStream[stream] = append(paramsByStream[stream], paramName)
		originalMnemonicMap[utils.GetComparableMnemonic(paramName)] = m
	}

	var wg sync.WaitGroup
	for stream, params := range paramsByStream {
		wg.Add(1)
		go subscribeToStream(&wg, stream, params, originalMnemonicMap, paramCache, output, isFetch, onlyOnChange)
	}

	wg.Wait()
}

func subscribeToStream(wg *sync.WaitGroup, stream string, params []string, origMap map[string]string, paramCache map[string]parameter, output chan<- Parameter, isFetch bool, onlyOnChange bool) {
	defer wg.Done()

	addr := utils.Config.TMServer.IP + ":" + strconv.Itoa(utils.Config.TMServer.PortNo)
	u := url.URL{Scheme: "ws", Host: addr, Path: utils.Config.TMServer.Path}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logger.Log.Error("TM server unavailable", "error", err)
		for _, p := range params {
			output <- Parameter{Param: origMap[utils.GetComparableMnemonic(p)], OK: false, Error: "TM server unavailable: " + err.Error()}
		}
		return
	}
	defer conn.Close()
	var pidMnemoicMap = make(map[string]string)
	var receivedMap = make(map[string]bool)

	var pidsToRequest []string
	for _, p := range params {
		paramInfo, ok := paramCache[utils.GetComparableMnemonic(p)]
		if !ok {
			logger.Log.Error("TM Parameter not found on server", "parameter", p)
			output <- Parameter{Param: origMap[utils.GetComparableMnemonic(p)], OK: false, Error: "Parameter not found on server"}
			continue
		}
		pidsToRequest = append(pidsToRequest, paramInfo.PID)
		pidMnemoicMap[paramInfo.PID] = utils.GetComparableMnemonic(p)
	}

	if len(pidsToRequest) == 0 {
		return // No valid parameters to subscribe to
	}

	req := request{
		UserID:   "PRISM",
		MsgType:  "ntm",
		OnChange: onlyOnChange,
		MsgPayload: messagePayload{
			ScID:       utils.Config.UMACSSatelliteName,
			Stream:     "ANY-" + stream,
			Action:     "subscribe",
			Parameters: pidsToRequest,
		},
	}

	if err := conn.WriteJSON(req); err != nil {
		logger.Log.Error("Failed to send subscription request", "error", err)
		for _, p := range params {
			output <- Parameter{Param: origMap[utils.GetComparableMnemonic(p)], OK: false, Error: "Failed to send subscription request: " + err.Error()}
		}
		return
	}

	paramsToReceive := len(pidsToRequest)
	for {
		var resp response
		if err := conn.ReadJSON(&resp); err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				logger.Log.Error("Connection read error", "error", err)
				for _, p := range params {
					output <- Parameter{Param: origMap[utils.GetComparableMnemonic(p)], OK: false, Error: "Connection read error: " + err.Error()}
				}
			}
			return
		}

		for _, pInfo := range resp.MsgPayload.ParametersInfo {
			comparableMnemonic := pidMnemoicMap[pInfo.Param]
			originalMnemonic := origMap[comparableMnemonic]
			if originalMnemonic == "" {
				continue // Not a param we requested
			}
			if !receivedMap[originalMnemonic] {
				pInfo.OK = true
				pInfo.Param = originalMnemonic
				output <- pInfo
				receivedMap[originalMnemonic] = true
				if isFetch {
					paramsToReceive--
				}
			}
		}

		if isFetch && paramsToReceive <= 0 {
			req.MsgPayload.Action = "unsubscribe"
			conn.WriteJSON(req) // Best effort, ignore error
			return
		}
	}
}
