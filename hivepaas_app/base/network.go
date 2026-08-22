package base

type NetworkProtocol string

const (
	NetworkProtocolTCP NetworkProtocol = "tcp"
	NetworkProtocolUDP NetworkProtocol = "udp"
)

var AllNetworkProtocols = []NetworkProtocol{NetworkProtocolTCP, NetworkProtocolUDP}
