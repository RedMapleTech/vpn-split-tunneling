# Overview
A quick Go tool to grab the IPs for m365 services and print out a unique list, for calculating split tunnel configuration that doesn't push e.g. Teams down the VPN.

# Example for Wireguard
We grab the data from, e.g. `https://endpoints.office.com/endpoints/worldwide?clientRequestId=e99b4f20-3062-4915-9c3a-17ae79e4ba65` (the request ID is just a client UUID).

This gives us [ms_routes.json](./ms_routes.json).

Our [main.go](main.go) grabs this, unmarshals the JSON, and creates a map of all the unique IPs for all the entries. And writes them to [ms_routes.txt](./ms_routes.txt).

You can add this to [Allowed IPs calculator](https://www.procustodibus.com/blog/2021/03/wireguard-allowedips-calculator/) to create an allow/blocklist:

![WIREGUARD ALLOWEDIPS CALCULATOR](assets/2024-06-14-17-08-25.png)

This gives us [allowed.txt](./allowed.txt).
