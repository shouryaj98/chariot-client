package model

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Attribute struct {
	Username string `dynamodbav:"username" json:"username"`
	Key      string `dynamodbav:"key" json:"key"`
	// Attributes
	Source  string `dynamodbav:"source" json:"source"`
	Name    string `dynamodbav:"name" json:"name"`
	Value   string `dynamodbav:"value" json:"value"`
	Created string `dynamodbav:"created" json:"created"`
	Updated string `dynamodbav:"updated" json:"updated"`
	TTL     int64  `dynamodbav:"ttl" json:"ttl"`
}

func (a *Attribute) Target(status string) Asset {
	v := strings.Split(a.Source, "#")
	fqdn := fmt.Sprintf("%s://%s:%s", a.Name, v[2], a.Value)

	asset := NewAsset(v[2], fqdn)
	asset.Key = a.Key
	asset.Status = status
	return asset
}

func (a *Attribute) Tasks() []string {
	var tasks []string
	if _, err := strconv.Atoi(a.Value); err != nil {
		return tasks
	}
	if strings.HasPrefix(a.Name, "http") {
		tasks = append(tasks, "crawler", "nuclei")
	}
	if a.Name == "ssh" {
		tasks = append(tasks, "ssh")
	}

	return tasks
}

func (a *Attribute) Valid() bool {
	pattern, _ := regexp.Compile(`#attribute(#[^\s#]+){2}#`)
	return pattern.MatchString(a.Key)
}

func (a *Attribute) Visit() Attribute {
	a.Updated = Now()
	a.TTL = Future(7 * 24)
	return *a
}

func NewAttribute(name, value, key string) Attribute {
	k := fmt.Sprintf("#attribute#%s#%s%s", name, value, key)
	return Attribute{
		Source:  key,
		Name:    name,
		Value:   value,
		Updated: Now(),
		Created: Now(),
		TTL:     Future(24 * 7),
		Key:     k[:min(1024, len(k))],
	}
}
