package capabilities

import (
	"encoding/json"
	"fmt"
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
	"log/slog"
	"os"
	"os/exec"
	"time"
)

type Azure struct {
	Job       model.Job
	Asset     model.Asset
	directory string
	XYZ
}

type row struct {
	DNS   string   `json:"dns"`
	Name  string   `json:"name"`
	Names []string `json:"names"`
	Group string   `json:"group"`
	ID    string   `json:"id"`
}

type response struct {
	Data []row `json:"data"`
}

func (a *Azure) parser(config map[string]string, foreach func(row)) func(string) {
	return func(all string) {
		var data response
		err := json.Unmarshal([]byte(all), &data)
		if err != nil {
			slog.Error("failed to parse azure response", "error", err, "response", all)
			return
		}

		for _, row := range data.Data {
			for _, name := range append(row.Names, row.Name) {
				if name == "" {
					continue
				}
				asset := model.NewAsset(normalize(row.DNS, name), name)
				asset.Config = config
				a.Job.Stream <- asset
				a.Job.Stream <- asset.Attribute("cloud", row.ID)

				if foreach != nil {
					foreach(row)
				}
			}
		}
	}
}

func (a *Azure) command(arg ...string) *exec.Cmd {
	cmd := exec.Command("az", arg...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("AZURE_CONFIG_DIR=%s", a.directory))
	cmd.Env = append(cmd.Env, fmt.Sprintf("AZURE_EXTENSION_DIR=%s", fmt.Sprintf("%s/.azure/cliextensions", os.Getenv("HOME"))))
	return cmd
}

func (a *Azure) exec(cmd *exec.Cmd, parser func(string)) {
	if err := a._Execute(cmd, parser); err != nil {
		slog.Error("failed to execute azure command", "error", err, "username", a.Job.Username)
	}
}

func (a *Azure) query(query string, foreach func(row)) {
	cmd := a.command(
		"graph", "query",
		"-q", query,
		"--subscriptions", a.Asset.Name,
	)
	a.exec(cmd, a.parser(nil, foreach))
}

func NewAzure(job model.Job) model.Capability {
	return &Azure{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (a *Azure) Match() bool {
	_, ok1 := a.Asset.Config["name"]
	_, ok2 := a.Asset.Config["secret"]
	_, ok3 := a.Asset.Config["tenant"]
	return ok1 && ok2 && (ok3 || !a.Asset.System())
}

func (a *Azure) Invoke() error {

	a.directory, _ = os.MkdirTemp("", fmt.Sprintf("azure-.%d", time.Now().UnixMilli()))
	defer os.Remove(a.directory)
	if _, ok := a.Asset.Config["tenant"]; !ok {
		a.Asset.Config["tenant"] = a.Asset.Name
	}

	out, err := a.command(
		"login", "--service-principal",
		"--username", a.Asset.Config["name"],
		"--password", a.Asset.Config["secret"],
		"--tenant", a.Asset.Config["tenant"],
		"--allow-no-subscriptions",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to authenticate as service principal: %s: %s", err, string(out))
	}

	if !a.Asset.System() {
		a.accounts()
		return nil
	}

	out, err = a.command(
		"account", "set",
		"--subscription", a.Asset.Name,
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set subscription %s: %s: %s", a.Asset.Name, err, string(out))
	}

	a.ip()
	a.zone()
	a.site()

	return nil
}

func (a *Azure) ip() {
	a.query("where type == \"microsoft.network/publicipaddresses\" and "+
		"(isnotempty(properties.dnsSettings.fqdn) or isnotempty(properties.ipAddress)) | "+
		"project dns=properties.dnsSettings.fqdn,name=properties.ipAddress,id=id", nil,
	)
}

func (a *Azure) record(r row, t string, query string) {
	a.exec(a.command(
		"network", "dns",
		"record-set", t,
		"list",
		"--zone-name", r.DNS,
		"--resource-group", r.Group,
		"--query", query,
	), a.parser(nil, nil))
}

func (a *Azure) records(r row) {
	a.record(r, "a", "{ \"data\": [].{ \"dns\": fqdn, \"id\": id, \"names\": ARecords[].ipv4Address }}")
	a.record(r, "aaaa", "{ \"data\": [].{ \"dns\": fqdn, \"id\": id, \"names\": AAAARecords[].ipv6Address }}")
	a.record(r, "cname", "{ \"data\": [].{ \"dns\": fqdn, \"id\": id, \"name\": CNAMERecord.cname }}")
}

func (a *Azure) zone() {
	a.query("where type == \"microsoft.network/dnszones\" | project name=name,dns=name,group=resourceGroup,id=id", a.records)
}

func (a *Azure) site() {
	a.query("where type == \"microsoft.web/sites\" | mv-expand dns=properties.enabledHostNames | project name=name,dns=dns,id=id", nil)
}

func (a *Azure) accounts() {
	a.exec(a.command(
		"account", "list",
		"--only-show-errors",
		"--query", "{ data: [].{ name : id, dns: 'azure', id: id } }",
	), a.parser(a.Asset.Config, nil))
}

func (a *Azure) Timeout() int {
	return 45
}
