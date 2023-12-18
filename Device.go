package onvif

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/icholy/digest"
	"github.com/mydragonfly00/onvif/device"
	"github.com/mydragonfly00/onvif/gosoap"
	wsdiscovery "github.com/mydragonfly00/onvif/ws-discovery"
)

//Xlmns XML Scheam
var Xlmns = map[string]string{
	"onvif":   "http://www.onvif.org/ver10/schema",
	"tds":     "http://www.onvif.org/ver10/device/wsdl",
	"trt":     "http://www.onvif.org/ver10/media/wsdl",
	"tev":     "http://www.onvif.org/ver10/events/wsdl",
	"tptz":    "http://www.onvif.org/ver20/ptz/wsdl",
	"timg":    "http://www.onvif.org/ver20/imaging/wsdl",
	"tan":     "http://www.onvif.org/ver20/analytics/wsdl",
	"xmime":   "http://www.w3.org/2005/05/xmlmime",
	"wsnt":    "http://docs.oasis-open.org/wsn/b-2",
	"xop":     "http://www.w3.org/2004/08/xop/include",
	"wsa":     "http://www.w3.org/2005/08/addressing",
	"wstop":   "http://docs.oasis-open.org/wsn/t-1",
	"wsntw":   "http://docs.oasis-open.org/wsn/bw-2",
	"wsrf-rw": "http://docs.oasis-open.org/wsrf/rw-2",
	"wsaw":    "http://www.w3.org/2006/05/addressing/wsdl",
}

//DeviceType alias for int
type DeviceType int

// Onvif Device Tyoe
const (
	NVD DeviceType = iota
	NVS
	NVA
	NVT
	ContentType = "Content-Type"
)

type AuthMode int

// ONVIF authentication modes
const (
	AuthModeNone AuthMode = iota
	AuthModeDigest
	AuthModeWSSecurity
)

func (devType DeviceType) String() string {
	stringRepresentation := []string{
		"NetworkVideoDisplay",
		"NetworkVideoStorage",
		"NetworkVideoAnalytics",
		"NetworkVideoTransmitter",
	}
	i := uint8(devType)
	switch {
	case i <= uint8(NVT):
		return stringRepresentation[i]
	default:
		return strconv.Itoa(int(i))
	}
}

//DeviceInfo struct contains general information about ONVIF device
type DeviceInfo struct {
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareId      string
}

//Device for a new device of onvif and DeviceInfo
//struct represents an abstract ONVIF device.
//It contains methods, which helps to communicate with ONVIF device
type Device struct {
	Params    DeviceParams
	endpoints map[string]string
	info      DeviceInfo
}

type DeviceParams struct {
	AuthMode
	Xaddr      string
	Username   string
	Password   string
	HttpClient *http.Client
}

//GetServices return available endpoints
func (dev *Device) GetServices() map[string]string {
	return dev.endpoints
}

//GetServices return available endpoints
func (dev *Device) GetDeviceInfo() DeviceInfo {
	return dev.info
}

func readResponse(resp *http.Response) string {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(b)
}

//GetAvailableDevicesAtSpecificEthernetInterface ...
func GetAvailableDevicesAtSpecificEthernetInterface(interfaceName string) ([]Device, error) {
	// Call a ws-discovery Probe Message to Discover NVT type Devices
	devices, err := wsdiscovery.SendProbe(interfaceName, nil, []string{"dn:" + NVT.String()}, map[string]string{"dn": "http://www.onvif.org/ver10/network/wsdl"})
	if err != nil {
		return nil, err
	}

	nvtDevicesSeen := make(map[string]bool)
	nvtDevices := make([]Device, 0)

	for _, j := range devices {
		doc := etree.NewDocument()
		if err := doc.ReadFromString(j); err != nil {
			return nil, err
		}

		for _, xaddr := range doc.Root().FindElements("./Body/ProbeMatches/ProbeMatch/XAddrs") {
			xaddr := strings.Split(strings.Split(xaddr.Text(), " ")[0], "/")[2]
			if !nvtDevicesSeen[xaddr] {
				dev, err := NewDevice(DeviceParams{Xaddr: strings.Split(xaddr, " ")[0]})
				if err != nil {
					// TODO(jfsmig) print a warning
				} else {
					nvtDevicesSeen[xaddr] = true
					nvtDevices = append(nvtDevices, *dev)
				}
			}
		}
	}

	return nvtDevices, nil
}

func (dev *Device) getSupportedServices(resp *http.Response) error {
	doc := etree.NewDocument()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	resp.Body.Close()

	if err := doc.ReadFromBytes(data); err != nil {
		//log.Println(err.Error())
		return err
	}

	services := doc.FindElements("./Envelope/Body/GetCapabilitiesResponse/Capabilities/*/XAddr")
	for _, j := range services {
		dev.addEndpoint(j.Parent().Tag, j.Text())
	}

	extension_services := doc.FindElements("./Envelope/Body/GetCapabilitiesResponse/Capabilities/Extension/*/XAddr")
	for _, j := range extension_services {
		dev.addEndpoint(j.Parent().Tag, j.Text())
	}

	return nil
}

