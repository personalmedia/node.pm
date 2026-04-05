// Copyright (c) 2026 DATAMIX.IO
// Author: L3DLP

package api

import (
	"encoding/json"
	"testing"
)

func FuzzNodeJSON(f *testing.F) {
	// Corpus initial
	f.Add(`[{"status":"OK", "domain":"test.com", "subdomain":"api", "certificate":"...", "private_key":"...", "local_target":"8080"}]`)
	f.Add(`{}`)
	f.Add(`[]`)
	f.Add(`{"status":"ERROR"}`)
	f.Add(`""`)

	f.Fuzz(func(t *testing.T, jsonStr string) {
		var nodes []NodeInfo
		// On teste juste que le parsing ne crash pas
		_ = json.Unmarshal([]byte(jsonStr), &nodes)
	})
}
