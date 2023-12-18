package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/beevik/etree"
	"github.com/cihub/seelog"
	goonvif "github.com/mydragonfly00/onvif"
	"github.com/mydragonfly00/onvif/device"
	"github.com/mydragonfly00/onvif/gosoap"
	"github.com/mydragonfly00/onvif/media"
	sdk "github.com/mydragonfly00/onvif/sdk/device"
	"github.com/mydragonfly00/onvif/xsd/onvif"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	login    = "admin"
	password = "Supervisor"
)

func main1() {
	ctx := context.Background()

	//Getting an camera instance
	dev, err := goonvif.NewDevice(goonvif.DeviceParams{
		Xaddr:      "192.168.13.14:80",
		Username:   login,
		Password:   password,
		HttpClient: new(http.Client),
	})
	if err != nil {
		panic(err)
	}

	//Preparing commands
	systemDateAndTyme := device.GetSystemDateAndTime{}
	getCapabilities := device.GetCapabilities{Category: "All"}
	createUser := device.CreateUsers{
		User: onvif.User{
			Username:  "TestUser",
			Password:  "TestPassword",
			UserLevel: "User",
		},
	}

	//Commands execution
	systemDateAndTymeResponse, err := sdk.Call_GetSystemDateAndTime(ctx, dev, systemDateAndTyme)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(systemDateAndTymeResponse)
	}
	getCapabilitiesResponse, err := sdk.Call_GetCapabilities(ctx, dev, getCapabilities)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(getCapabilitiesResponse)
	}

	createUserResponse, err := sdk.Call_CreateUsers(ctx, dev, createUser)
	if err != nil {
		log.Println(err)
	} else {
		// You could use https://github.com/use-go/onvif/gosoap for pretty printing response
		fmt.Println(createUserResponse)
	}

}
func main() {
	fmt.Println(2333)
	_, rtsp, err_rtsp := GetRtsp("192.168.25.24", "admin", "rza2@123", "VideoSourceToken")
	//_, rtsp, err_rtsp := GetRtsp("192.168.125.157", "admin", "rza2@123", "VideoSourceToken")
	//_, rtsp, err_rtsp := tools.GetRtsp("192.168.128.142", "admin", "rza2@123", "VideoSourceToken")
	if err_rtsp != nil {
		fmt.Println(222, err_rtsp)
	}
	fmt.Println(rtsp)
}

func GetRtsp(ip, account, password, videoSourceConfiguration string) (*GoOnvifClient, string, error) {
	//初始化一个客户端
	var client *GoOnvifClient
	client = &GoOnvifClient{}

	//摄像头ip端口
	dev, err := goonvif.NewDevice(goonvif.DeviceParams{
		//Xaddr:    "192.168.25.24", // BOSCH
		Xaddr:    ip, // BOSCH
		Username: account,
		Password: password,
		//AuthMode: goonvif.AuthModeWSSecurity,
		AuthMode: goonvif.AuthModeDigest,
	})
	if err != nil {
		seelog.Error("初始化摄像头onvif客户端失败:", err)
		return nil, "", err
	}
	//账号与密码
	//dev.Authenticate("username", "password")
	//赋值给客户端
	client.Dev = *dev
	client.IsPrintRespSoap = true
	////获取token
	//res := client.GetProfileToken()
	//if res.Code != 200 {
	//	seelog.Error(ip, "获取profile token失败", res.Info)
	//	return client, "", errors.New("获取profile token失败")
	//}
	////筛选合适的token
	//res = client.SelectLocalProfileToken()
	//if res.Code != 200 {
	//	seelog.Error(ip, "查找合适的profile token失败", res.Info)
	//	return client, "", errors.New("查找合适的profile token失败")
	//}
	res := client.GetChannels()
	if res.Code != 200 {
		fmt.Println(777, res.Info)
		seelog.Error(ip, "获取通道数量失败：", err)
		return client, "", errors.New(ip + " 获取通道数量失败")
	}
	//从map中，选择profilesToken

	if _, ok := client.Channels[videoSourceConfiguration]; ok {
		value := client.Channels[videoSourceConfiguration]
		for _, v := range value {
			client.LocalSelectProfileToken = v
			if client.LocalSelectProfileToken != "" {
				break
			}
		}
	} else {
		seelog.Error(ip, "获取profile token失败", res.Info)
		return client, "", errors.New("获取profile token失败")
	}
	//获取rtsp流
	returnInfo := client.GetStreamUri()
	var rtspUrl string
	if returnInfo.Code != 200 {
		seelog.Error(ip, "rtsp视频流获取失败")
		return client, "", errors.New(ip + "rtsp视频流获取失败")
	} else {
		rtspUrl = returnInfo.Info
		if account != "" {
			index1 := strings.Index(rtspUrl, "rtsp://")
			rtspUrl = "rtsp://" + account + ":" + password + "@" + rtspUrl[index1+7:]
		}
		//seelog.Info("rtsp视频流：", rtspUrl)
		return client, rtspUrl, nil
	}
}

