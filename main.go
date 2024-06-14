package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type routeStruct []struct {
	ID                     int      `json:"id"`
	ServiceArea            string   `json:"serviceArea"`
	ServiceAreaDisplayName string   `json:"serviceAreaDisplayName"`
	Urls                   []string `json:"urls,omitempty"`
	Ips                    []string `json:"ips,omitempty"`
	TCPPorts               string   `json:"tcpPorts,omitempty"`
	UDPPorts               string   `json:"udpPorts,omitempty"`
	ExpressRoute           bool     `json:"expressRoute"`
	Category               string   `json:"category"`
	Required               bool     `json:"required"`
	Notes                  string   `json:"notes,omitempty"`
}

func main() {
	data, err := getData()

	if err != nil {
		log.Fatalf("Failed to get data: %s\n", err.Error())
	}

	log.Println("Parsing data")
	parseData(data)

	log.Println("Fin.")
}

func getData() ([]byte, error) {
	uuid := uuid.New()
	url := fmt.Sprintf("https://endpoints.office.com/endpoints/worldwide?clientRequestId=%s", uuid.String())
	log.Printf("Getting data from %s\n", url)
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	var content []byte
	content, err = io.ReadAll(resp.Body)
	resp.Body.Close()

	return content, err
}

func parseData(data []byte) {
	var parsed routeStruct
	err := json.Unmarshal(data, &parsed)

	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %q\n", err.Error())
	}

	ips := make(map[string]bool)

	for _, route := range parsed {

		if len(route.Ips) > 0 {
			fmt.Printf("\t%s: %d IPs\n", route.ServiceAreaDisplayName, len(route.Ips))

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

	// write the IPs to file
	filename := "ms_routes.txt"
	f, err := os.Create(filename)

	if err != nil {
		log.Fatalf("Failed to create file: %q\n", err.Error())
	}

	w := bufio.NewWriter(f)

	for ip := range ips {
		w.WriteString(fmt.Sprintf("%s, ", ip))
	}

	w.Flush()
	f.Close()
	log.Printf("Wrote addresses to %q\n", filename)
}
