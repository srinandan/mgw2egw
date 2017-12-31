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

package proxyutils

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

type ProxyEndpoint struct {
	XMLName             xml.Name            `xml:"ProxyEndpoint"`
	Name                string              `xml:"name,attr"`
	Description         string              `xml:"Description,omitempty"`
	FaultRules          string              `xml:"FaultRules,omitempty"`
	PreFlow             PreFlow             `xml:"PreFlow,omitempty"`
	PostFlow            PostFlow            `xml:"PostFlow,omitempty"`
	HTTPProxyConnection HTTPProxyConnection `xml:"HTTPProxyConnection,omitempty"`
	RouteRule           RouteRule           `xml:"RouteRule,omitempty"`
}

type PreFlow struct {
	XMLName  xml.Name `xml:"PreFlow"`
	Name     string   `xml:"name,attr"`
	Request  Request  `xml:"Request,omitempty"`
	Response Response `xml:"Response,omitempty"`
}

type Request struct {
	XMLName xml.Name `xml:"Request,omitempty"`
	Step    []Step   `xml:"Step,omitempty"`
}

type Step struct {
	Name string `xml:"Name,omitempty"`
}

type Response struct {
	XMLName xml.Name `xml:"Response,omitempty"`
	Step    []Step   `xml:"Step,omitempty"`
}

type PostFlow struct {
	XMLName  xml.Name `xml:"PostFlow"`
	Name     string   `xml:"name,attr"`
	Request  Request  `xml:"Request,omitempty"`
	Response Response `xml:"Response,omitempty"`
}

type HTTPProxyConnection struct {
	XMLName     xml.Name `xml:"HTTPProxyConnection"`
	BasePath    string   `xml:"BasePath"`
	Properties  string   `xml:"Properties"`
	VirtualHost []string `xml:"VirtualHost"`
}

type RouteRule struct {
	XMLName        xml.Name `xml:"RouteRule"`
	TargetEndpoint string   `xml:"TargetEndpoint"`
}

type APIProxy struct {
	XMLName              xml.Name             `xml:"APIProxy"`
	Name                 string               `xml:"name,attr"`
	Revision             string               `xml:"revision,attr"`
	Basepaths            string               `xml:"Basepaths,omitempty"`
	ConfigurationVersion ConfigurationVersion `xml:"ConfigurationVersion,omitempty"`
	CreatedAt            string               `xml:"CreatedAt,omitempty"`
	CreatedBy            string               `xml:"CreatedBy,omitempty"`
	Description          string               `xml:"Description,omitempty"`
	DisplayName          string               `xml:"DisplayName,omitempty"`
	LastModifiedAt       string               `xml:"LastModifiedAt,omitempty"`
	LastModifiedBy       string               `xml:"LastModifiedBy,omitempty"`
	Policies             Policies             `xml:"Policies,omitempty"`
	ProxyEndpoints       ProxyEndpoints       `xml:"ProxyEndpoints,omitempty"`
	Resources            string               `xml:"Resources,omitempty"`
	Spec                 string               `xml:"Spec,omitempty"`
	TargetServers        string               `xml:"TargetServers,omitempty"`
	TargetEndpoints      TargetEndpoints      `xml:"TargetEndpoints,omitempty"`
	Validate             string               `xml:"validate,omitempty"`
}

type ConfigurationVersion struct {
	XMLName      xml.Name `xml:"ConfigurationVersion,omitempty"`
	MajorVersion string   `xml:"majorVersion,attr"`
	MinorVersion string   `xml:"minorVersion,attr"`
}

type Policies struct {
	XMLName xml.Name `xml:"Policies"`
	Policy  []string `xml:"Policy,omitempty"`
}

type ProxyEndpoints struct {
	XMLName       xml.Name `xml:"ProxyEndpoints"`
	ProxyEndpoint []string `xml:"ProxyEndpoint,omitempty"`
}

type TargetEndpoints struct {
	XMLName        xml.Name `xml:"TargetEndpoints"`
	TargetEndpoint []string `xml:"TargetEndpoint,omitempty"`
}

func WriteAPIProxy(apiProxy APIProxy, fileName string) error {
	fileWriter, err := os.Create(fileName)
	if err != nil {
		return err
	}
	enc := xml.NewEncoder(fileWriter)
	err = enc.Encode(apiProxy)
	if err != nil {
		return err
	}
	defer fileWriter.Close()
	return nil
}

func WriteProxyEndpoint(proxyEndpoint ProxyEndpoint, fileName string) error {
	fileWriter, err := os.Create(fileName)
	if err != nil {
		return err
	}
	enc := xml.NewEncoder(fileWriter)
	err = enc.Encode(proxyEndpoint)
	if err != nil {
		return err
	}
	defer fileWriter.Close()
	return nil
}

func AddPolicyProxyEndpoint(proxyEndpoint ProxyEndpoint, policyNames ...string) ProxyEndpoint {

	for _, policyName := range policyNames {
		step := new(Step)
		step.Name = policyName
		proxyEndpoint.PreFlow.Request.Step = append(proxyEndpoint.PreFlow.Request.Step, *step)
	}
	return proxyEndpoint
}

func AddPolicyAPIProxy(apiProxy APIProxy, policyNames ...string) APIProxy {
	for _, policyName := range policyNames {
		apiProxy.Policies.Policy = append(apiProxy.Policies.Policy, policyName)
	}
	return apiProxy
}

func ReadProxyEndpoint(fileName string) (ProxyEndpoint, error) {
	var proxyEndpoint ProxyEndpoint
	content, err := ioutil.ReadFile(fileName)
	err = xml.Unmarshal(content, &proxyEndpoint)
	if err != nil {
		return proxyEndpoint, err
	}
	return proxyEndpoint, nil
}

func ReadAPIProxy(fileName string) (APIProxy, error) {
	var apiProxy APIProxy
	content, err := ioutil.ReadFile(fileName)
	err = xml.Unmarshal(content, &apiProxy)
	if err != nil {
		return apiProxy, err
	}
	return apiProxy, nil
}
