package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"go4.org/netipx"
)

const (
	// see more: http://aka.ms/ipurlws
	apiURL = "https://endpoints.office.com/endpoints/worldwide?clientRequestId="

	all             = "all"
	timestampFormat = "20060102_150405"

	// allowlist stuff
	fileStart       = "AllowedIPs = "
	ipv4RangeString = "0.0.0.0/0"
	ipv6RangeString = "::/0"
)

// thanks https://transform.tools/json-to-go
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
	var filter string
	flag.StringVar(&filter, "f", all, "Service filter: 'Exchange', 'Skype' (inc. Teams), 'SharePoint', 'Common'. Leave blank for all services.")
	flag.Parse()

	if filter != all &&
		filter != "Exchange" &&
		filter != "Skype" &&
		filter != "SharePoint" &&
		filter != "Common" {
		log.Printf("Unrecognised filter %q\n", filter)
		flag.Usage()
		os.Exit(-1)
	}

	// get the data from the endpoint
	data, err := getData()

	if err != nil {
		log.Fatalf("Failed to get data: %s\n", err.Error())
	}

	// parse the received data
	log.Println("Parsing data")
	ips, err := parseData(data, filter)

	if err != nil {
		log.Fatalf("Failed to parse data: %q\n", err.Error())
	}

	// write the discovered routes to file
	err = writeRoutesToFile(ips, filter)

	if err != nil {
		log.Fatalf("Failed to write routes file: %s\n", err.Error())
	}

	// write a Wireguard allow list from the discovered routes
	err = outputWGAllowList(ips, filter)

	if err != nil {
		log.Fatalf("Failed to write Wireguard file: %s\n", err.Error())
	}

	// all done
	log.Println("Fin.")
}

func getData() ([]byte, error) {
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

func parseData(data []byte, filter string) (map[string]bool, error) {
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

func writeRoutesToFile(ips map[string]bool, filter string) error {
	// create the file
	filename := fmt.Sprintf("%s_m365_routes_%s.txt", time.Now().Format(timestampFormat), filter)
	f, err := os.Create(filename)

	if err != nil {
		return fmt.Errorf("failed to create file %q: %s", filename, err.Error())
	}

	defer f.Close()

	// write the IPs to file
	w := bufio.NewWriter(f)

	for ip := range ips {
		w.WriteString(fmt.Sprintf("%s, ", ip))
	}

	// tidy up
	w.Flush()
	f.Close()
	log.Printf("Wrote addresses to %q\n", filename)

	return nil
}

func outputWGAllowList(ips map[string]bool, filter string) error {
	// make the file
	filename := fmt.Sprintf("%s_wireguard_allowList_%s.txt", time.Now().Format(timestampFormat), filter)
	outFile, err := os.Create(filename)

	if err != nil {
		return fmt.Errorf("failed to make output file: %s", err.Error())
	}

	defer outFile.Close()
	outFile.WriteString(fileStart)

	// build the IP set
	var b netipx.IPSetBuilder

	// start with the whole v4 and v6 address space
	b.AddPrefix(netip.MustParsePrefix(ipv4RangeString))
	b.AddPrefix(netip.MustParsePrefix(ipv6RangeString))

	// for each IP range we got
	for ip := range ips {
		// parse it as a prefix
		prefix, err := netip.ParsePrefix(ip)

		if err != nil {
			log.Printf("Error parsing input %q as range: %s. Skipping it...", ip, err.Error())
		} else {
			// remove this range from our set
			b.RemovePrefix(prefix)
		}
	}

	// build the set
	s, _ := b.IPSet()

	// collate them all as strings
	var prefixStrings []string

	for _, r := range s.Ranges() {
		prefixes := r.Prefixes()

		for _, p := range prefixes {
			prefixStrings = append(prefixStrings, p.String())
		}
	}

	// write out the collated prefixes
	outFile.WriteString(strings.Join(prefixStrings, ", "))
	log.Printf("Wrote wireguard allowlist to %q\n", filename)

	return nil
}
