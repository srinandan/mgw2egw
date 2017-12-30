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

package main

import (
	"flag"
	"fmt"
	apigee "github.com/srinandan/go-apigee-edge"
	"io"
	"io/ioutil"
	"log"
	mgconfig "mgw2egw/microgatewayconfig"
	proxyutils "mgw2egw/proxyutils"
	utils "mgw2egw/utils"
	"os"
	"strings"
)

var org, env, username, password, configFile, fldr string
var infoLogger bool = false
var clientLogger bool = false

const version string = "1.0.0"
const proxyprefix string = "edgemicro_"
const oauthPolicyName string = "OAuth-v20-1"
const quotaPolicyName string = "Quota-1"
const verifyApiKeyName string = "Verify-API-Key-1"
const spikeArrestName string = "Spike-Arrest-1"

var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func Init(
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func checkParams(org, env, username, password, configFile string) {
	if org == "" {
		usage("orgname cannot be empty")
	} else if env == "" {
		usage("envname cannot be empty")
	} else if username == "" {
		usage("username cannot be empty")
	} else if password == "" {
		usage("password cannot be empty")
	} else if configFile == "" {
		usage("configFile cannot be empty")
	}
}

func main() {

	flag.StringVar(&org, "org", "", "Apigee Organization Name")
	flag.StringVar(&env, "env", "", "Apigee Environment Name")
	flag.StringVar(&username, "user", "", "Apigee Organization Username")
	flag.StringVar(&password, "pass", "", "Apigee Organization Password")
	flag.StringVar(&configFile, "conf", "", "Apigee Microgateway Config File")
	flag.StringVar(&fldr, "fldr", "/var/tmp", "Destination Folder to import proxies")
	flag.BoolVar(&infoLogger, "debug", false, "Enable debug mode")
	flag.BoolVar(&clientLogger, "trace", false, "Enable trace on Apigee Edge Client")

	flag.Parse()

	checkParams(org, env, username, password, configFile)

	if infoLogger {
		Init(os.Stdout, os.Stdout, os.Stderr)
	} else {
		Init(ioutil.Discard, os.Stdout, os.Stderr)
	}

	Info.Println("Reading Microgateway configuration file ", configFile)
	config, err := mgconfig.ReadConfig(configFile)
	if err != nil {
		Error.Fatalln("Unable to parse Microgateway configuration file: ", err)
		return
	}

	auth := apigee.EdgeAuth{Username: username, Password: password}
	opts := &apigee.EdgeClientOptions{Org: org, Auth: &auth, Debug: clientLogger}
	Info.Println("Initializing Apigee Edge client...")
	client, err := apigee.NewEdgeClient(opts)

	if err != nil {
		Error.Fatalln("Error initializing Edge client:\n%#v\n", err)
		return
	}
	Info.Println("Initialization successful!")

	edgemicroproxies, err := GetEdgemicroProxies(client)

	if err != nil {
		Error.Fatalln("Error downloading proxies:\n%#v\n", err)
		return
	}

	if len(edgemicroproxies) < 1 {
		Warning.Println("No Edgemicro proxies were found. Exiting program")
		return
	} else {
		Info.Println("Found Edgemicro proxies: ", edgemicroproxies)
	}

	for _, edgemicroproxy := range edgemicroproxies {
		if mgconfig.IsProxySet(edgemicroproxy, config) {
			Info.Println("Changing Proxy: ", edgemicroproxy)
			revision, err := GetLatestRevision(edgemicroproxy, client)
			if err != nil {
				return
			}
			Info.Println("Latest proxy revision is: ", revision)

			bundleName, err := DownloadProxy(edgemicroproxy, revision, client)
			if err != nil {
				return
			}
			Info.Println("Downloaded bundle: ", bundleName)
			Info.Println("Extracting bundle...")
			ExtractBundle(bundleName)

			err = AddPolicies(edgemicroproxy, bundleName, config)
			if err != nil {
				return
			}

			importtedRevision, err := ImportProxy(edgemicroproxy, bundleName, client)
			if err != nil {
				return
			}

			Info.Println("Deploying proxy ", edgemicroproxy, " with revision ", importtedRevision, " to ", env)
			DeployProxy(edgemicroproxy, env, revision, importtedRevision, client)

			Info.Println("Cleaning up ", bundleName)
			utils.Cleanup(bundleName)
		} else {
			Info.Println("Skipping Proxy: ", edgemicroproxy)
		}
	}
}

func AddPolicies(proxyName string, bundleName string, config mgconfig.Microgateway) error {

	Info.Println("Adding Edge policies to proxy...")
	plugins := mgconfig.GetPlugins(config)
	bundlePart := strings.Split(bundleName, ".")[0]
	policiesFolder := bundlePart + "/apiproxy/policies"
	apiProxyXMLFile := bundlePart + "/apiproxy/" + proxyName + ".xml"
	proxyEndpointXMLFile := bundlePart + "/apiproxy/proxies/default.xml"
	oauth := true

	apiProxy, err := proxyutils.ReadAPIProxy(apiProxyXMLFile)
	if err != nil {
		Error.Fatalln("Error reading APIProxy file:\n%#v\n", err)
		return err
	}

	proxyEndpoint, err := proxyutils.ReadProxyEndpoint(proxyEndpointXMLFile)
	if err != nil {
		Error.Fatalln("Error reading ProxyEndpoint file:\n%#v\n", err)
		return err
	}

	//create the policies folder
	os.Mkdir(policiesFolder, 0777)

	for _, plugin := range plugins {
		if plugin == "oauth" {
			if mgconfig.APIKeyOnly(config) {
				Info.Println("Adding VerifyAPIKey policy")
				utils.CopyAPIKey(policiesFolder)
				apiProxy = proxyutils.AddPolicyAPIProxy(verifyApiKeyName, apiProxy)
				proxyEndpoint = proxyutils.AddPolicyProxyEndpoint(verifyApiKeyName, proxyEndpoint)
				oauth = false
			} else {
				Info.Println("Adding OAuth v2.0 policy")
				utils.CopyOAuth(policiesFolder)
				apiProxy = proxyutils.AddPolicyAPIProxy(oauthPolicyName, apiProxy)
				proxyEndpoint = proxyutils.AddPolicyProxyEndpoint(oauthPolicyName, proxyEndpoint)
			}
		} else if plugin == "quota" {
			Info.Println("Adding Quota policy")
			utils.CopyQuota(policiesFolder, oauth)
			apiProxy = proxyutils.AddPolicyAPIProxy(quotaPolicyName, apiProxy)
			proxyEndpoint = proxyutils.AddPolicyProxyEndpoint(quotaPolicyName, proxyEndpoint)
		} else if plugin == "spikearrest" {
			Info.Println("Adding SpikeArrest policy")
			Timeunit, Allow := mgconfig.GetSpikeArrestDetails(config)
			utils.CopySpikeArrest(policiesFolder, Timeunit, Allow)
			apiProxy = proxyutils.AddPolicyAPIProxy(spikeArrestName, apiProxy)
			proxyEndpoint = proxyutils.AddPolicyProxyEndpoint(spikeArrestName, proxyEndpoint)
		}
	}

	err = proxyutils.WriteProxyEndpoint(proxyEndpoint, proxyEndpointXMLFile)
	if err != nil {
		Error.Fatalln("Error writing to ProxyEndpoint file:\n%#v\n", err)
		return err
	}
	err = proxyutils.WriteAPIProxy(apiProxy, apiProxyXMLFile)
	if err != nil {
		Error.Fatalln("Error writing to APIProxy file:\n%#v\n", err)
		return err
	}

	return nil
}

func ExtractBundle(bundleName string) {
	bundlePart := strings.Split(bundleName, ".")[0]
	utils.Unzip(bundleName, bundlePart)
}

func GetLatestRevision(proxyName string, client *apigee.EdgeClient) (apigee.Revision, error) {

	var revision apigee.Revision
	proxyRevs, resp, e := client.Proxies.Get(proxyName)

	if e != nil {
		Error.Fatalln("Error getting revision:\n%#v\n", e)
		return revision, e
	}
	defer resp.Body.Close()
	return proxyRevs.Revisions[len(proxyRevs.Revisions)-1], nil
}

func ImportProxy(proxyName string, bundleName string, client *apigee.EdgeClient) (apigee.Revision, error) {
	bundlePart := strings.Split(bundleName, ".")[0]
	proxyRev, resp, e := client.Proxies.Import(proxyName, bundlePart)
	if e != nil {
		Error.Fatalln("Error while importing proxy:\n%#v\n", e)
		return proxyRev.Revision, e
	}
	defer resp.Body.Close()
	return proxyRev.Revision, nil
}

func DeployProxy(proxyName string, env string, oldRevision apigee.Revision, newRevision apigee.Revision, client *apigee.EdgeClient) error {

	_, resp, e := client.Proxies.Undeploy(proxyName, env, oldRevision)
	if e != nil {
		Error.Fatalln("Error undeploying proxy:\n%#v\n", e)
		return e
	}
	resp.Body.Close()

	_, resp, e = client.Proxies.Deploy(proxyName, env, newRevision)
	if e != nil {
		Error.Fatalln("Error deploying proxy:\n%#v\n", e)
		return e
	}
	resp.Body.Close()
	return nil
}

func DownloadProxy(proxyName string, revision apigee.Revision, client *apigee.EdgeClient) (string, error) {

	proxyRev, resp, e := client.Proxies.Export(proxyName, revision)
	if e != nil {
		Error.Fatalln("Error downloading proxy:\n%#v\n", e)
		return proxyRev, e
	}
	defer resp.Body.Close()
	return proxyRev, nil
}

func GetEdgemicroProxies(client *apigee.EdgeClient) ([]string, error) {

	var edgemicroproxies []string
	Info.Println("Downloading Edgemicro proxy list")
	proxies, resp, e := client.Proxies.List()
	if e != nil {
		return edgemicroproxies, e
	}
	defer resp.Body.Close()
	for _, proxy := range proxies {
		if strings.HasPrefix(proxy, proxyprefix) {
			edgemicroproxies = append(edgemicroproxies, proxy)
		}
	}
	return edgemicroproxies, nil
}

func usage(message string) {
	fmt.Println("")
	if message != "" {
		fmt.Println("Incorrect or incomplete parameters, ", message)
	}
	fmt.Println("mgw2egw version ", version)
	fmt.Println("")
	fmt.Println("Usage: mgw2egw -org=<orgname> -env=<envname> -user=<username> -pass=<password> -conf=<conf file>")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("org  = Apigee Edge Organization name (mandatory)")
	fmt.Println("env  = Apigee Edge Environment name (mandatory)")
	fmt.Println("user = Apigee Edge Username (mandatory)")
	fmt.Println("pass = Apigee Edge Password (mandatory)")
	fmt.Println("conf = Apigee Edge Microgateway configuration file (mandatory)")
	fmt.Println("")
	fmt.Println("Other options:")
	fmt.Println("fldr  = Folder to extract Apigee Bundle (default: /var/tmp)")
	fmt.Println("debug = Enable debug mode (default: false)")
	fmt.Println("trace = Enable trace on go-apigee-edge (default: false)")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("Example: mgw2egw -org=trial -env=test -user=trial@apigee.com -pass=Secret123 -config=trial-test-config.yaml")
	os.Exit(1)
}
