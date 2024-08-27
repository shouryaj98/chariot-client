package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type Risk struct {
	Username string `dynamodbav:"username" json:"username"`
	Key      string `dynamodbav:"key" json:"key"`
	// Attributes
	DNS     string    `dynamodbav:"dns" json:"dns"`
	Name    string    `dynamodbav:"name" json:"name"`
	Source  string    `dynamodbav:"source" json:"source"`
	Status  string    `dynamodbav:"status" json:"status"`
	Created string    `dynamodbav:"created" json:"created"`
	Updated string    `dynamodbav:"updated" json:"updated"`
	History []History `dynamodbav:"history" json:"history"`
	TTL     int64     `dynamodbav:"ttl" json:"ttl"`
	Comment string    `dynamodbav:"-" json:"comment"`
}

func (r *Risk) Raw() string {
	rawJSON, _ := json.Marshal(r)
	return string(rawJSON)
}

func (r *Risk) Valid() bool {
	pattern, _ := regexp.Compile(`^#risk#(\S+)#(\S+)$`)
	return pattern.MatchString(r.Key)
}

func (r *Risk) Is(status string) bool {
	return strings.HasPrefix(r.Status, status)
}

func (r *Risk) Merge(update Risk) {
	if update.Status != "" && r.Status != update.Status {
		r.History = append(r.History, History{
			From:    r.Status,
			To:      update.Status,
			By:      update.Source,
			Comment: update.Comment,
			Updated: Now(),
		})
		r.Status = update.Status
	}
	if update.Created != "" {
		r.Created = update.Created
	}
	if !r.Is(Triage) {
		r.TTL = 0
	}
}

func (r *Risk) Visit() Risk {
	r.Updated = Now()

	if r.Is(Triage) {
		r.TTL = Future(7 * 24)
	}

	if r.Is(Remediated) {
		r.Set(Open)
	}

	return *r
}

func (r *Risk) Set(status string) {
	update := *r
	update.Status = status + r.Severity()
	r.Merge(update)
}

func (r *Risk) Proof(bits []byte) File {
	file := NewFile(fmt.Sprintf("proofs/%s/%s", r.DNS, r.Name))
	file.Bytes = bits
	return file
}

func (r *Risk) Severity() string {
	return string(r.Status[len(r.Status)-1])
}

func (r *Risk) State() string {
	return string(r.Status[:len(r.Status)-1])
}

func (r *Risk) Link() string {
	return fmt.Sprintf("https://preview.chariot.praetorian.com/risks?riskKey=%s", r.Key)
}

func (r *Risk) Attribute(name, value string) Attribute {
	return NewAttribute(name, value, r.Key)
}

func NewRisk(asset Asset, name string) Risk {
	return Risk{
		DNS:     asset.DNS,
		Name:    name,
		Status:  TriageInfo,
		Source:  Provided,
		Created: Now(),
		Updated: Now(),
		TTL:     Future(7 * 24),
		Key:     fmt.Sprintf("#risk#%s#%s", asset.DNS, name),
	}
}
