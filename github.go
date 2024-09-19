package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

const (
	githubURL = "https://api.github.com/meta"
)

type GitHubData struct {
	VerifiablePasswordAuthentication bool `json:"verifiable_password_authentication"`
	SSHKeyFingerprints               struct {
		SHA256ECDSA   string `json:"SHA256_ECDSA"`
		SHA256ED25519 string `json:"SHA256_ED25519"`
		SHA256RSA     string `json:"SHA256_RSA"`
	} `json:"ssh_key_fingerprints"`
	SSHKeys                  []string `json:"ssh_keys"`
	Hooks                    []string `json:"hooks"`
	Web                      []string `json:"web"`
	API                      []string `json:"api"`
	Git                      []string `json:"git"`
	GithubEnterpriseImporter []string `json:"github_enterprise_importer"`
	Packages                 []string `json:"packages"`
	Pages                    []string `json:"pages"`
	Importer                 []string `json:"importer"`
	Actions                  []string `json:"actions"`
	ActionsMacos             []string `json:"actions_macos"`
	Codespaces               []string `json:"codespaces"`
	Dependabot               []string `json:"dependabot"`
	Copilot                  []string `json:"copilot"`
	Domains                  struct {
		Website              []string `json:"website"`
		Codespaces           []string `json:"codespaces"`
		Copilot              []string `json:"copilot"`
		Packages             []string `json:"packages"`
		Actions              []string `json:"actions"`
		ArtifactAttestations struct {
			TrustDomain string   `json:"trust_domain"`
			Services    []string `json:"services"`
		} `json:"artifact_attestations"`
	} `json:"domains"`
}

func getGitHubIPs() (map[string]bool, error) {
	// get the data from the endpoint
	data, err := getGitHubData()

	if err != nil {
		log.Fatalf("Failed to get data: %s\n", err.Error())
	}

	// parse the received data
	log.Println("Parsing data")

	ips, err := parseGitHubData(data)

	return ips, err
}

func getGitHubData() ([]byte, error) {
	// prep the URL
	log.Printf("Getting data from %s\n", githubURL)

	// get the data
	resp, err := http.Get(githubURL)

	if err != nil {
		return nil, err
	}

	// read the data from the response body
	var content []byte
	content, err = io.ReadAll(resp.Body)
	resp.Body.Close()

	if len(content) == 0 {
		return nil, fmt.Errorf("failed to get any data")
	}

	return content, err
}

func parseGitHubData(data []byte) (map[string]bool, error) {
	// unmarshal the JSON data
	var parsed GitHubData
	err := json.Unmarshal(data, &parsed)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %q", err.Error())
	}

	// map to store all unique IPs
	ips := make(map[string]bool)

	// for all the groups in the struct
	for _, ip := range parsed.Hooks {
		address, _, err := net.ParseCIDR(ip)

		if err == nil {
			// add it to the map if we don't have it already
			_, ok := ips[address.String()]

			if !ok {
				ips[address.String()] = true
			}
		}
	}

	for _, ip := range parsed.API {
		address, _, err := net.ParseCIDR(ip)

		if err == nil {
			// add it to the map if we don't have it already
			_, ok := ips[address.String()]

			if !ok {
				ips[address.String()] = true
			}
		}
	}

	for _, ip := range parsed.Git {
		address, _, err := net.ParseCIDR(ip)

		if err == nil {
			// add it to the map if we don't have it already
			_, ok := ips[address.String()]

			if !ok {
				ips[address.String()] = true
			}
		}
	}

	for _, ip := range parsed.Pages {
		address, _, err := net.ParseCIDR(ip)

		if err == nil {
			// add it to the map if we don't have it already
			_, ok := ips[address.String()]

			if !ok {
				ips[address.String()] = true
			}
		}
	}

	for _, ip := range parsed.Codespaces {
		address, _, err := net.ParseCIDR(ip)

		if err == nil {
			// add it to the map if we don't have it already
			_, ok := ips[address.String()]

			if !ok {
				ips[address.String()] = true
			}
		}
	}

	for _, ip := range parsed.Copilot {
		address, _, err := net.ParseCIDR(ip)

		if err == nil {
			// add it to the map if we don't have it already
			_, ok := ips[address.String()]

			if !ok {
				ips[address.String()] = true
			}
		}
	}

	for _, ip := range parsed.Dependabot {
		address, _, err := net.ParseCIDR(ip)

		if err == nil {
			// add it to the map if we don't have it already
			_, ok := ips[address.String()]

			if !ok {
				ips[address.String()] = true
			}
		}
	}

	return ips, nil
}
