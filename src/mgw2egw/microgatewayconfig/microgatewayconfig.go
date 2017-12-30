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

package microgatewayconfig

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Microgateway struct {
	Edgeconfig  EdgeConfig  `yaml:"edge_config,omitempty"`
	Edgemicro   EdgeMicro   `yaml:"edgemicro,omitempty"`
	Header      Headers     `yaml:"headers,omitempty"`
	Spikearrest SpikeArrest `yaml:"spikearrest,omitempty"`
	Oauth       OAuth       `yaml:"oauth,omitempty"`
	Ax          Analytics   `yaml:"analytics,omitempty"`
}

type EdgeConfig struct {
	Bootstrap        string `yaml:"bootstrap,omitempty"`
	JwtPublicKey     string `yaml:"jwt_public_key,omitempty"`
	Management       string `yaml:"managementUri,omitempty"`
	Vaultname        string `yaml:"vaultName,omitempty"`
	Auth             string `yaml:"authUri,omitempty"`
	Base             string `yaml:"baseUri,omitempty"`
	BootstrapMessage string `yaml:"bootstrapMessage,omitempty"`
	KeysecretMessage string `yaml:"keySecretMessage,omitempty"`
	Products         string `yaml:"products,omitempty"`
}

type EdgeMicro struct {
	Port                     int      `yaml:"port,omitempty"`
	MaxConnections           int      `yaml:"max_connections,omitempty"`
	ConfigChangePollInterval int      `yaml:"config_change_poll_interval,omitempty"`
	Log                      Logging  `yaml:"logging,omitempty"`
	Plugin                   Plugins  `yaml:"plugins,omitempty"`
	Proxies                  []string `yaml:"proxies,omitempty"`
}

type Logging struct {
	Level            string `yaml:"level,omitempty"`
	Dir              string `yaml:"dir,omitempty"`
	StatslogInterval int    `yaml:"stats_log_interval,omitempty"`
	RotateInterval   int    `yaml:"rotate_interval,omitempty"`
}

type Plugins struct {
	Sequence []string `yaml:"sequence,omitempty"`
}

type Headers struct {
	XforwardedFor  bool `yaml:"x-forwarded-for,omitempty"`
	XForwardedHost bool `yaml:"x-forwarded-host,omitempty"`
	XRequestId     bool `yaml:"x-request-id,omitempty"`
	XResponseTime  bool `yaml:"x-response-time,omitempty"`
	Via            bool `yaml:"via,omitempty"`
}

type Analytics struct {
	Uri string `yaml:"uri,omitempty"`
}

type OAuth struct {
	AllowNoAuthorization      bool   `json:"allowNoAuthorization,omitempty" yaml:"allowNoAuthorization,omitempty"`
	AllowInvalidAuthorization bool   `yaml:"allowInvalidAuthorization,omitempty"`
	ProductOnly               string `yaml:"productOnly,omitempty"`
	GracePeriod               string `yaml:"gracePeriod,omitempty"`
	CacheKey                  string `yaml:"cacheKey,omitempty"`
	VerifyApiKeyUrl           string `yaml:"verify_api_key_url,omitempty"`
	AllowAPIKeyOnly           bool   `yaml:"allowAPIKeyOnly,omitempty"`
}

type SpikeArrest struct {
	TimeUnit   string `yaml:"timeUnit,omitempty"`
	Allow      int    `yaml:"allow,omitempty"`
	Buffersize int    `yaml:"buffersize,omitempty"`
}

var proxyMap = map[string]string{}

func ReadConfig(filepath string) (microgateway Microgateway, err error) {
	var config Microgateway
	source, err := ioutil.ReadFile(filepath)

	if err != nil {
		return Microgateway{}, err
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		return Microgateway{}, err
	}

	if len(config.Edgemicro.Proxies) > 0 {
		for _, proxy := range config.Edgemicro.Proxies {
			proxyMap[proxy] = proxy
		}
	}

	return config, nil
}

func GetPlugins(microgateway Microgateway) []string {
	return microgateway.Edgemicro.Plugin.Sequence
}

func GetProxies(microgateway Microgateway) []string {
	return microgateway.Edgemicro.Proxies
}

func IsProxySet(proxyName string, microgateway Microgateway) bool {

	if len(microgateway.Edgemicro.Proxies) == 0 {
		return true
	}
	_, ok := proxyMap[proxyName]
	return ok
}

func APIKeyOnly(microgateway Microgateway) bool {
	return microgateway.Oauth.AllowAPIKeyOnly
}

func GetSpikeArrestDetails(microgateway Microgateway) (string, int) {
	return microgateway.Spikearrest.TimeUnit, microgateway.Spikearrest.Allow
}

/*func main() {
	filename := os.Args[1]
	config, _ := readConfig(filename)
	fmt.Printf("Value: %+v\n", config)
}*/
