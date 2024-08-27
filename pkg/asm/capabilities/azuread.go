package capabilities

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

type AzureAD struct {
	Stream chan interface{}
	Asset  model.Asset
	XYZ
}

type Realm struct {
	Domain string `json:"DomainName"`
	Type   string `json:"NameSpaceType"`
}

func NewAzureAD(job model.Job) model.Capability {
	return &AzureAD{Asset: job.Target, Stream: job.Stream, XYZ: NewXYZ()}
}

func (task *AzureAD) Match() bool {
	return task.Asset.Is("tld") && !task.Asset.System()
}

func (task *AzureAD) Invoke() error {
	for _, domain := range task.domains() {
		realm, err := task.realm(domain)
		if err != nil {
			continue
		}
		if strings.EqualFold(realm.Domain, task.Asset.DNS) {
			continue
		}
		if realm.Type == "Federated" {
			task.adjacent(realm.Domain)
		}
	}
	return nil
}

func (task *AzureAD) adjacent(domain string) {
	asset := model.NewAsset(domain, domain)
	asset.Status = model.FrozenLow
	task.Stream <- asset
}

func (task *AzureAD) realm(domain string) (*Realm, error) {
	url := "https://login.microsoftonline.com/GetUserRealm.srf?login=" + "nn@" + domain
	return request[Realm]("GET", url, nil)
}

func (task *AzureAD) domains() []string {
	url := "https://autodiscover-s.outlook.com/autodiscover/autodiscover.svc"
	body := `<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:exm="http://schemas.microsoft.com/exchange/services/2006/messages" xmlns:ext="http://schemas.microsoft.com/exchange/services/2006/types" xmlns:a="http://www.w3.org/2005/08/addressing" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
		<soap:Header>
			<a:Action soap:mustUnderstand="1">http://schemas.microsoft.com/exchange/2010/Autodiscover/Autodiscover/GetFederationInformation</a:Action>
			<a:To soap:mustUnderstand="1">https://autodiscover-s.outlook.com/autodiscover/autodiscover.svc</a:To>
			<a:ReplyTo>
				<a:Address>http://www.w3.org/2005/08/addressing/anonymous</a:Address>
			</a:ReplyTo>
		</soap:Header>
		<soap:Body>
			<GetFederationInformationRequestMessage xmlns="http://schemas.microsoft.com/exchange/2010/Autodiscover">
				<Request>
					<Domain>%s</Domain>
				</Request>
			</GetFederationInformationRequestMessage>
		</soap:Body>
	</soap:Envelope>`
	resp, err := _request("POST", url, []byte(fmt.Sprintf(body, task.Asset.DNS)), "Content-Type", "text/xml", "SOAPAction",
		"http://schemas.microsoft.com/exchange/2010/Autodiscover/Autodiscover/GetFederationInformation", "User-Agent", "AutodiscoverClient")

	domains := make([]string, 0)
	if err == nil {
		re := regexp.MustCompile(`<Domain>(.*?)</Domain>`)
		matches := re.FindAllStringSubmatch(string(resp), -1)
		for _, match := range matches {
			domains = append(domains, match[1])
		}
	}

	return domains
}
