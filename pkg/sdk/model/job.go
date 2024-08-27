package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Job struct {
	Username string `dynamodbav:"username" json:"username"`
	Key      string `dynamodbav:"key" json:"key"`
	// Attributes
	DNS     string            `dynamodbav:"dns" json:"dns"`
	Source  string            `dynamodbav:"source" json:"source"`
	Comment string            `dynamodbav:"comment" json:"comment"`
	Config  map[string]string `dynamodbav:"config" json:"config"`
	Created string            `dynamodbav:"created" json:"created"`
	Updated string            `dynamodbav:"updated" json:"updated"`
	Status  string            `dynamodbav:"status" json:"status"`
	TTL     int64             `dynamodbav:"ttl" json:"ttl"`
	Name    string            `dynamodbav:"name,omitempty" json:"name,omitempty"`
	Queue   string            `dynamodbav:"-"`
	Target  Asset             `dynamodbav:"-"`
	Stream  chan interface{}  `dynamodbav:"-" json:"-"`
}

func (j *Job) Raw() string {
	rawJSON, _ := json.Marshal(j)
	return string(rawJSON)
}

func (j *Job) Is(status string) bool {
	return strings.HasPrefix(j.Status, status)
}

func (job *Job) Update(status string) {
	job.Status = status
	job.Updated = Now()
	job.TTL = Future(24 * 7)
}

func ConstructJob(key string) (Job, error) {
	re := regexp.MustCompile("#job#([^#]+)#([^#]+)#([^#]+)$")
	matches := re.FindStringSubmatch(key)
	if matches == nil {
		return Job{}, errors.New("invalid job key")
	}
	return NewJob(matches[3], NewAsset(matches[1], matches[2])), nil
}

func NewJob(source string, asset Asset) Job {
	return Job{
		DNS:     asset.DNS,
		Source:  source,
		Target:  asset,
		Status:  Queued,
		Config:  make(map[string]string),
		Created: Now(),
		Updated: Now(),
		TTL:     Future(12),
		Queue:   os.Getenv("CHARIOT_QUEUE_STANDARD"),
		Key:     fmt.Sprintf("#job#%s#%s#%s", asset.DNS, asset.Name, source),
	}
}
