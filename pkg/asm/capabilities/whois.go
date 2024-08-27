package capabilities

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

type WhoxyResponse struct {
	TotalPages   int            `json:"total_pages"`
	SearchResult []SearchResult `json:"search_result"`
}

type SearchResult struct {
	Domain    string `json:"domain_name"`
	QueryTime string `json:"query_time"`
}

type Whois struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

func NewWhois(job model.Job) model.Capability {
	return &Whois{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (task *Whois) Match() bool {
	return task.Asset.Is("tld")
}

func (task *Whois) Invoke() error {
	return task.whoisLookup("whois.iana.org")
}

func (task *Whois) Secret() string {
	return "whoxy"
}

func (task *Whois) whoisLookup(server string) error {
	raw, err := whois.Whois(task.Asset.DNS, server)
	if err != nil {
		return err
	}

	result, err := whoisparser.Parse(raw)
	if err != nil {
		return err
	}

	if result.Domain.WhoisServer != "" && server != result.Domain.WhoisServer {
		return task.whoisLookup(result.Domain.WhoisServer)
	}

	if result.Domain != nil {
		if expiring(result.Domain.ExpirationDateInTime) {
			risk := model.NewRisk(task.Asset, "domain-expiration")
			task.Job.Stream <- risk
			task.Job.Stream <- risk.Proof([]byte(raw))
		}
		task.send("purchased", result.Domain.CreatedDate)
		task.send("updated", result.Domain.UpdatedDate)
		task.send("expiration", result.Domain.ExpirationDate)
	}

	if result.Registrant != nil {
		task.send("country", result.Registrant.Country)
		task.send("province", result.Registrant.Province)
		task.send("city", result.Registrant.City)
	}

	if result.Registrar != nil {
		task.send("registrar", result.Registrar.Name)
	}

	emails := extractEmails([]*whoisparser.Contact{
		result.Registrant, result.Administrative, result.Billing, result.Technical}, task.Asset.DNS)

	domains := make(map[string]struct{})
	for email := range emails {
		results, err := task.reverseWhoisLookup(email)
		if err != nil {
			continue
		}
		for _, domain := range results {
			domains[domain] = struct{}{}
		}
	}

	for related := range domains {
		if related != task.Asset.DNS {
			asset := model.NewAsset(related, related)
			asset.Status = model.FrozenLow
			task.Job.Stream <- asset
		}
	}

	return nil
}

func (task *Whois) send(name, value string) {
	if value != "" {
		task.Job.Stream <- task.Asset.Attribute(name, value)
	}
}

func (task *Whois) reverseWhoisLookup(email string) ([]string, error) {
	var domains []string
	apiKey := task.Job.Config["secret"]
	if apiKey == "" {
		return nil, fmt.Errorf("whoxy API token not found")
	}

	page := 1
	totalPages := 1

	for {
		response, err := fetchWhoisPage(apiKey, email, page)
		if err != nil {
			return domains, err
		}

		for _, domain := range response.SearchResult {
			if !isOld(domain) {
				domains = append(domains, domain.Domain)
			}
		}
		if page == 1 {
			totalPages = response.TotalPages
		}

		page++
		if page > totalPages {
			break
		}
	}

	return domains, nil
}

func fetchWhoisPage(apiKey, email string, page int) (WhoxyResponse, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.whoxy.com/?key=%s&reverse=whois&email=%s&mode=micro&page=%d", apiKey, email, page))
	if err != nil {
		return WhoxyResponse{}, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return WhoxyResponse{}, fmt.Errorf("error reading response body: %v", err)
	}

	var result WhoxyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return WhoxyResponse{}, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return result, nil
}

func expiring(expirationDate *time.Time) bool {
	return expirationDate != nil && time.Until(*expirationDate) <= 30*24*time.Hour
}

func extractEmails(contacts []*whoisparser.Contact, domain string) map[string]struct{} {
	emails := make(map[string]struct{})
	for _, contact := range contacts {
		if contact != nil && contact.Email != "" && isMatch(domain, contact.Email) {
			emails[contact.Email] = struct{}{}
		}
	}
	return emails
}

func isOld(result SearchResult) bool {
	queryTime, _ := time.Parse("2006-01-02 15:04:05", result.QueryTime)
	return queryTime.Before(time.Now().AddDate(-10, 0, 0))
}

func isMatch(domain, email string) bool {
	parsed, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return strings.HasSuffix(parsed.Address, "@"+domain)
}
