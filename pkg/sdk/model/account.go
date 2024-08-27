package model

import "fmt"

type Account struct {
	Username string `dynamodbav:"username" json:"username"`
	Key      string `dynamodbav:"key" json:"key"`
	// Attributes
	Name    string            `dynamodbav:"name" json:"name"`
	Member  string            `dynamodbav:"member" json:"member"`
	Value   string            `dynamodbav:"value,omitempty" json:"value"`
	Config  map[string]string `dynamodbav:"config" json:"config"`
	Updated string            `dynamodbav:"updated" json:"updated"`
	TTL     int64             `dynamodbav:"ttl" json:"ttl"`
}

func NewAccount(name, member, value string, config map[string]string) Account {
	return Account{
		Name:    name,
		Member:  member,
		Value:   value,
		Config:  config,
		Updated: Now(),
		Key:     fmt.Sprintf("#account#%s#%s#%s", name, member, value),
	}
}
