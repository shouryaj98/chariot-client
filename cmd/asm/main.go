package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"

	"github.com/praetorian-inc/chariot-client/pkg/asm/capabilities"
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

var Registry = map[string]func(job model.Job) model.Capability{
	"nuclei":            capabilities.NewNuclei,
	"whois":             capabilities.NewWhois,
	"subdomain":         capabilities.NewSubdomain,
	"portscan":          capabilities.NewPortScan,
	"github":            capabilities.NewGit,
	"secrets":           capabilities.NewSecrets,
	"amazon":            capabilities.NewAmazon,
	"azure":             capabilities.NewAzure,
	"gcp":               capabilities.NewGCP,
	"ns1":               capabilities.NewNS1,
	"gato":              capabilities.NewGato,
	"crowdstrike":       capabilities.NewCrowdstrike,
	"crawler":           capabilities.NewCrawler,
	"gitlab":            capabilities.NewGitlab,
	"ssh":               capabilities.NewSSH,
	"github-discovery":  capabilities.NewDiscovery,
	"azuread-discovery": capabilities.NewAzureAD,
	"edgar":             capabilities.NewEdgar,
}

func process(job model.Job) []string {
	results := []string{}
	job.Stream = make(chan interface{})
	defer close(job.Stream)

	go func() error {
		for item := range job.Stream {
			iJson, _ := json.Marshal(item)
			results = append(results, string(iJson))
			switch i := item.(type) {
			case model.Asset:
				slog.Info("ASSET FOUND", "asset", i.Key)
			case model.Risk:
				slog.Info("RISK FOUND", "risk", i.Key)
			case model.Attribute:
				slog.Info("ATTRIBUTE FOUND", "attribute", i.Key)
			case model.File:
				slog.Info("FILE CREATED", "file", i.Key)
			case model.Job:
				slog.Info("JOB CREATED", "job", i.Key)
				process(i)
			}
		}
		return nil
	}()

	task := Registry[job.Source]
	if err := task(job).Invoke(); err != nil {
		job.Update(model.Fail)
		slog.Error("job failed", "composite", job.Key, "error", err, "username", job.Username)
	}

	return results
}

func main() {
	capability := flag.String("capability", "", "Capability type")
	name := flag.String("name", "", "Target name")
	flag.Parse()

	if *capability == "" || *name == "" {
		fmt.Println("Error: both --capability and --name flags are required.")
		return
	}

	if Registry[*capability] == nil {
		fmt.Printf("Invalid capability: %s\n", *capability)
		return
	}

	job := model.NewJob(*capability, model.NewAsset(*name, *name))
	process(job)
}
