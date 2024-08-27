package capabilities

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

type Gitlab struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

func NewGitlab(job model.Job) model.Capability {
	return &Gitlab{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (task *Gitlab) Match() bool {
	return task.Asset.Is("gitlab") && task.Asset.Config["pat"] != ""
}

func (task *Gitlab) Invoke() error {
	re := regexp.MustCompile(`^(https://)?gitlab\.com/([^/]+)/?$`)
	matches := re.FindStringSubmatch(task.Asset.Name)
	for page := 0; ; page++ {
		if remaining, err := task.request(matches[len(matches)-1], page); err != nil || !remaining {
			return err
		}
	}
}

func (task *Gitlab) request(group string, page int) (bool, error) {
	api := fmt.Sprintf(`/v4/groups/%s/projects?order_by=created_at&page=%d&pagination=keyset&per_page=100&include_subgroups=true`, group, page)
	req, err := http.NewRequest("GET", `https://gitlab.com/api`+api, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("PRIVATE-TOKEN", task.Asset.Config["pat"])
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	return resp.Header.Get("x-next-page") != "", task.parse(body)
}

func (task *Gitlab) parse(body []byte) error {
	var projects []struct {
		Name string `json:"name"`
		URL  string `json:"web_url"`
	}
	if err := json.Unmarshal(body, &projects); err != nil {
		return err
	}
	for _, project := range projects {
		asset := model.NewAsset(project.URL, project.Name)
		asset.Config = task.Asset.Config
		task.Job.Stream <- asset
	}
	return nil
}