type PresetInfo struct {
	//要获取的预置点的token，字符串"1"-"300"
	queryPresetToken string
	//通过搜索预置点token获取到的预置点的名称
	QueryPresetNameRes string
	//通过搜索获取到的预置点的坐标值X
	QueryXRes string
	//通过搜索获取到的预置点的坐标值X
	QueryYRes string
	//获取到的预置点的光圈缩放值
	QueryPresetZoomRes string
	//要设置的预置点名称
	setPresetName string
	//要设置的预置点token
	setPresetToken string
	//要到达的预置点token
	gotoPresetToken string
	//要删除的预置点token
	removePresetToken string
}

//设备信息
type DeviceInfo struct {
	//厂家信息
	Manufacturer string
	//设备类型
	Model string
	//固件版本
	FirmwareVersion string
	//设备序列号
	SerialNumber string
	//固件ID
	HardwareId string
}

type NetWorkConfigInfo struct {
	//是否进行IPV4网络配置
	EnableIPV4NetworkConfig bool
	//要配置的IPV4地址，不包括端口
	IPV4Address string
	//要配置的IPV4地址前缀长度，默认可设置为24
	IPV4PrefixLen int
	//是否进行IPV6网络配置
	EnableIPV6NetworkConfig bool
	//要配置的IPV6地址，例如：0:0:0:0:0:0:0:0
	IPV6Address string
	//要配置的IPV6地址前缀长度，默认可设置为120
	IPV6PrefixLen int
}

type GoOnvifClient struct {
	//搜索到的设备ip端口的字符串数组
	Devices []goonvif.Device
	//获取到的设备信息
	DevInfo DeviceInfo
	//选择要筛选的设备的信息，提供设备类型或者设备序列号任意信息即可
	SelectDevInfo DeviceInfo
	//选择要进行onvif交互的设备，从搜索到的设备中选择一个
	Dev goonvif.Device
	//是否打印获取到的response中的soap信息
	IsPrintRespSoap bool
	//需要获取的模块的能力，包括PTZ云台控制、Media流媒体控制等，一般直接选择All即可
	CateGoryName string
	//鉴权认证的用户名
	LoginUserName string
	//鉴权认证使用的密码
	LoginPassWord string
	//需要创建的用户名
	CreateUserName string
	//需要创建用户对应的密码
	CreateUserPassWord string
	//创建的用户的等级，包括 Administrator、Operator、User、Anonymous、Extended
	CreateUserLevel string
	//网络接口的token，通过获取网络接口获取，设置网络时需要使用
	NetworkInterfaceToken string
	//获取到的profile token，不同码流的token不同，一般会有三种码流
	ProfilesToken []string
	//选中要获取的流的token以及要进行PTZ的token，一般选择获取到的profiles token的一个
	LocalSelectProfileToken string
	//获取ConfigurationToken
	VideoSourceConfiguration []string
	//选中要获取的OSD的token，一般选择第一个
	LocalSelectConfigurationToken string
	//获取OSDToken
	OSDToken []string
	//用来存放通道与对应的token
	Channels map[string][]string
	//ptz控制的速度
	PtzSpeed float64
	////ptz控制的方向：1：上；2：下；3：左；4：右；5：左上；6：左下；7：右上；8：右下；9：停
	//Direction PTZDirection
	////ptz移动模式，1：连续移动，0：断续移动，连续移动时选择方向不停的话则会一直移动，否则移动一次后会停下来
	//PtzMoveMode MoveMode
	//预置点信息
	PresetInfo PresetInfo
	//要配置的网络信息
	NetWorkConfigInfo NetWorkConfigInfo
}

type Code int32

