package common

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"
)

// DarksideZcashdState Equivalent to the cache but for test purposes so it
// is more easily modified.
type DarksideZcashdState struct {
	startHeight       int
	saplingActivation int
	branchID          string
	chainSubName      string
	// Should always be nonempty. Index 0 is the block at height startHeight.
	blocks               []string
	incomingTransactions [][]byte
	serverStart          time.Time
}

var state *DarksideZcashdState = nil

// DarkSideRawRequest Use this when DarkSide is enabled
func DarkSideRawRequest(method string, params []json.RawMessage) (json.RawMessage, error) {

	if state == nil {
		state = &DarksideZcashdState{
			startHeight:          1000,
			saplingActivation:    1000,
			branchID:             "2bb40e60", // Blossom
			chainSubName:         "darkside",
			blocks:               make([]string, 0),
			incomingTransactions: make([][]byte, 0),
			serverStart:          time.Now(),
		}

		testBlocks, err := os.Open("./testdata/default-darkside-blocks")
		if err != nil {
			Log.Fatal("Error loading default darksidewalletd blocks")
		}
		scan := bufio.NewScanner(testBlocks)
		for scan.Scan() { // each line (block)
			block := scan.Bytes()
			state.blocks = append(state.blocks, string(block))
		}
	}

	if time.Now().Sub(state.serverStart).Minutes() >= 30 {
		Log.Fatal("Shutting down darksidewalletd to prevent accidental deployment in production.")
	}

	switch method {
	case "getblockchaininfo":
		type upgradeinfo struct {
			// there are other fields that aren't needed here, omit them
			ActivationHeight int `json:"activationheight"`
		}
		type consensus struct {
			Nextblock string `json:"nextblock"`
			Chaintip  string `json:"chaintip"`
		}
		blockchaininfo := struct {
			Chain     string                 `json:"chain"`
			Upgrades  map[string]upgradeinfo `json:"upgrades"`
			Headers   int                    `json:"headers"`
			Consensus consensus              `json:"consensus"`
		}{
			Chain: state.chainSubName,
			Upgrades: map[string]upgradeinfo{
				"76b809bb": upgradeinfo{ActivationHeight: state.saplingActivation},
			},
			Headers:   state.startHeight + len(state.blocks) - 1,
			Consensus: consensus{state.branchID, state.branchID},
		}
		return json.Marshal(blockchaininfo)

	case "getblock":
		var height string
		err := json.Unmarshal(params[0], &height)
		if err != nil {
			return nil, errors.New("failed to parse getblock request")
		}

		heightI, err := strconv.Atoi(height)
		if err != nil {
			return nil, errors.New("error parsing height as integer")
		}
		index := heightI - state.startHeight

		if index == len(state.blocks) {
			// The current ingestor keeps going until it sees this error,
			// meaning it's up to the latest height.
			return nil, errors.New("-8:")
		}

		if index < 0 || index > len(state.blocks) {
			// If an integration test can reach this, it could be a bug, so generate an error.
			Log.Errorf("getblock request made for out-of-range height %d (have %d to %d)", heightI, state.startHeight, state.startHeight+len(state.blocks)-1)
			return nil, errors.New("-8:")
		}

		return []byte("\"" + state.blocks[index] + "\""), nil

	case "getaddresstxids":
		// Not required for minimal reorg testing.
		return nil, errors.New("not implemented yet")

	case "getrawtransaction":
		// Not required for minimal reorg testing.
		return nil, errors.New("not implemented yet")

	case "sendrawtransaction":
		var rawtx string
		err := json.Unmarshal(params[0], &rawtx)
		if err != nil {
			return nil, errors.New("failed to parse sendrawtransaction JSON")
		}
		txbytes, err := hex.DecodeString(rawtx)
		if err != nil {
			return nil, errors.New("failed to parse sendrawtransaction value as a hex string")
		}
		state.incomingTransactions = append(state.incomingTransactions, txbytes)
		return nil, nil

	case "x_setstate":
		var newState map[string]interface{}

		err := json.Unmarshal(params[0], &newState)
		if err != nil {
			Log.Fatal("Could not unmarshal the provided state.")
		}

		blockStrings := make([]string, 0)
		for _, blockStr := range newState["blocks"].([]interface{}) {
			blockStrings = append(blockStrings, blockStr.(string))
		}

		state = &DarksideZcashdState{
			startHeight:          int(newState["startHeight"].(float64)),
			saplingActivation:    int(newState["saplingActivation"].(float64)),
			branchID:             newState["branchID"].(string),
			chainSubName:         newState["chainSubName"].(string),
			blocks:               blockStrings,
			incomingTransactions: state.incomingTransactions,
			serverStart:          state.serverStart,
		}

		return nil, nil

	case "x_getincomingtransactions":
		txlist := "["
		for i, tx := range state.incomingTransactions {
			txlist += "\"" + hex.EncodeToString(tx) + "\""
			// add commas after all but the last
			if i < len(state.incomingTransactions)-1 {
				txlist += ", "
			}
		}
		txlist += "]"

		return []byte(txlist), nil

	default:
		return nil, errors.New("there was an attempt to call an unsupported RPC")
	}
}
