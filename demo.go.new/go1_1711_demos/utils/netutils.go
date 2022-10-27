package utils

import "net"

func GetHostIpAddrs() ([]string, []string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, nil, err
	}

	var (
		localIPV4    = make([]string, 0, 2)
		nonLocalIPV4 = make([]string, 0, 2)
	)
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
			if ipNet.IP.IsLoopback() {
				localIPV4 = append(localIPV4, ipNet.IP.String())
			} else {
				nonLocalIPV4 = append(nonLocalIPV4, ipNet.IP.String())
			}
		}
	}

	return localIPV4, nonLocalIPV4, nil
}
