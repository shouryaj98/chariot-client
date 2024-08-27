package capabilities

import (
	"fmt"
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

type Cloudflare struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

type Info struct {
	Count int `json:"count"`
	Page  int `json:"page"`
}

type Zone struct {
	ID string `json:"id"`
}

type Zones struct {
	Info   Info   `json:"result_info"`
	Result []Zone `json:"result"`
}

type Record struct {
	Content string `json:"content"`
	Name    string `json:"name"`
	Type    string `json:"type"`
}

type Records struct {
	Info   Info     `json:"result_info"`
	Result []Record `json:"result"`
}

func NewCloudflare(job model.Job) model.Capability {
	return &Cloudflare{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (task *Cloudflare) Match() bool {
	_, ok := task.Asset.Config["token"]
	return task.Asset.Is("cloudflare") && ok
}

func (task *Cloudflare) zones(headers ...string) ([]Zone, error) {
	page := 1
	zones := []Zone{}
	for {
		resp, err := request[Zones]("GET", fmt.Sprintf("https://api.cloudflare.com/client/v4/zones?page=%v", page), nil, headers...)
		if err != nil {
			return nil, err
		}
		zones = append(zones, resp.Result...)
		if resp.Info.Count == 0 {
			break
		}
		page++
	}
	return zones, nil
}

func (task *Cloudflare) zone(id string, headers ...string) error {
	page := 1
	for {
		resp, err := request[Records]("GET", fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?page=%v", id, page), nil, headers...)
		if err != nil {
			return err
		}
		for _, r := range resp.Result {
			switch r.Type {
			case "A", "AAAA":
				task.Job.Stream <- model.NewAsset(r.Name, r.Content)
			case "CNAME":
				task.Job.Stream <- model.NewAsset(normalize(r.Name, r.Name), normalize(r.Content, r.Content))
			}
		}
		if resp.Info.Count == 0 {
			break
		}
		page++
	}
	return nil
}

func (task *Cloudflare) Invoke() error {
	headers := []string{"Authorization", fmt.Sprintf("Bearer %s", task.Asset.Config["token"])}

	zones, err := task.zones(headers...)
	if err != nil {
		return err
	}

	for _, z := range zones {
		err = task.zone(z.ID, headers...)
		if err != nil {
			return err
		}
	}
	return nil
}
