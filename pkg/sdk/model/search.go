package model

type SearchResult struct {
	Offset     map[string]string `json:"offset,omitempty"`
	Assets     []Asset           `json:"assets,omitempty"`
	Attributes []Attribute       `json:"attributes,omitempty"`
	Risks      []Risk            `json:"risks,omitempty"`
	Files      []File            `json:"files,omitempty"`
	Jobs       []Job             `json:"jobs,omitempty"`
	Accounts   []Account         `json:"accounts,omitempty"`
	Term       string            `json:"-"`
}
