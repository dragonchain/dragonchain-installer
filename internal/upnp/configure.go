package upnp

import (
	"errors"
	"net"
	"time"

	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway1"
	"github.com/huin/goupnp/dcps/internetgateway2"
)

// AddUPNPPortMapping attempts to use UPNP in order to forward a port from the NAT to this device
func AddUPNPPortMapping(port int) error {
	upnp, err := discover()
	if err != nil {
		return err
	}
	ip, err := upnp.internalAddress()
	if err != nil {
		return err
	}
	// Try and delete any existing mapping before creating it to clean it up if it already exists
	upnp.client.DeletePortMapping("", uint16(port), "TCP")
	// Attempt to create the actual mapping
	return upnp.client.AddPortMapping("", uint16(port), "TCP", uint16(port), ip.String(), true, "dragonchain", 0)
}

type upnp struct {
	device *goupnp.RootDevice
	client upnpClient
}

type upnpClient interface {
	DeletePortMapping(NewRemoteHost string, NewExternalP uint16, NewProtocol string) error
	AddPortMapping(NewRemoteHost string, NewExternalPort uint16, NewProtocol string, NewInternalPort uint16, NewInternalClient string, NewEnabled bool, NewPortMappingDescription string, NewLeaseDuration uint32) error
	GetNATRSIPStatus() (NewRSIPAvailable bool, NewNATEnabled bool, err error)
}

func (upnp *upnp) internalAddress() (net.IP, error) {
	deviceAddr, err := net.ResolveUDPAddr("udp4", upnp.device.URLBase.Host)
	if err != nil {
		return nil, err
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			if x, ok := addr.(*net.IPNet); ok && x.Contains(deviceAddr.IP) {
				return x.IP, nil
			}
		}
	}
	return nil, errors.New("Could not find local address in network " + deviceAddr.String())
}

func discover() (*upnp, error) {
	devices1, err := goupnp.DiscoverDevices(internetgateway1.URN_WANConnectionDevice_1)
	if err != nil {
		return nil, err
	}
	devices2, err := goupnp.DiscoverDevices(internetgateway2.URN_WANConnectionDevice_2)
	if err != nil {
		return nil, err
	}
	var upnpClient *upnp = nil
	for i := 0; i < 2; i++ {
		var devices []goupnp.MaybeRootDevice
		if i == 0 {
			devices = devices2
		} else {
			devices = devices1
		}
		for _, device := range devices {
			if upnpClient != nil || device.Root == nil {
				continue
			}
			device.Root.Device.VisitServices(func(service *goupnp.Service) {
				if upnpClient != nil {
					return
				}
				sc := goupnp.ServiceClient{
					SOAPClient: service.NewSOAPClient(),
					RootDevice: device.Root,
					Location:   device.Location,
					Service:    service,
				}
				sc.SOAPClient.HTTPClient.Timeout = 3 * time.Second
				if i == 0 {
					if service := checkIGDv2(sc); service != nil {
						upnpClient = &upnp{device.Root, service}
					}
				} else {
					if service := checkIGDv1(sc); service != nil {
						upnpClient = &upnp{device.Root, service}
					}
				}
				if upnpClient == nil {
					return
				}
				if _, nat, err := upnpClient.client.GetNATRSIPStatus(); err != nil || !nat {
					upnpClient = nil
					return
				}
			})
		}
	}
	if upnpClient == nil {
		return nil, errors.New("Couldn't find any UPNP compatible router")
	}
	return upnpClient, nil
}

func checkIGDv1(sc goupnp.ServiceClient) upnpClient {
	switch sc.Service.ServiceType {
	case internetgateway1.URN_WANIPConnection_1:
		return &internetgateway1.WANIPConnection1{ServiceClient: sc}
	case internetgateway1.URN_WANPPPConnection_1:
		return &internetgateway1.WANPPPConnection1{ServiceClient: sc}
	}
	return nil
}

func checkIGDv2(sc goupnp.ServiceClient) upnpClient {
	switch sc.Service.ServiceType {
	case internetgateway2.URN_WANIPConnection_1:
		return &internetgateway2.WANIPConnection1{ServiceClient: sc}
	case internetgateway2.URN_WANIPConnection_2:
		return &internetgateway2.WANIPConnection2{ServiceClient: sc}
	case internetgateway2.URN_WANPPPConnection_1:
		return &internetgateway2.WANPPPConnection1{ServiceClient: sc}
	}
	return nil
}
