package capabilities

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/google/go-github/v61/github"
)

type GithubDiscovery struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

func NewDiscovery(job model.Job) model.Capability {
	return &GithubDiscovery{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (task *GithubDiscovery) Match() bool {
	return task.Asset.Is("tld") && !task.Asset.System()
}

func (task *GithubDiscovery) Invoke() error {
	org := strings.Split(task.Asset.DNS, ".")[0]

	pat := task.Job.Config["secret"]
	client := github.NewClient(nil).WithAuthToken(pat)

	result, _, err := client.Search.Users(context.Background(), fmt.Sprintf("type:org %s", org), nil)
	if err != nil {
		return err
	}

	for _, user := range result.Users {
		org, _, err := client.Organizations.Get(context.Background(), *user.Login)
		if err != nil {
			continue
		}

		uri, err := url.Parse(org.GetBlog())
		if err != nil {
			continue
		}

		temp := model.NewAsset(uri.Host, uri.Host)
		if !temp.Is("tld") {
			continue
		}

		asset := model.NewAsset(org.GetHTMLURL(), org.GetHTMLURL())
		asset.Status = model.ActiveLow
		task.Job.Stream <- asset
	}

	return nil
}

func (a *GithubDiscovery) Secret() string {
	return "/pats/discovery"
}
