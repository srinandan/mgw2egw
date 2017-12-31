# mgw2egw
mgw2egw is an open source project that extends Apigee Edge functionality to convert Apigee Edge Microgateway proxies to Apigee Edge Enterprise Gateway proxies

## Support
This is an open-source project of the Apigee Corporation. It is not covered by Apigee support contracts. However, we will support you as best we can. For help, please open an issue in this GitHub project. You are also always welcome to submit a pull request.

## Download
Download the binary from [here](https://github.com/srinandan/mgw2egw/releases)

## Usage
```
mgw2egw -org=<orgname> -env=<envname> -user=<username> -pass=<password> -conf=<conf file> 
```

### Options
```
org  = Apigee Edge Organization name (mandatory)
env  = Apigee Edge Environment name (mandatory)
user = Apigee Edge Username (mandatory)
pass = Apigee Edge Password (mandatory)
conf = Apigee Edge Microgateway configuration file (mandatory)
```

#### Other Options
```
fldr    = Folder to extract Apigee Bundle (default: /var/tmp)
debug   = Enable debug mode (default: false)
trace   = Enable trace on go-apigee-edge (default: false)
importonly = Import the proxies only, do not deploy
genonly = Generate the bundles only, do not import
usejwt  = Use JWT policies to validate OAuth tokens
```

### How does it work?
A typical Apige Edge Microgateway configuration file looks like this (some details omitted for brevity):
```
edge_config:
  bootstrap: >-
    http://localhost:9001/edgemicro/bootstrap/organization/trial/environment/test
  ...
edgemicro:
  port: 8000
  ...
  plugins:
    sequence:
      - oauth
  proxies:
    - edgemicro_httpbin
headers:
  ...
oauth:
  allowNoAuthorization: false
  allowInvalidAuthorization: false
  verify_api_key_url: 'http://localhost:9001/edgemicro-auth/verifyApiKey'
analytics:
  uri: >-
    http://localhost:9001/edgemicro/axpublisher/organization/trial/environment/test
```
Micrgoateway uses plugins to enable policies. The standard set of plugins offered by Apigee Edge Microgateway can be found [here](https://github.com/apigee/microgateway-plugins). This tool scans the Microgateway configuration file for plugins enabled and adds the appropriate Apigee Edge [policies](https://docs.apigee.com/api-services/reference/reference-overview-policy).

#### Which proxies are converted?
By default, all proxies which follow the pattern `edgemicro_*` are converted. However, if the `proxies` tag is specified, then only proxies specified in the tag are converted.

#### List of supported plugins
* OAuth
* Verify API Key
* Spike Arrest
* Quota

#### What about custom plugins?
Custom plugins are not supported. They'll have to be reimplemented manually using Apigee Edge policies.  

### Build Instructions
MGW2EGW_HOME = The folder where you've downloaded the code
 
```
git clone https://github.com/srinandan/mgw2egw.git && cd mgw2egw/src/mgw2egw

export GOPATH=$GOPATH:$MGW2EGW_HOME

go get github.com/srinandan/go-apigee-edge

go install
```

The binary file will be stored in `MGW2EGW_HOME/bin`

### TODO
* Automated testing
