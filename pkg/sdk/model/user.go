package model

import "strings"

type User struct {
	Name     string    `dynamodbav:"username"`
	Key      string    `dynamodbav:"key"`
	Updated  string    `dynamodbav:"updated"`
	Accounts []Account `dynamodbav:"-"`
}

func (u *User) Linked(username string) bool {
	for _, account := range u.Accounts {
		if account.Name == username && account.Member == u.Name {
			return true
		}
	}
	return false
}

func (u *User) Domain() string {
	parts := strings.Split(u.Name, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

func NewUser(name string, accounts []Account) User {
	return User{Name: name, Key: "#user", Updated: Now(), Accounts: accounts}
}
