package model

import "fmt"

type File struct {
	Username string `dynamodbav:"username" json:"username"`
	Key      string `dynamodbav:"key" json:"key"`
	// Attributes
	Name    string `dynamodbav:"name" json:"name"`
	Updated string `dynamodbav:"updated" json:"updated"`
	Bytes   []byte `dynamodbav:"-" json:"-"`
}

func NewFile(name string) File {
	return File{
		Name:    name,
		Updated: Now(),
		Key:     fmt.Sprintf("#file#%s", name),
	}
}