//NewDevice function construct a ONVIF Device entity
func NewDevice(params DeviceParams) (*Device, error) {
	dev := new(Device)
	dev.Params = params
	dev.endpoints = make(map[string]string)
	dev.addEndpoint("Device", "http://"+dev.Params.Xaddr+"/onvif/device_service")

	if dev.Params.HttpClient == nil {
		dev.Params.HttpClient = new(http.Client)
		dev.Params.HttpClient.Timeout = time.Minute * 1
	}

	getCapabilities := device.GetCapabilities{Category: "All"}

	resp, err := dev.CallMethod(getCapabilities)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("camera is not available at " + dev.Params.Xaddr + " or it does not support ONVIF services")
	}

	err = dev.getSupportedServices(resp)
	if err != nil {
		return nil, err
	}

	return dev, nil
}

func (dev *Device) addEndpoint(Key, Value string) {
	//use lowCaseKey
	//make key having ability to handle Mixed Case for Different vendor devcie (e.g. Events EVENTS, events)
	lowCaseKey := strings.ToLower(Key)

	// Replace host with host from device params.
	if u, err := url.Parse(Value); err == nil {
		u.Host = dev.Params.Xaddr
		Value = u.String()
	}

	dev.endpoints[lowCaseKey] = Value
}

//GetEndpoint returns specific ONVIF service endpoint address
func (dev *Device) GetEndpoint(name string) string {
	return dev.endpoints[name]
}

func (dev Device) buildMethodSOAP(msg string) (gosoap.SoapMessage, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(msg); err != nil {
		//log.Println("Got error")

		return "", err
	}
	element := doc.Root()

	soap := gosoap.NewEmptySOAP()
	soap.AddBodyContent(element)

	return soap, nil
}

//getEndpoint functions get the target service endpoint in a better way
func (dev Device) getEndpoint(endpoint string) (string, error) {

	// common condition, endpointMark in map we use this.
	if endpointURL, bFound := dev.endpoints[endpoint]; bFound {
		return endpointURL, nil
	}

	//but ,if we have endpoint like event、analytic
	//and sametime the Targetkey like : events、analytics
	//we use fuzzy way to find the best match url
	var endpointURL string
	for targetKey := range dev.endpoints {
		if strings.Contains(targetKey, endpoint) {
			endpointURL = dev.endpoints[targetKey]
			return endpointURL, nil
		}
	}
	return endpointURL, errors.New("target endpoint service not found")
}

//CallMethod functions call an method, defined <method> struct.
//You should use Authenticate method to call authorized requests.
func (dev Device) CallMethod(method interface{}) (*http.Response, error) {
	pkgPath := strings.Split(reflect.TypeOf(method).PkgPath(), "/")
	pkg := strings.ToLower(pkgPath[len(pkgPath)-1])

	endpoint, err := dev.getEndpoint(pkg)
	if err != nil {
		return nil, err
	}
	return dev.callMethodDo(endpoint, method)
}

//CallMethod functions call an method, defined <method> struct with authentication data
func (dev Device) callMethodDo(endpoint string, method interface{}) (*http.Response, error) {
	output, err := xml.MarshalIndent(method, "  ", "    ")
	if err != nil {
		return nil, err
	}

	soap, err := dev.buildMethodSOAP(string(output))
	if err != nil {
		return nil, err
	}

	soap.AddRootNamespaces(Xlmns)
	soap.AddAction()

	//Auth Handling
	//if dev.Params.Username != "" && dev.Params.Password != "" {
	//	soap.AddWSSecurity(dev.Params.Username, dev.Params.Password)
	//}
	//
	//return networking.SendSoap(dev.Params.HttpClient, endpoint, soap.String())
	return dev.SendSoap(endpoint, soap)
}

func (dev Device) SendSoap(endpoint string, soap gosoap.SoapMessage) (*http.Response, error) {
	if dev.Params.HttpClient == nil {
		dev.Params.HttpClient = &http.Client{}
	}

	if dev.Params.Username != "" && dev.Params.Password != "" {
		switch dev.Params.AuthMode {
		case AuthModeNone:
		case AuthModeWSSecurity:
			soap.AddWSSecurity(dev.Params.Username, dev.Params.Password)
		case AuthModeDigest:
			if _, ok := dev.Params.HttpClient.Transport.(*digest.Transport); !ok {
				dev.Params.HttpClient.Transport = &digest.Transport{Username: dev.Params.Username, Password: dev.Params.Password}
			}
		default:
			return nil, fmt.Errorf("invalid SecurityType: %d", dev.Params.AuthMode)
		}
	}

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBufferString(soap.String()))
	if err != nil {
		return nil, fmt.Errorf("could not create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/soap+xml")

	// send request
	soapResp, err := dev.Params.HttpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("could not POST request: %w", err)
	}
	//defer soapResp.Body.Close()

	if soapResp.StatusCode == http.StatusUnauthorized {
		if dev.Params.AuthMode != AuthModeDigest && dev.Params.Username != "" && dev.Params.Password != "" {
			dev.Params.AuthMode = AuthModeDigest
			if _, ok := dev.Params.HttpClient.Transport.(*digest.Transport); !ok {
				// replay request to save digest headers
				d := &digest.Transport{Transport: &fakeTransport{resp: soapResp}, Username: dev.Params.Username, Password: dev.Params.Password}
				d.RoundTrip(httpReq)
				d.Transport = nil
				dev.Params.HttpClient.Transport = d
			}
			return dev.SendSoap(endpoint, soap)
		}
		return nil, errors.New(soapResp.Status)
	}

	return soapResp, nil
}

type fakeTransport struct {
	resp *http.Response
}

func (f *fakeTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f.resp, nil
}
