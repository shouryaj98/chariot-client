package capabilities

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

type Gato struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

type Runner struct {
	Name string `json:"name"`
}

type Repository struct {
	CanFork           bool     `json:"can_fork"`
	AccessibleRunners []Runner `json:"accessible_runners"`
	RunnerWorkflows   []string `json:"runner_workflows"`
}

type Enumeration struct {
	Repositories []Repository `json:"repositories"`
}

type GatoOutput struct {
	Enumeration Enumeration `json:"enumeration"`
}

func NewGato(job model.Job) model.Capability {
	return &Gato{Job: job, Asset: job.Target, XYZ: NewXYZ()}
}

func (task *Gato) Match() bool {
	low := task.Asset.Is(model.ActiveLow)
	class := task.Asset.Is("repository")
	return !low && class && strings.Contains(task.Asset.DNS, "github.com")
}

func (task *Gato) Invoke() error {
	var output GatoOutput

	resp, err := http.DefaultClient.Get(task.Asset.DNS)
	if err != nil || resp.StatusCode != 200 {
		return err
	}

	re := regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+)$`)
	parts := re.FindStringSubmatch(task.Asset.DNS)
	if len(parts) != 3 {
		return nil
	}
	name := fmt.Sprintf("%s/%s", parts[1], parts[2])

	bytes, err := task.exploit(name)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	json.Unmarshal(bytes, &output)
	for _, repo := range output.Enumeration.Repositories {
		if len(repo.RunnerWorkflows) > 0 && repo.CanFork {
			risk := model.NewRisk(task.Asset, "public-repo-with-self-hosted-runner")

			task.Job.Stream <- risk
			task.Job.Stream <- risk.Proof(bytes)
		}
	}

	return nil
}

func (task *Gato) Secret() string {
	return "/pats/gato"
}

func (task *Gato) Timeout() int {
	return 20
}

func (task *Gato) exploit(name string) ([]byte, error) {
	directory, _ := os.MkdirTemp("", fmt.Sprintf("gato.%d", time.Now().UnixMilli()))
	temp, err := os.CreateTemp(directory, "gato-*.json")
	if err != nil {
		return []byte{}, err
	}
	temp.Close()
	defer os.Remove(directory)
	defer os.Remove(temp.Name())

	cmd := exec.Command("gato", "enum", "--repository", name, "--output-json", temp.Name())
	token := task.Job.Config["secret"]
	if pat, ok := task.Asset.Config["pat"]; ok && !strings.HasPrefix(pat, "github_pat_") && strings.HasPrefix(pat, "ghp_") {
		token = pat
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("GH_TOKEN=%s", token))

	err = task.XYZ._Execute(cmd, func(line string) {})
	data, _ := os.ReadFile(temp.Name())
	return data, err
}