const (
	OK                   Code = 200
	SearchErr            Code = -1
	ConnectErr           Code = -2
	GetTimeErr           Code = -10
	CreateUserErr        Code = -20
	GetProfilesErr       Code = -30
	GetConfigurationsErr Code = -31
	GetOSDsErr           Code = -32
	GetOSDErr            Code = -33
	GetStreamUriErr      Code = -40
	PTZErr               Code = -50
	GetDeviceInfoErr     Code = -60
	SetPresetErr         Code = -70
	GotoPresetErr        Code = -71
	RemovePresetErr      Code = -72
	GetNetWorkInfoErr    Code = -80
	GetSnapShotUriErr    Code = -100
)

type returnInfo struct {
	//状态码
	Code Code
	//错误或者返回信息
	Info string
}

//搜索设备，返回搜索到的设备列表
//func (client *GoOnvifClient) SearchDevice() returnInfo {
//	devices, _ := goonvif.GetAvailableDevicesAtSpecificEthernetInterface("eth0")
//	if devices == nil {
//		return returnInfo{SearchErr, "search devices failed."}
//	}
//	client.Devices = devices
//	return returnInfo{OK, "search device success."}
//}

//读取response结果
func readResponse(resp *http.Response) string {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

//进行onvif客户端请求发送和response读取处理
func (client *GoOnvifClient) SendReqGetResp(errCode Code, method interface{}) returnInfo {
	resp, err := client.Dev.CallMethod(method)
	message := ""
	if err != nil {
		return returnInfo{errCode, err.Error()}
	} else {
		defer func() {
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		}()
		message = readResponse(resp)
		if client.IsPrintRespSoap {
			fmt.Println(gosoap.SoapMessage(message).StringIndent())
		}
	}
	return returnInfo{OK, gosoap.SoapMessage(message).StringIndent()}
}

//获取设备信息
func (client *GoOnvifClient) GetDeviceInfo() returnInfo {
	getDevInfoReq := device.GetDeviceInformation{}
	res := client.SendReqGetResp(GetDeviceInfoErr, getDevInfoReq)
	return client.getDeviceInfoFromXml(res.Info)
}

//从xml文件中读取设备信息
func (client *GoOnvifClient) getDeviceInfoFromXml(message string) returnInfo {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(message); err != nil {
		return returnInfo{GetDeviceInfoErr, "read device xml info failed."}
	}
	root := doc.SelectElement("Envelope")
	if root == nil {
		return returnInfo{GetDeviceInfoErr, "read device xml info failed."}
	}
	modelNode := root.FindElement("./Body/GetDeviceInformationResponse/Model")
	if modelNode != nil {
		client.DevInfo.Model = modelNode.Text()
	}
	SNNode := root.FindElement("./Body/GetDeviceInformationResponse/SerialNumber")
	if SNNode != nil {
		client.DevInfo.SerialNumber = SNNode.Text()
	}
	if client.DevInfo.SerialNumber == "" && client.DevInfo.Model == "" {
		return returnInfo{GetNetWorkInfoErr, "read network xml info failed."}
	}
	return returnInfo{OK, "get device info success."}
}

//调用获取网络信息接口获取网络接口token
func (client *GoOnvifClient) GetNetWokToken() returnInfo {
	getNetWorkToken := device.GetNetworkInterfaces{}
	res := client.SendReqGetResp(GetNetWorkInfoErr, getNetWorkToken)
	return client.GetNetWorkTokenFromXml(res.Info)
}

//从xml中读取网络token
func (client *GoOnvifClient) GetNetWorkTokenFromXml(message string) returnInfo {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(message); err != nil {
		return returnInfo{GetNetWorkInfoErr, "read network xml info failed."}
	}
	root := doc.SelectElement("Envelope")
	if root == nil {
		return returnInfo{GetNetWorkInfoErr, "read network xml info failed."}
	}
	token := root.FindElements("./Body/GetNetworkInterfacesResponse/NetworkInterfaces")
	for _, res := range token {
		client.NetworkInterfaceToken = res.SelectAttr("token").Value
	}

	if client.NetworkInterfaceToken == "" {
		return returnInfo{GetNetWorkInfoErr, "read network xml info failed."}
	}

	return returnInfo{OK, "get network token success."}
}

//获取摄像头时间
func (client *GoOnvifClient) GetSystemDateAndTime() returnInfo {
	return client.GetSystemTime()
}

func (client *GoOnvifClient) GetSystemTime() returnInfo {
	systemDateAndTyme := device.GetSystemDateAndTime{}
	res := client.SendReqGetResp(GetTimeErr, systemDateAndTyme)
	if res.Code != OK {
		return returnInfo{GetTimeErr, res.Info}
	}
	return client.GetTimeFromXml(res.Info)
}

//解析xml，获取摄像头当前时间
func (client *GoOnvifClient) GetTimeFromXml(message string) returnInfo {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(message); err != nil {
		return returnInfo{GetTimeErr, "read xml failed."}
	}
	root := doc.SelectElement("Envelope")
	if root == nil {
		return returnInfo{GetTimeErr, "read xml failed."}
	}
	//时
	hour := root.FindElements("./Body/GetSystemDateAndTimeResponse/SystemDateAndTime/LocalDateTime/Time/Hour")
	if hour == nil {
		return returnInfo{GetTimeErr, "read xml failed."}
	}
	h := hour[0].Text()
	h = fmt.Sprintf("%02s", h)
	//分
	minute := root.FindElements("./Body/GetSystemDateAndTimeResponse/SystemDateAndTime/LocalDateTime/Time/Minute")
	if minute == nil {
		return returnInfo{GetTimeErr, "read xml failed."}
	}
	m := minute[0].Text()
	m = fmt.Sprintf("%02s", m)
	//秒
	second := root.FindElements("./Body/GetSystemDateAndTimeResponse/SystemDateAndTime/LocalDateTime/Time/Second")
	if second == nil {
		return returnInfo{GetTimeErr, "read xml failed."}
	}
	s := second[0].Text()
	s = fmt.Sprintf("%02s", s)
	//年
	year := root.FindElements("./Body/GetSystemDateAndTimeResponse/SystemDateAndTime/LocalDateTime/Date/Year")
	if year == nil {
		return returnInfo{GetTimeErr, "read xml failed."}
	}
	y := year[0].Text()
	//月
	month := root.FindElements("./Body/GetSystemDateAndTimeResponse/SystemDateAndTime/LocalDateTime/Date/Month")
	if month == nil {
		return returnInfo{GetTimeErr, "read xml failed."}
	}
	mo := month[0].Text()
	mo = fmt.Sprintf("%02s", mo)
	//日
	day := root.FindElements("./Body/GetSystemDateAndTimeResponse/SystemDateAndTime/LocalDateTime/Date/Day")
	if day == nil {
		return returnInfo{GetTimeErr, "read xml failed."}
	}
	d := day[0].Text()
	d = fmt.Sprintf("%02s", d)
	timeStr := y + "-" + mo + "-" + d + " " + h + ":" + m + ":" + s
	return returnInfo{OK, timeStr}
}

//获取profile token
func (client *GoOnvifClient) GetProfileToken() returnInfo {
	return client.GetProfiles()
}

//获取onvif的Profiles token
func (client *GoOnvifClient) GetProfiles() returnInfo {
	mediaProfilesReq := media.GetProfiles{}
	res := client.SendReqGetResp(GetProfilesErr, mediaProfilesReq)
	if res.Code != OK {
		return returnInfo{GetProfilesErr, res.Info}
	}
	return client.GetProfilesFromXml(res.Info)
}

//从xml中读取Profiles字段
func (client *GoOnvifClient) GetProfilesFromXml(message string) returnInfo {
	client.ProfilesToken = make([]string, 0)
	doc := etree.NewDocument()
	if err := doc.ReadFromString(message); err != nil {
		return returnInfo{GetProfilesErr, "read xml failed."}
	}
	root := doc.SelectElement("Envelope")
	if root == nil {
		return returnInfo{GetProfilesErr, "read xml failed."}
	}
	token := root.FindElements("./Body/GetProfilesResponse/Profiles")
	if token == nil || len(token) == 0 {
		return returnInfo{GetProfilesErr, "read xml failed."}
	}
	for _, res := range token {
		client.ProfilesToken = append(client.ProfilesToken, res.SelectAttr("token").Value)
	}

	if client.ProfilesToken[0] == "" {
		return returnInfo{GetProfilesErr, "read xml failed."}
	}

	return returnInfo{OK, "get profiles from xml success."}
}

//选择profile token
func (client *GoOnvifClient) SelectLocalProfileToken() returnInfo {
	for _, value := range client.ProfilesToken {
		client.LocalSelectProfileToken = value
		if client.LocalSelectProfileToken != "" {
			return returnInfo{OK, "select profile token ok!"}
		}
	}
	return returnInfo{GetProfilesErr, "select profile token failed!"}
}

//获取rtsp流媒体信息
func (client *GoOnvifClient) GetStreamUri() returnInfo {
	if client.LocalSelectProfileToken == "" {
		return returnInfo{GetStreamUriErr, "profile token is nil."}
	}
	return client.getStreamUri()
}

//根据选中的Profile token获取码流rtsp地址等
func (client *GoOnvifClient) getStreamUri() returnInfo {
	mediaUrlReq := media.GetStreamUri{}
	mediaUrlReq.ProfileToken = onvif.ReferenceToken(client.LocalSelectProfileToken)
	res := client.SendReqGetResp(GetStreamUriErr, mediaUrlReq)
	if res.Code != OK {
		return returnInfo{GetStreamUriErr, res.Info}
	}
	return client.GetRtspFromXml(res.Info)
}

//解析xml，获取rtsp
func (client *GoOnvifClient) GetRtspFromXml(message string) returnInfo {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(message); err != nil {
		return returnInfo{GetStreamUriErr, "read xml failed."}
	}
	root := doc.SelectElement("Envelope")
	if root == nil {
		return returnInfo{GetStreamUriErr, "read xml failed."}
	}
	url := root.FindElements("./Body/GetStreamUriResponse/MediaUri/Uri")
	if url == nil {
		return returnInfo{GetStreamUriErr, "read xml failed."}
	}
	return returnInfo{OK, url[0].Text()}
}

//获取快照的URL
func (client *GoOnvifClient) GetSnapUri() returnInfo {
	if client.LocalSelectProfileToken == "" {
		return returnInfo{GetSnapShotUriErr, "profile token is nil."}
	}
	return client.getSnapshotUri()
}

//抓怕快照
func (client *GoOnvifClient) getSnapshotUri() returnInfo {
	ptzSnapReq := media.GetSnapshotUri{
		ProfileToken: onvif.ReferenceToken(client.LocalSelectProfileToken),
	}
	return client.SendReqGetResp(GetSnapShotUriErr, ptzSnapReq)
}

//获取OSD所需的VideoSourceConfiguration
func (client *GoOnvifClient) GetConfigurationTokens() returnInfo {
	mediaProfilesReq := media.GetProfiles{}
	res := client.SendReqGetResp(GetConfigurationsErr, mediaProfilesReq)
	if res.Code != OK {
		return returnInfo{GetConfigurationsErr, res.Info}
	}
	return client.GetVideoSourceConfigurationFromXml(res.Info)
}

//解析xml，获取VideoSourceConfiguration
func (client *GoOnvifClient) GetVideoSourceConfigurationFromXml(message string) returnInfo {
	client.VideoSourceConfiguration = make([]string, 0)
	doc := etree.NewDocument()
	if err := doc.ReadFromString(message); err != nil {
		return returnInfo{GetConfigurationsErr, "read xml failed."}
	}
	root := doc.SelectElement("Envelope")
	if root == nil {
		return returnInfo{GetConfigurationsErr, "read xml failed."}
	}
	token := root.FindElements("./Body/GetProfilesResponse/Profiles/VideoSourceConfiguration")
	if token == nil || len(token) == 0 {
		return returnInfo{GetConfigurationsErr, "read xml failed."}
	}
	for _, res := range token {
		client.VideoSourceConfiguration = append(client.VideoSourceConfiguration, res.SelectAttr("token").Value)
	}

	if client.VideoSourceConfiguration[0] == "" {
		return returnInfo{GetConfigurationsErr, "read xml failed."}
	}

	return returnInfo{OK, "get VideoSourceConfiguration from xml success."}
}

//选择ConfigurationToken
func (client *GoOnvifClient) SelectLocalSelectConfigurationToken() returnInfo {
	for _, value := range client.VideoSourceConfiguration {
		client.LocalSelectConfigurationToken = value
		if client.LocalSelectConfigurationToken != "" {
			return returnInfo{OK, "select profile token ok!"}
		}
	}
	return returnInfo{GetProfilesErr, "select configurationToken failed!"}
}

//获取OSDToken
func (client *GoOnvifClient) GetOSDs() returnInfo {
	hostname := media.GetOSDs{}
	hostname.ConfigurationToken = onvif.ReferenceToken(client.LocalSelectConfigurationToken)
	res := client.SendReqGetResp(GetOSDsErr, hostname)
	if res.Code != OK {
		return returnInfo{GetOSDsErr, res.Info}
	}
	return client.GetOSDsFromXml(res.Info)
}

//解析xml，获取OSDToken
func (client *GoOnvifClient) GetOSDsFromXml(message string) returnInfo {
	client.OSDToken = make([]string, 0)
	doc := etree.NewDocument()
	if err := doc.ReadFromString(message); err != nil {
		return returnInfo{GetOSDsErr, "read xml failed."}
	}
	root := doc.SelectElement("Envelope")
	if root == nil {
		return returnInfo{GetOSDsErr, "read xml failed."}
	}

	token := root.FindElements("./Body/GetOSDsResponse/OSDs")
	if token == nil || len(token) == 0 {
		return returnInfo{GetOSDsErr, "read xml failed."}
	}
	for _, res := range token {
		client.OSDToken = append(client.OSDToken, res.SelectAttr("token").Value)
	}

	if client.OSDToken[0] == "" {
		return returnInfo{GetOSDsErr, "read xml failed."}
	}

	return returnInfo{OK, "get OSDToken from xml success."}
}

//获取OSDName
func (client *GoOnvifClient) GetOSD(OSDToken string) returnInfo {
	hostname := media.GetOSD{}
	hostname.OSDToken = onvif.ReferenceToken(OSDToken)
	res := client.SendReqGetResp(GetOSDErr, hostname)
	if res.Code != OK {
		return returnInfo{GetOSDErr, res.Info}
	}
	return client.GetOSDFromXml(res.Info)
}

//解析xml，获取name
func (client *GoOnvifClient) GetOSDFromXml(message string) returnInfo {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(message); err != nil {
		return returnInfo{GetOSDErr, "read xml failed."}
	}
	root := doc.SelectElement("Envelope")
	if root == nil {
		return returnInfo{GetOSDErr, "read xml failed."}
	}

	res := root.FindElements("./Body/GetOSDResponse/OSD/TextString/Extension/ChannelName")
	if res == nil {
		return returnInfo{GetOSDErr, "read xml failed."}
	}
	text := res[0].Text()
	if text == "true" {
		res := root.FindElements("./Body/GetOSDResponse/OSD/TextString/PlainText")
		if res == nil {
			return returnInfo{GetOSDErr, "read xml failed."}
		}
		return returnInfo{OK, res[0].Text()}
	}
	return returnInfo{GetOSDErr, "read xml failed."}
}

//获取通道数量，并将token进行分类 map[osdToken:[]profileToken]
func (client *GoOnvifClient) GetChannels() returnInfo {
	mediaProfilesReq := media.GetProfiles{}

	res := client.SendReqGetResp(GetProfilesErr, mediaProfilesReq)
	if res.Code != OK {
		return returnInfo{GetProfilesErr, res.Info}
	}

	return client.GetChannelsFromXml(res.Info)
}

func saveFile(content string) {
	filename := "123.txt"
	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		fmt.Println("saveFile err :", err.Error())
	} else {
		_, err = f.Write([]byte(content))
		if err != nil {
			fmt.Println("saveFile err f.Write :", err.Error())
		}
	}
}

//解析xml，获取对应的通道map
func (client *GoOnvifClient) GetChannelsFromXml(message string) returnInfo {
	client.Channels = make(map[string][]string)
	doc := etree.NewDocument()
	if err := doc.ReadFromString(message); err != nil {
		return returnInfo{GetProfilesErr, "read xml failed1."}
	}
	root := doc.SelectElement("Envelope")
	if root == nil {
		return returnInfo{GetProfilesErr, "read xml failed."}
	}
	token := root.FindElements("./Body/GetProfilesResponse/Profiles")
	if token == nil || len(token) == 0 {
		return returnInfo{GetProfilesErr, "read xml failed2."}
	}

	for _, v := range token {
		res := v.FindElement("./VideoSourceConfiguration")
		if res == nil {
			return returnInfo{GetProfilesErr, "read xml failed3."}
		}
		value := res.Attr[0].Value
		client.Channels[value] = append(client.Channels[value], v.SelectAttr("token").Value)
	}

	return returnInfo{OK, "get channels from xml success."}
}
