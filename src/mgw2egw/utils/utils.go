// Copyright 2017 Apigee
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//const oauthPath string = "./templates/OAuth-v20-1.xml"
const oauthPath string = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<OAuthV2 async="false" continueOnError="false" enabled="true" name="OAuth-v20-1">
    <DisplayName>OAuth v2.0-1</DisplayName>
    <Properties/>
    <Attributes/>
    <ExternalAuthorization>false</ExternalAuthorization>
    <Operation>VerifyAccessToken</Operation>
    <SupportedGrantTypes/>
    <GenerateResponse enabled="true"/>
    <Tokens/>
</OAuthV2>`
const oauthName string = "/OAuth-v20-1.xml"

//const quotaPath string = "./templates/Quota-1.xml"
const quotaPathAPIKey string = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Quota async="false" continueOnError="false" enabled="true" name="Quota-1" type="calendar">
    <DisplayName>Quota-1</DisplayName>
    <Properties/>
    <Allow count="200" countRef="verifyapikey.Verify-API-Key-1.apiproduct.developer.quota.limit"/>
    <Interval ref="verifyapikey.Verify-API-Key-1.apiproduct.developer.quota.interval">1</Interval>
    <Distributed>true</Distributed>
    <Synchronous>true</Synchronous>
	<PreciseAtSecondsLevel>false</PreciseAtSecondsLevel>
    <TimeUnit ref="verifyapikey.Verify-API-Key-1.apiproduct.developer.quota.timeunit">hour</TimeUnit>
</Quota>`
const quotaPathOAuth string = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Quota async="false" continueOnError="false" enabled="true" name="Quota-1" type="calendar">
    <DisplayName>Quota-1</DisplayName>
    <Properties/>
    <Allow count="200" countRef="apiproduct.developer.quota.limit"/>
    <Interval ref="apiproduct.developer.quota.interval">1</Interval>
    <Distributed>true</Distributed>
    <Synchronous>true</Synchronous>
	<PreciseAtSecondsLevel>false</PreciseAtSecondsLevel>
    <TimeUnit ref="apiproduct.developer.quota.timeunit">hour</TimeUnit>
</Quota>`
const quotaName string = "/Quota-1.xml"

//const spikeArrestPath string = "./templates/Spike-Arrest-1.xml"
var spikeArrestPath string = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<SpikeArrest async="false" continueOnError="false" enabled="true" name="Spike-Arrest-1">
    <DisplayName>Spike Arrest-1</DisplayName>
    <Properties/>
    <Rate>30ps</Rate>
</SpikeArrest>`

const spikeArrestName string = "/Spike-Arrest-1.xml"

//const verifyApiKeyPath string = "./templates/Verify-API-Key-1.xml"
const verifyApiKeyPath string = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<VerifyAPIKey async="false" continueOnError="false" enabled="true" name="Verify-API-Key-1">
    <DisplayName>Verify API Key-1</DisplayName>
    <Properties/>
    <APIKey ref="request.header.x-api-key"/>
</VerifyAPIKey>`

const verifyApiKeyName string = "/Verify-API-Key-1.xml"

// Unzip will uncompress a zip archive
func Unzip(src, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)
		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, os.ModePerm)
			if err != nil {
				log.Fatal(err)
				return filenames, err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return filenames, err
			}

		}
	}
	return filenames, nil
}

func readFile(fileName string) ([]byte, error) {
	absFileName, _ := filepath.Abs(fileName)
	content, err := ioutil.ReadFile(absFileName)
	return content, err
}

func writeFile(fileName string, content []byte) error {
	err := ioutil.WriteFile(fileName, content, 0777)
	return err
}

func CopyOAuth(folder string) error {
	err := writeFile(folder+oauthName, []byte(oauthPath))
	if err != nil {
		return err
	}
	return nil
}

func CopyQuota(folder string, oauth bool) error {

	var err error

	if oauth {
		err = writeFile(folder+quotaName, []byte(quotaPathOAuth))
	} else {
		err = writeFile(folder+quotaName, []byte(quotaPathAPIKey))
	}

	if err != nil {
		return err
	}

	return nil
}

func CopySpikeArrest(folder string, Timeunit string, Allow int) error {
	var rate string

	if Timeunit == "minute" {
		rate = strconv.Itoa(Allow) + "pm"
	} else {
		rate = strconv.Itoa(Allow) + "ps"
	}

	spikeArrestPath = strings.Replace(spikeArrestPath, "30ps", rate, 1)

	err := writeFile(folder+spikeArrestName, []byte(spikeArrestPath))
	if err != nil {
		return err
	}
	return nil
}

func CopyAPIKey(folder string) error {
	err := writeFile(folder+verifyApiKeyName, []byte(verifyApiKeyPath))
	if err != nil {
		return err
	}
	return nil
}

func Cleanup(bundleName string) error {
	err := os.Remove(bundleName)
	err = os.RemoveAll(strings.Split(bundleName, ".")[0])
	return err
}
