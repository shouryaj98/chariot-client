package capabilities

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

type Secrets struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

type NPLine struct {
	FindingID string `json:"finding_id"`
	RuleID    string `json:"rule_text_id"`
}

func NewSecrets(job model.Job) model.Capability {
	return &Secrets{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (task *Secrets) Match() bool {
	low := task.Asset.Is(model.ActiveLow)
	return !low && task.Asset.Is("repository")
}

func (task *Secrets) Invoke() error {
	nilparser := func(line string) {}
	parser := func(line string) {
		var npLine NPLine
		if err := json.Unmarshal([]byte(line), &npLine); err == nil {
			if strings.Contains(strings.ToLower(line), "no-nosey-parker") {
				return
			}
			asset := model.NewAsset(fmt.Sprintf("%s:%s", task.Asset.DNS, npLine.FindingID[:8]), task.Asset.Name)
			risk := model.NewRisk(asset, fmt.Sprintf("git-secrets-%s", strings.Split(npLine.RuleID, ".")[1]))

			task.Job.Stream <- risk
			task.Job.Stream <- risk.Proof([]byte(line))
		}
	}

	parts := strings.Split(task.Asset.DNS, "://")
	url := fmt.Sprintf("%s://user:%s@%s", parts[0], task.Asset.Config["pat"], parts[1])
	repo := fmt.Sprintf("/tmp/%d.clone", time.Now().UnixMilli())
	exec.Command("git", "clone", "--bare", url, repo).Run()
	// remove git config file: it contains the PAT used for cloning and we don't want to report that
	os.Remove(fmt.Sprintf("%s/config", repo))
	defer os.RemoveAll(repo)

	store := fmt.Sprintf("/tmp/%d.noseyparker", time.Now().UnixMilli())
	defer os.RemoveAll(store)
	cmd := exec.Command(
		"noseyparker",
		"scan",
		"--datastore", store,
		repo,
	)
	task.XYZ.Execute(cmd, nilparser)

	cmd = exec.Command(
		"noseyparker",
		"report",
		"--datastore", store,
		"-f", "jsonl",
	)
	return task.XYZ.Execute(cmd, parser)
}
