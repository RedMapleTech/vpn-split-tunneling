# Overview
A quick Go tool to grab the IPs for Microsoft services (M365 and GitHub) and print out a unique list, for calculating split tunnel VPN configuration that doesn't push e.g. Teams down the VPN. Now also outputs a Wireguard formatted allowlist.

Uses the data published on [endpoints.office.com](https://endpoints.office.com/) - see [here](http://aka.ms/ipurlws) for documentation. Example data is in [m365_routes.json](./examples/m365_routes.json).

# Running
Our [main.go](main.go) grabs the data from the endpoint, unmarshals the JSON, and creates a map of all the unique IPs for all the entries. It them writes them to a text file of comma separated values.

You can filter with one of `Exchange`, `Skype`, `SharePoint` or `Common`. Leave blank for all. `Skype` means `Skype and Teams`.

```shell
go run *.go -f Common
2024/06/17 09:37:18 Getting data from https://endpoints.office.com/endpoints/worldwide?clientRequestId=b9dd4f30-2d30-42c1-8ada-cc26221351e2
2024/06/17 09:37:18 Parsing data
        Microsoft 365 Common and Office Online 46: 20 IPs
        Microsoft 365 Common and Office Online 56: 16 IPs
        Microsoft 365 Common and Office Online 64: 5 IPs
2024/06/17 09:37:18 Wrote addresses to "20240617_093718_m365_routes_Common.txt"
2024/06/17 09:37:18 Fin.
```

See the [examples](./examples/).

## Now with Wireguard output
Following the example from [this site](https://www.procustodibus.com/blog/2021/03/wireguard-allowedips-calculator/), it now outputs a Wireguard allowlist. This is the entire IPv4 and v6 address ranges, minus any addresses discovered from Microsoft.

For example, a Wireguard allowlist that blocks all the addresses returned for Teams/Skype:

```
2024/10/22 16:25:30 Getting data from https://endpoints.office.com/endpoints/worldwide?clientRequestId=79c715a9-5a31-4d01-83d9-23295af0857f
2024/10/22 16:25:31 Parsing data
        Microsoft Teams 11: 3 IPs
        Microsoft Teams 12: 11 IPs
2024/10/22 16:25:31 Got 11 IP addresses from Microsoft
2024/10/22 16:25:31 Wrote addresses to "20241022_162531_m365_routes_Skype.txt"
2024/10/22 16:25:31 Getting data from https://api.github.com/meta
2024/10/22 16:25:31 Parsing data
2024/10/22 16:25:31 Got 103 IP addresses from GitHub
2024/10/22 16:25:31 Wrote addresses to "20241022_162531_m365_routes_GitHub.txt"
2024/10/22 16:25:31 Wrote wireguard allowlist to "20241022_162531_wireguard_allowList_Skype.txt"
2024/10/22 16:25:31 Fin.
```

Produces this:

```
AllowedIPs = 0.0.0.0/3, 32.0.0.0/4, 48.0.0.0/6, 52.0.0.0/10, 52.64.0.0/11, 52.96.0.0/12, 52.116.0.0/14, 52.120.0.0/15, 52.124.0.0/14, 52.128.0.0/10, 52.192.0.0/11, 52.224.0.0/13, 52.232.0.0/14, 52.236.0.0/15, 52.238.0.0/18, 52.238.64.0/19, 52.238.96.0/20, 52.238.112.0/22, 52.238.116.0/23, 52.238.118.0/24, 52.238.119.0/25...
```

