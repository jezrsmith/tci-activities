/*
 * Copyright © 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package createComplaint

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/stretchr/testify/assert"
)

var activityMetadata *activity.Metadata

const (
	user = "<USERNAME>"
	pass = "<PASSWORD>"
	//accountid = "<ACCOUNTID>"
	accountid = ""
)

func getActivityMetadata() *activity.Metadata {
	if activityMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
		if err != nil {
			panic("No Json Metadata found for activity.json path")
		}
		activityMetadata = activity.NewMetadata(string(jsonMetadataBytes))
	}
	return activityMetadata
}

func TestActivityRegistration(t *testing.T) {
	act := NewActivity(getActivityMetadata())
	if act == nil {
		t.Error("Activity Not Registered")
		t.Fail()
		return
	}
}

func TestEval(t *testing.T) {
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	// setup connection
	// Generate a connection object
	connectionData := make(map[string]interface{})
	connectionSettings := make([]interface{}, 4)

	// Add a username
	usernameKey := make(map[string]interface{})
	usernameKey["name"] = "username"
	usernameKey["value"] = user
	connectionSettings[0] = usernameKey

	// Add a password
	passwordKey := make(map[string]interface{})
	passwordKey["name"] = "password"
	passwordKey["value"] = pass
	connectionSettings[1] = passwordKey

	// Add a region
	regionKey := make(map[string]interface{})
	regionKey["name"] = "region"
	regionKey["value"] = "EU"
	connectionSettings[2] = regionKey

	// Add accountId
	accountidKey := make(map[string]interface{})
	accountidKey["name"] = "accountid"
	accountidKey["value"] = accountid
	connectionSettings[3] = accountidKey

	connectionData["settings"] = connectionSettings

	//setup attrs
	tc.SetInput(ivConnection, connectionData)
	tc.SetInput("reference", 148)
	tc.SetInput("name", "Dave Smith")
	tc.SetInput("contact", "dave@gmail.com")
	tc.SetInput("type", "Customer Service")
	tc.SetInput("summary", "Waiting for callback")
	tc.SetInput("description", "Please call me back now!")
	_, err := act.Eval(tc)
	assert.Nil(t, err)
	result := tc.GetOutput("result")

	fmt.Println("Result is:" + result.(string))
	assert.NotNil(t, result)
	// use this to see output on a valid test
	// assert.NotNil(t, nil)
}
