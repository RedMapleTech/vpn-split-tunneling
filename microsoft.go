package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func getMicrosoftIPs(filter string) (map[string]bool, error) {
	// get the data from the endpoint
	data, err := getMSData()

	if err != nil {
		log.Fatalf("Failed to get data: %s\n", err.Error())
	}

	// parse the received data
	log.Println("Parsing data")

	ips, err := parseMSData(data, filter)

	return ips, err
}

func getMSData() ([]byte, error) {
	// prep the URL
	uuid := uuid.New()
	url := fmt.Sprintf("%s%s", apiURL, uuid.String())
	log.Printf("Getting data from %s\n", url)

	// get the data
	resp, err := http.Get(url)

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

func parseMSData(data []byte, filter string) (map[string]bool, error) {
	// unmarshal the JSON data
	var parsed routeStruct
	err := json.Unmarshal(data, &parsed)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %q", err.Error())
	}

	// map to store all unique IPs
	ips := make(map[string]bool)

	// for all the routes in the struct
	for _, route := range parsed {
		// we don't care about entries without an IP
		if len(route.Ips) > 0 {
			{
				// if we're including all or it matches the filter
				if filter == all || strings.Contains(route.ServiceArea, filter) {
					fmt.Printf("\t%s %d: %d IPs\n", route.ServiceAreaDisplayName, route.ID, len(route.Ips))

					// for all the IPs it has
					for _, i := range route.Ips {

						// add it to the map if we don't have it already
						_, ok := ips[i]

						if !ok {
							ips[i] = true
						}
					}
				}
			}
		}
	}

	return ips, nil
}
