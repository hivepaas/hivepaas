package base

type NetworkProtocol string

const (
	NetworkProtocolHTTP NetworkProtocol = "http"
	NetworkProtocolTCP  NetworkProtocol = "tcp"
	NetworkProtocolUDP  NetworkProtocol = "udp"
)

var (
	AllNetworkProtocols = []NetworkProtocol{NetworkProtocolHTTP, NetworkProtocolTCP, NetworkProtocolUDP}
	AllRoutingProtocols = []NetworkProtocol{NetworkProtocolHTTP, NetworkProtocolTCP}
)
