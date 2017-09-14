/*
 * Copyright © 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package createComplaint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

const (

	// Live Apps config
	tibcoAccountURL = "https://sso-ext.tibco.com:443"
	tenantID        = "BPM"
	sandboxID       = "31"

	// accountID. This is for accounts with multiple organizations
	// I dont want to confuse the connector config since it is hard to get this ID so
	// for now it has to be set here
	//accountID = "01BHEWGDNHCPGS8PYZMMGKBAMN"
	accountID = ""

	// Live Apps application config
	applicationID        = 193
	applicationCreatorID = 908
	applicationVersion   = 3
	activityID           = "Z83CGdFT4Eeec68g_fwbarg"
	applicationName      = "CustomerComplaint1"
	activityName         = "Task"
	processName          = "EnterComplaint1"
	processLabel         = "Enter Complaint"

	// params
	ivConnection     = "liveappsConnection"
	referenceParam   = "reference"
	nameParam        = "name"
	contactParam     = "contact"
	typeReqParam     = "type"
	summaryParam     = "summary"
	descriptionParam = "description"
	ovResult         = "result"
)

var username string
var password string
var region string
var liveAppsURL string

var activityLog = logger.GetLogger("tibco-activity-complaint-creator")

type CreateComplaintActivity struct {
	metadata *activity.Metadata
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &CreateComplaintActivity{metadata: metadata}
}

func (a *CreateComplaintActivity) Metadata() *activity.Metadata {
	return a.metadata
}
func (a *CreateComplaintActivity) Eval(context activity.Context) (done bool, err error) {
	activityLog.Info("Executing Live Apps Create Complaint Case activity")

	// sequence:
	// read data from activity inputs
	// authenticate and get access token
	// login to get Atmosphere session cookie
	// (Atmosphere cookie needs to be on subsequent calls)
	// run caseCreatory() and get process id
	// updateProcess to release creator task with data
	// output the process id that we get back from the updateProcess call

	// Validates that the connection has been set. The connection is mandatory
	if context.GetInput(ivConnection) == nil {
		return false, activity.NewError("Live Apps connection is not configured", "LIVEAPPS-CON-2000", nil)
	}
	activityLog.Info("Getting conn")
	connectionInfo := context.GetInput(ivConnection).(map[string]interface{})
	fmt.Println("Connection is:")
	fmt.Println(connectionInfo)
	activityLog.Info("Getting settings")
	connectionSettings := connectionInfo["settings"].([]interface{})

	// get the username password and region from the connection
	activityLog.Info("Getting connection details")
	for _, v := range connectionSettings {
		setting := v.(map[string]interface{})
		if setting["name"] == "username" {
			username = setting["value"].(string)
		} else if setting["name"] == "password" {
			password = setting["value"].(string)
		} else if setting["name"] == "region" {
			region = setting["value"].(string)
		}
	}
	activityLog.Info("Got connection details")

	if region == "eu" || region == "EU" {
		liveAppsURL = "https://eu.liveapps.cloud.tibco.com"
	} else if region == "us" || region == "US" {
		liveAppsURL = "https://liveapps.cloud.tibco.com"
	} else if region == "au" || region == "AU" {
		liveAppsURL = "https://au.liveapps.cloud.tibco.com"
	} else {
		return false, activity.NewError("Live Apps region (EU/US/AU) is invalid", "LIVEAPPS-REGION-3000", nil)
	}

	// read data from activity inputs
	reference, _ := context.GetInput(referenceParam).(int)
	name, _ := context.GetInput(nameParam).(string)
	contact, _ := context.GetInput(contactParam).(string)
	typeReq, _ := context.GetInput(typeReqParam).(string)
	summary, _ := context.GetInput(summaryParam).(string)
	desc, _ := context.GetInput(descriptionParam).(string)

	if typeReq != "Customer Service" && typeReq != "Wrong Product" && typeReq != "Faulty Product" && typeReq != "Other" && typeReq != "Billing" {
		// I do some extra checking here just to stop invalid values getting into Live Apps
		typeReq = "Other"
	}

	// ok now make the API calls
	token := getToken()
	atmosphereCookie := performLogin(token)
	id := startPF(atmosphereCookie)
	result := updatePF(atmosphereCookie, id, reference, name, contact, typeReq, summary, desc)

	// we are done so return the result from updating the pageflow
	context.SetOutput(ovResult, result)

	return true, nil
}

func getToken() string {
	method := "POST"
	uri := tibcoAccountURL + "/as/token.oauth2?username=" + username + "&password=" + password + "&client_id=ropc_ipass&grant_type=password"

	var reqBody io.Reader

	contentType := "application/json; charset=UTF-8"
	reqBody = nil
	req, err := http.NewRequest(method, uri, reqBody)
	if reqBody != nil {
		req.Header.Set("Content-Type", contentType)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		activityLog.Error(err.Error())
	}
	defer resp.Body.Close()

	type Message struct {
		AccessToken string `json:"access_token"`
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		activityLog.Error(err.Error())
	}

	var s = new(Message)
	err2 := json.Unmarshal(body, &s)

	if err2 != nil {
		activityLog.Error("Error from unmarshal Tibco Account auth response:", err2)
	}

	return s.AccessToken
}

func performLogin(token string) []*http.Cookie {
	method := "POST"
	uri := liveAppsURL + ":443/idm/v1/login-oauth"

	var reqBody io.Reader

	contentType := "application/x-www-form-urlencoded"

	data := url.Values{}
	data.Set("TenantId", tenantID)
	data.Add("AccessToken", token)
	if accountID != "" {
		data.Add("AccountId", accountID)
	}
	reqBody = bytes.NewBufferString(data.Encode())

	req, err := http.NewRequest(method, uri, reqBody)

	if reqBody != nil {
		req.Header.Set("Content-Type", contentType)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// save the cookies as these are needed for subsequent API calls
	cookie := resp.Cookies()
	return cookie
}

func startPF(cookies []*http.Cookie) string {

	method := "POST"
	uri := liveAppsURL + ":443/pageflow/start?$sandbox=" + sandboxID

	contentType := "application/json"

	type CreatorMessage struct {
		ID              int      `json:"id"`
		Name            string   `json:"name"`
		Label           string   `json:"label"`
		Version         int      `json:"version"`
		ApplicationID   int      `json:"applicationId"`
		ApplicationName string   `json:"applicationName"`
		ActivityID      string   `json:"activityId"`
		ActivityName    string   `json:"activityName"`
		Roles           []string `json:"roles"`
	}

	creator := new(CreatorMessage)

	creator.ID = applicationCreatorID
	creator.Name = processName
	creator.Label = processLabel
	creator.Version = applicationVersion
	creator.ApplicationID = applicationID
	creator.ApplicationName = applicationName
	creator.ActivityID = activityID
	creator.ActivityName = activityName
	creator.Roles = []string{}

	// create JSON representation of creator payload

	bodyStr, err1 := json.Marshal(*creator)
	if err1 != nil {
		activityLog.Error("Error creating start creator payload from input data: " + err1.Error())
	}

	req, err := http.NewRequest(method, uri, bytes.NewBuffer(bodyStr))
	req.Header.Set("Content-Type", contentType)

	cookieLen := len(cookies)
	atmosCookie := new(http.Cookie)

	// find the atmosphere session cookie in the passed cookies attach it to this request
	for i := 0; i < cookieLen; i++ {
		if cookies[i].Name == "AtmosphereSession" {
			atmosCookie = cookies[i]
		}
	}
	req.AddCookie(atmosCookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		activityLog.Error(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		activityLog.Error(err.Error())
	}

	// to understand the response we create a struct then unmarshal the JSON response to that struct
	// hint: there are various web tools that generate struct from JSON and vice versa
	type AppID struct {
		AppID          int `json:"appId"`
		SandboxID      int `json:"sandboxId"`
		SubscriptionID int `json:"subscriptionId"`
	}

	type Message struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Version       int    `json:"version"`
		ActivityName  string `json:"activityName"`
		ApplicationID AppID  `json:"applicationId"`
		OldCaseState  string `json:"oldCaseState"`
	}

	var s = new(Message)
	err2 := json.Unmarshal(body, &s)

	if err2 != nil {
		activityLog.Info("Error from unmarshal startPf response:", err2)
	}
	activityLog.Info("Case Creator Started")

	return s.ID
}

func updatePF(cookies []*http.Cookie, id string, referenceValue int, nameValue string, contactValue string, typeValue string, summaryValue string, descValue string) string {

	method := "POST"
	uri := liveAppsURL + "/pageflow/update?$sandbox=" + sandboxID

	contentType := "application/json"

	// we need to create the object for the updatePF action
	type Attrib struct {
		Op    string      `json:"op"`
		Path  string      `json:"path"`
		Rank  int         `json:"rank"`
		Value interface{} `json:"value"`
	}

	type Complaint struct {
		CustomerComplaint1 []Attrib `json:"CustomerComplaint1"`
	}

	// note: data is actually the JSON string representation of Complaint
	type Req1 struct {
		Data string `json:"data"`
		ID   string `json:"id"`
	}

	reference := new(Attrib)
	reference.Op = "add"
	reference.Path = "/Reference_v1/"
	reference.Rank = 0
	reference.Value = referenceValue

	name := new(Attrib)
	name.Op = "add"
	name.Path = "/Name_v1/"
	name.Rank = 0
	name.Value = nameValue

	contact := new(Attrib)
	contact.Op = "add"
	contact.Path = "/Contact_v1/"
	contact.Rank = 0
	contact.Value = contactValue

	type1 := new(Attrib)
	type1.Op = "add"
	type1.Path = "/Type_v1/"
	type1.Rank = 0
	type1.Value = typeValue

	summary := new(Attrib)
	summary.Op = "add"
	summary.Path = "/Summary_v1/"
	summary.Rank = 0
	summary.Value = summaryValue

	desc := new(Attrib)
	desc.Op = "add"
	desc.Path = "/Description_v1/"
	desc.Rank = 0
	desc.Value = descValue

	comp := new(Complaint)

	comp.CustomerComplaint1 = make([]Attrib, 6)

	comp.CustomerComplaint1[0] = *reference
	comp.CustomerComplaint1[1] = *name
	comp.CustomerComplaint1[2] = *contact
	comp.CustomerComplaint1[3] = *type1
	comp.CustomerComplaint1[4] = *summary
	comp.CustomerComplaint1[5] = *desc

	requestObj := new(Req1)

	// Data is actually a JSON string inside this JSON payload
	datastr, err1 := json.Marshal(*comp)
	if err1 != nil {
		activityLog.Error("Error creating complaint from input data: " + err1.Error())
	}

	requestObj.Data = bytes.NewBuffer(datastr).String()
	requestObj.ID = id

	jsonstr, err := json.Marshal(requestObj)
	if err != nil {
		activityLog.Error("Error creating complaint json string: " + err.Error())
	}

	bodyStr := jsonstr

	req, err := http.NewRequest(method, uri, bytes.NewBuffer(bodyStr))
	req.Header.Set("Content-Type", contentType)

	cookieLen := len(cookies)
	atmosCookie := new(http.Cookie)

	for i := 0; i < cookieLen; i++ {
		if cookies[i].Name == "AtmosphereSession" {
			atmosCookie = cookies[i]
		}
	}

	req.AddCookie(atmosCookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		activityLog.Error(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		activityLog.Error(err.Error())
	}

	// to understand the response we create a struct then unmarshal the JSON response to that struct
	// hint: attribute names should start with upperclass letter so they are exported
	// hint: the json tag after the definition tells go the real json attribute name
	// hint: there are various web tools that generate struct from JSON and vice versa
	type UpdateResp struct {
		UpdatedInstID string `json:"updatedInstId"`
	}

	activityLog.Info(bytes.NewBuffer(body).String())

	var s = new(UpdateResp)
	err2 := json.Unmarshal(body, &s)
	if err2 != nil {
		activityLog.Error("Error from unmarshal updatePF response::", err2)
	}

	activityLog.Info("Case Creator complete: ID " + s.UpdatedInstID)

	return s.UpdatedInstID
}
