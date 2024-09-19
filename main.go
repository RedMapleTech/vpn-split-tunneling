package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/netip"
	"os"
	"strings"
	"time"

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

	// get all the MS IPs
	ips, err := getMicrosoftIPs(filter)

	if err != nil {
		log.Fatalf("Failed to get IPs from Microsoft: %q\n", err.Error())
	}

	log.Printf("Got %d IP addresses from Microsoft\n", len(ips))

	// write the discovered routes to file
	err = writeRoutesToFile(ips, filter)

	if err != nil {
		log.Fatalf("Failed to write routes file: %s\n", err.Error())
	}

	// get all the GitHub IPs
	gitHubIPs, err := getGitHubIPs()

	if err != nil {
		log.Fatalf("Failed to get IPs from GitHub: %q\n", err.Error())
	}

	log.Printf("Got %d IP addresses from GitHub\n", len(gitHubIPs))

	// add them in
	for ip := range gitHubIPs {
		_, ok := ips[ip]

		if !ok {
			ips[ip] = true
		}
	}

	// write the discovered routes to file
	err = writeRoutesToFile(gitHubIPs, "GitHub")

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

func writeRoutesToFile(ips map[string]bool, fileType string) error {
	// create the file
	filename := fmt.Sprintf("%s_m365_routes_%s.txt", time.Now().Format(timestampFormat), fileType)
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
	log.Printf("Wrote addresses to %q\n", filename)

	return nil
}

// Use the IPs to calculate and output a Wireguard allowlist
// This is the entire IPv4 and v6 address ranges, minus any addresses discovered from Microsoft.
func outputWGAllowList(discoveredIPs map[string]bool, filter string) error {
	// make the file
	filename := fmt.Sprintf("%s_wireguard_allowList_%s.txt", time.Now().Format(timestampFormat), filter)
	outFile, err := os.Create(filename)

	if err != nil {
		return fmt.Errorf("failed to make output file: %s", err.Error())
	}

	defer outFile.Close()
	outFile.WriteString(fileStart)

	// build the IP set
	var ipSetBuilder netipx.IPSetBuilder

	// start with the whole v4 and v6 address space
	ipSetBuilder.AddPrefix(netip.MustParsePrefix(ipv4RangeString))
	ipSetBuilder.AddPrefix(netip.MustParsePrefix(ipv6RangeString))

	// for each IP range we got
	for ip := range discoveredIPs {
		// parse it as a prefix
		prefix, err := netip.ParsePrefix(ip)

		if err != nil {
			// try as an address
			add, err := netip.ParseAddr(ip)

			if err != nil {
				log.Printf("Error parsing input %q as range or address: %s. Skipping it...", ip, err.Error())
			} else {
				// remove this range from our set
				ipSetBuilder.Remove(add)
			}

		} else {
			// remove this range from our set
			ipSetBuilder.RemovePrefix(prefix)
		}
	}

	// build the set
	set, _ := ipSetBuilder.IPSet()

	// collate them all as strings
	var prefixStrings []string

	for _, r := range set.Ranges() {
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
