package capabilities

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
	"io"
	"log/slog"
	"os"
	"os/exec"
)

type Nuclei struct {
	Job    model.Job
	Asset  model.Asset
	Header string
	XYZ
}

type NucleiFinding struct {
	ID   string `json:"template-id"`
	Path string `json:"template-path"`
	Info struct {
		Description string   `json:"description"`
		Severity    string   `json:"severity"`
		Author      []string `json:"author"`
		Metadata    struct {
			Praetorian struct {
				Detection  bool              `json:"detection"`
				Attributes map[string]string `json:"attributes"`
			} `json:"praetorian"`
		} `json:"metadata"`
	} `json:"info"`
	IP   string `json:"ip"`
	Port string `json:"port"`
	URL  string `json:"matched-at"`
}

func NewNuclei(job model.Job) model.Capability {
	return &Nuclei{Asset: job.Target, Job: job, Header: hash(job.Username), XYZ: NewXYZ()}
}

func (task *Nuclei) Match() bool {
	return !task.Asset.Is(model.ActiveLow) && task.Asset.Is("http") && IsValidWebsite(task.Asset.Name)
}

func (task *Nuclei) Encode(path string) string {
	f, _ := os.Open(path)
	buf := &bytes.Buffer{}
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	io.Copy(enc, f)
	enc.Close()
	return buf.String()
}

func (task *Nuclei) Invoke() error {
	var parser = func(line string) {
		var finding NucleiFinding
		if err := json.Unmarshal([]byte(line), &finding); err != nil {
			slog.Error("nuclei finding parse", "error", err, "line", line)
			return
		}

		praetorian := finding.Info.Metadata.Praetorian
		a := model.NewAsset(task.Asset.DNS, finding.IP)
		for key := range praetorian.Attributes {
			task.Job.Stream <- a.Attribute(key, praetorian.Attributes[key])
		}

		if praetorian.Detection {
			return
		}

		risk := model.NewRisk(task.Asset, finding.ID)

		switch finding.Info.Severity {
		case "info":
			risk.Status = model.TriageInfo
		case "low":
			risk.Status = model.TriageLow
		case "medium":
			risk.Status = model.TriageMedium
		case "high":
			risk.Status = model.TriageHigh
		case "critical":
			risk.Status = model.TriageCritical
		}

		task.Job.Stream <- risk

		var m = map[string]any{}
		json.Unmarshal([]byte(line), &m)
		m["template-encoded"] = task.Encode(finding.Path)
		bites, _ := json.Marshal(m)

		task.Job.Stream <- risk.Proof(bites)
		task.Job.Stream <- risk.Attribute("asset", finding.URL)
		slog.Info("nuclei detection", "asset", task.Asset.Key, "risk", finding.ID, "ip", finding.IP)
	}

	args := []string{
		"-u", task.Asset.Name,
		"-es", "unknown",
		"-pt", "http,ssl,tcp",
		"-duc",
		"-silent",
		"-j",
		"-t", os.Getenv("NUCLEI_TEMPLATES"),
		"-no-stdin",
		"-no-mhe",
		"-c", "100",
		"-tags", fmt.Sprintf("case-reviewed,%s", task.Job.Username),
		"-H", fmt.Sprintf("Chariot: %s", task.Header),
		"-H", fmt.Sprintf("User-Agent: chariot-%s", task.Header),
	}
	template, rescan := task.Job.Config["test"]
	if rescan {
		args = append(args, "-id", template)
	} else {
		args = append(args, "-dialer-keep-alive", "5",
			"-timeout", "1",
			"-retries", "0",
		)
	}

	cmd := exec.Command("nuclei", args...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=/tmp/%s", os.TempDir()))
	cmd.Env = append(cmd.Env, "GOMEMLIMIT=1250MiB")

	return task.XYZ.Execute(cmd, parser)
}
