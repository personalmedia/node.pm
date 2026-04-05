// Copyright (c) 2026 DATAMIX.IO
// Author: L3DLP

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// NodeInfo représente les informations d'un nœud reçues du Bastion
type NodeInfo struct {
	Status      string `json:"status"`
	Domain      string `json:"domain"`
	Subdomain   string `json:"subdomain"`
	Certificate string `json:"certificate,omitempty"`
	PrivateKey  string `json:"private_key,omitempty"`
	LocalTarget string `json:"local_target"`
}

// FetchNodes interroge le Bastion pour obtenir les configurations de nœuds
func FetchNodes(token string) ([]NodeInfo, error) {
	serverURL := os.Getenv("NODE_PM_SERVER")
	if serverURL == "" {
		serverURL = "https://nodes.pm/ping"
	}

	req, err := http.NewRequest("GET", serverURL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("node", token)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erreur réseau : %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("token invalide")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit atteint, attendez 1 minute")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur serveur : %d", resp.StatusCode)
	}

	var nodes []NodeInfo
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, fmt.Errorf("erreur JSON : %w", err)
	}

	return nodes, nil
}
