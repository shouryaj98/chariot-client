package capabilities

import (
	"fmt"
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

type GCP struct {
	Job       model.Job
	Asset     model.Asset
	directory string
	XYZ
}

func NewGCP(job model.Job) model.Capability {
	return &GCP{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (g *GCP) Match() bool {
	_, ok1 := g.Asset.Config["keyfile"]
	return ok1
}

func (g *GCP) Invoke() error {
	g.directory, _ = os.MkdirTemp("", fmt.Sprintf("gcp.%d", time.Now().UnixMilli()))
	f, err := os.CreateTemp(g.directory, "gcp-*.json")
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(g.directory)

	f.Write([]byte(g.Asset.Config["keyfile"]))
	err = g.command("auth", "activate-service-account", "--key-file", f.Name()).Run()
	if err != nil {
		return fmt.Errorf("failed to activate service account: %s", err)
	}

	if !g.Asset.System() {
		g.projects()
	}
	g.compute()
	g.functions()
	g.forwarding()
	g.zone()

	return nil
}

func (g *GCP) parser(config map[string]string, foreach func(string, string, string)) func(string) {
	return func(line string) {
		if line == "\t\t" {
			return
		}

		v := strings.Split(line, "\t")
		asset := model.NewAsset(normalize(v[0], v[1]), v[1])
		asset.Config = config
		g.Job.Stream <- asset
		g.Job.Stream <- asset.Attribute("cloud", v[2])

		if foreach != nil {
			foreach(v[0], v[1], v[2])
		}
	}
}

func (g *GCP) command(args ...string) *exec.Cmd {
	cmd := exec.Command("gcloud", args...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("CLOUDSDK_CONFIG=%v", g.directory))
	cmd.Args = append(cmd.Args, "--project", g.Asset.Name, "--quiet", "--verbosity=error")
	return cmd
}

func (g *GCP) exec(cmd *exec.Cmd, parser func(string)) {
	err := g.XYZ.Execute(cmd, parser)
	if err != nil {
		slog.Error("failed to execute gcloud", "error", err, "args", cmd.Args)
	}
}

func (g *GCP) projects() {
	g.exec(g.command("projects",
		"list",
		"--format", "value(format(\"gcp\t{}\t{}\", projectId, uri()))",
	), g.parser(g.Asset.Config, nil))
}

func (g *GCP) compute() {
	g.exec(g.command("compute",
		"instances", "list",
		"--flatten", "networkInterfaces, networkInterfaces.accessConfigs",
		"--filter", "networkInterfaces.accessConfigs.natIP:*",
		"--format",
		"value(networkInterfaces.accessConfigs.publicPtrDomainName, networkInterfaces.accessConfigs.natIP, uri())",
	), g.parser(nil, nil))
	g.exec(g.command("compute",
		"instances", "list",
		"--flatten", "networkInterfaces, networkInterfaces.ipv6AccessConfigs",
		"--filter", "networkInterfaces.ipv6AccessConfigs.externalIpv6:*",
		"--format",
		"value(networkInterfaces.accessConfigs.publicPtrDomainName, networkInterfaces.ipv6AccessConfigs.externalIpv6, uri())",
	), g.parser(nil, nil))
}

func (g *GCP) functions() {
	g.exec(g.command("functions",
		"list",
		"--format", fmt.Sprintf("value(format(\"{}\t{}\tprojects/%s/functions/{}\",name,name,name))", g.Asset.Name),
	), g.parser(nil, nil))
}

func (g *GCP) forwarding() {
	g.exec(g.command("compute",
		"forwarding-rules", "list",
		"--format", "value(IPAddress, IPAddress, uri())",
	), g.parser(nil, nil))
}

func (g *GCP) recordset(_, name, uri string) {
	g.exec(g.command("dns",
		"record-sets", "list",
		"--zone", name,
		"--flatten", "rrdatas",
		"--format", fmt.Sprintf("value(format(\"{}\t{}\t%s\", name, rrdatas))", uri),
		"--filter", "type=A OR type=AAAA OR type=CNAME",
	), g.parser(nil, nil))
}

func (g *GCP) zone() {
	g.exec(g.command("dns",
		"managed-zones", "list",
		"--format", "value(dnsName, name, uri())",
	), g.parser(nil, g.recordset))
}

func (g *GCP) Timeout() int {
	return 45
}
