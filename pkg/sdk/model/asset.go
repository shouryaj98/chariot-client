package model

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/publicsuffix"
)

type Asset struct {
	Username string `dynamodbav:"username" json:"username"`
	Key      string `dynamodbav:"key" json:"key"`
	// Attributes
	Source  string            `dynamodbav:"source" json:"source"`
	DNS     string            `dynamodbav:"dns" json:"dns"`
	Name    string            `dynamodbav:"name" json:"name"`
	Status  string            `dynamodbav:"status" json:"status"`
	Config  map[string]string `dynamodbav:"config" json:"config"`
	Created string            `dynamodbav:"created" json:"created"`
	Updated string            `dynamodbav:"updated" json:"updated"`
	TTL     int64             `dynamodbav:"ttl" json:"ttl"`
	History []History         `dynamodbav:"history" json:"history"`
}

func (a *Asset) Class() string {
	classes := []func(string) (string, bool){
		func(s string) (string, bool) {
			domain, _ := publicsuffix.EffectiveTLDPlusOne(a.DNS)
			return "tld", domain == a.DNS
		},
		func(s string) (string, bool) {
			_, _, err := net.ParseCIDR(s)
			return "cidr", err == nil
		},
		func(s string) (string, bool) {
			re := regexp.MustCompile(`^(https://)?github\.com/([^/]+)/?$`)
			return "github", re.FindStringSubmatch(s) != nil
		},
		func(s string) (string, bool) {
			re := regexp.MustCompile(`^(https://)?gitlab\.com/([^/]+)/?$`)
			return "gitlab", re.FindStringSubmatch(s) != nil
		},
		func(s string) (string, bool) {
			re := regexp.MustCompile(`^(https://)?(github\.com|gitlab\.com)/([^/]+)/(([^/]+/)*[^/]+)$`)
			return "repository", re.FindStringSubmatch(s) != nil
		},
		func(s string) (string, bool) {
			return s, s == "amazon" || s == "azure" || s == "gcp" || s == "ns1" || s == "cloudflare" || s == "crowdstrike"
		},
		func(s string) (string, bool) {
			pattern := `^(https?://)?((xn--[a-zA-Z0-9-]+|[a-zA-Z0-9-]+)\.)+([a-zA-Z]{2,})$`
			valid, _ := regexp.MatchString(pattern, s)
			return "domain", valid
		},
		func(s string) (string, bool) {
			url, err := url.Parse(s)
			if err != nil || url.Scheme == "" {
				return "", false
			}
			return url.Scheme, true
		},
		func(s string) (string, bool) {
			ip := net.ParseIP(s)
			if ip == nil {
				return "", false
			}
			if ip.IsPrivate() {
				return "private", true
			}
			if ip4 := ip.To4(); ip4 != nil {
				return "ipv4", true
			}
			return "ipv6", true
		},
		func(s string) (string, bool) {
			re := regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
			return "endpoint", re.MatchString(s)
		},
		func(s string) (string, bool) {
			re := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)
			return "agent", re.FindStringSubmatch(s) != nil
		},
	}

	for _, target := range []string{a.Name, a.DNS} {
		for _, class := range classes {
			if c, ok := class(target); ok {
				return c
			}
		}
	}
	return ""
}

func (a *Asset) Valid() bool {
	pattern, _ := regexp.Compile(`#asset(#[^\s#]+){2}$`)
	return pattern.MatchString(a.Key)
}

func (a *Asset) Is(value string) bool {
	return strings.HasPrefix(a.Status, value) || strings.HasPrefix(a.Class(), value)
}

func (a *Asset) Merge(update Asset) {
	if update.Status != "" && a.Status != update.Status {
		event := History{From: a.Status, To: update.Status, Updated: Now(), By: update.Source}
		a.History = append(a.History, event)
		a.Status = update.Status
	}
	if update.Created != "" {
		a.Created = update.Created
	}
	if !a.Is(Active) {
		a.TTL = 0
	}
}

func (a *Asset) Visit(config map[string]string) Asset {
	a.Updated = Now()
	if a.Is(Active) {
		a.TTL = Future(7 * 24)
	}
	a.Config = config
	return *a
}

func (a *Asset) Attribute(name, value string) Attribute {
	return NewAttribute(name, value, a.Key)
}

func (a *Asset) System() bool {
	return a.Source == Discovered
}

func (a *Asset) Spawn(dns, name string) Asset {
	asset := NewAsset(dns, name)
	asset.Status = a.Status
	return asset
}

func NewAsset(dns, name string) Asset {
	return Asset{
		DNS:     dns,
		Name:    name,
		Status:  Active,
		Source:  Discovered,
		Created: Now(),
		Updated: Now(),
		TTL:     Future(7 * 24),
		Key:     fmt.Sprintf("#asset#%s#%s", dns, name),
	}
}
