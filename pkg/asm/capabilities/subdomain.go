package capabilities

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

type Subdomain struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

func NewSubdomain(job model.Job) model.Capability {
	return &Subdomain{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (task *Subdomain) Match() bool {
	return task.Asset.Is("domain") || task.Asset.Is("tld")
}

func (task *Subdomain) Invoke() error {
	wg := sync.WaitGroup{}
	resolve := func(domain string) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, ip := range task.XYZ.Resolve(domain) {
				task.Job.Stream <- model.NewAsset(domain, ip)
			}
		}()
	}
	resolve(task.Asset.DNS)

	var parser = func(line string) {
		subdomain := strings.TrimSpace(strings.TrimSuffix(line, "."))
		valid, _ := regexp.Match(`(?i)^.*\.`+task.Asset.DNS+"$", []byte(subdomain))
		if !valid {
			return
		}
		resolve(subdomain)
	}

	subfinder := exec.Command("subfinder", "-d", task.Asset.DNS, "-silent")
	task.XYZ.Execute(subfinder, parser)

	assetfinder := exec.Command("assetfinder", "-subs-only", task.Asset.DNS)
	task.XYZ.Execute(assetfinder, parser)
	wg.Wait()
	return nil
}
