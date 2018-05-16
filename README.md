# Goonvif
Simple management of IP-devices, including cameras. Goonvif is an implementation of the ONVIF protocol for managing IP devices. The purpose of this library is convenient and easy management of IP cameras and other devices that support the ONVIF standard.

## Installation
To install the library,  use **go get**:
```
go get github.com/yakovlevdmv/goonvif
```
## Supported services
The following services are fully implemented:
- Device
- Media
- PTZ
- Imaging

## Using

### General concept
1) Connecting to the device
2) Authentication (if necessary)
3) Defining Data Types
4) Carrying out the required method

#### Connecting to the device
If there is a device on the network at the address *192.168.13.42*, and its ONVIF services use the *1234* port, then you can connect to the device in the following way:
```
dev, err := goonvif.NewDevice("192.168.13.42:1234")
```

*The ONVIF port may differ depending on the device and to find out which port to use, you can go to the web interface of the device. **Usually this is 80 port.***

#### Authentication
If any function of one of the ONVIF services requires authentication, you must use the `Authenticate` method.
```
device := onvif.NewDevice("192.168.13.42:1234")
device.Authenticate("username", "password")
```

#### Defining Data Types
Each ONVIF service in this library has its own package, in which all data types of this service are defined, and the package name is identical to the service name and begins with a capital letter. 41 Goonvif defines the structures for each function of each ONVIF service supported by this library. 42 Define the data type of the function `GetCapabilities` of the` Device` service. This is done as follows:
```
capabilities := Device.GetCapabilities{Category:"All"}
```
Why does the GetCapabilities structure have the Category field and why is the value of this field All?

The figure below shows the documentation for the [GetCapabilities](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl). It can be seen that the function takes one Category parameter and its value should be one of the following: 'All', 'Analytics',' Device ',' Events', 'Imaging', 'Media' or 'PTZ'`.

![Device GetCapabilities](img/exmp_GetCapabilities.png)

An example of defining the data type of the GetServiceCapabilities function in [PTZ](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl):
```
ptzCapabilities := PTZ.GetServiceCapabilities{}
```
The figure below shows that GetServiceCapabilities does not accept any arguments.

![PTZ GetServiceCapabilities](img/GetServiceCapabilities.png)

*Common data types are in the xsd / onvif package. The types of data (structures) that can be shared by all services are defined in the onvif package.*

An example of how to define the data type of the CreateUsers function in [Devicemgmt](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl):
```
createUsers := Device.CreateUsers{User: onvif.User{Username:"admin", Password:"qwerty", UserLevel:"User"}}
```

The figure below shows that in this example, the CreateUsers structure field must be a User whose data type is the User structure containing the Username, Password, UserLevel, and optional Extension fields. The User structure is in the onvif package.

![Device CreateUsers](img/exmp_CreateUsers.png)

#### Carrying out the required method
To perform any function of one of the ONVIF services whose structure has been defined, you must use the `CallMethod` of the device object.
```
createUsers := Device.CreateUsers{User: onvif.User{Username:"admin", Password:"qwerty", UserLevel:"User"}}
device := onvif.NewDevice("192.168.13.42:1234")
device.Authenticate("username", "password")
resp, err := dev.CallMethod(createUsers)
```