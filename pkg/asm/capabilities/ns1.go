package capabilities

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
	"log/slog"
	"net/http"
	"time"

	api "gopkg.in/ns1/ns1-go.v2/rest"
)

type NS1 struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

func NewNS1(job model.Job) model.Capability {
	return &NS1{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (task *NS1) Match() bool {
	_, ok := task.Asset.Config["ns1_api_key"]
	return ok
}

func (task *NS1) Invoke() error {
	client := task.client()

	zones, _, err := client.Zones.List()
	if err != nil {
		return err
	}

	for _, z := range zones {
		zone, _, err := client.Zones.Get(z.Zone, true)

		if err != nil {
			slog.Error("NS1: Failed to get zone details", "zone", z.Zone, "error", err)
			return err
		}
		for _, record := range zone.Records {
			switch record.Type {
			case "A", "AAAA":
				for _, ip := range record.ShortAns {
					task.Job.Stream <- model.NewAsset(record.Domain, ip)
				}
			case "CNAME":
				task.Job.Stream <- model.NewAsset(record.Domain, normalize(record.ShortAns[0], record.ShortAns[0]))
			}
		}
	}

	return nil
}

func (task *NS1) client() *api.Client {
	httpClient := &http.Client{Timeout: 3 * time.Second}
	client := api.NewClient(httpClient, api.SetAPIKey(task.Asset.Config["ns1_api_key"]))
	client.FollowPagination = true
	return client
}
