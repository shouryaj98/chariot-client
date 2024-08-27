package capabilities

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

var HONEYPOT = 25

type PortScan struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

func NewPortScan(job model.Job) model.Capability {
	return &PortScan{
		Asset: job.Target,
		Job:   job,
		XYZ:   NewXYZ(),
	}
}

func (task *PortScan) Match() bool {
	low := task.Asset.Is(model.ActiveLow)
	return !low && (task.Asset.Is("ip") || task.Asset.Is("cidr"))
}

func (task *PortScan) Invoke() error {
	var wg sync.WaitGroup
	var err error
	addresses := map[string][]string{}

	switch {
	case task.Asset.Is("ipv6"):
		err = task.nmap(addresses)
	case task.Asset.Is("ipv4"):
		err = task.masscan(addresses, "5000")
	case task.Asset.Is("cidr"):
		var network *net.IPNet
		_, network, err = net.ParseCIDR(task.Asset.Name)
		if err != nil {
			return err
		}

		if size, _ := network.Mask.Size(); size >= 24 {
			err = task.masscan(addresses, "50000")
			break
		}

		for _, cidr := range split(task.Asset.Name) {
			child := model.NewAsset(cidr, cidr)
			child.Key = task.Asset.Key
			task.Job.Stream <- model.NewJob("portscan", child)
		}
	}
	if err != nil {
		return err
	}

	for ip, ports := range addresses {
		asset := task.Job.Target

		if task.Asset.Is("cidr") {
			asset = task.Job.Target.Spawn(ip, ip)
			task.Job.Stream <- asset
		}

		if len(ports) > HONEYPOT {
			task.Job.Stream <- asset.Attribute("interesting", "honeypot")
			continue
		}

		for _, port := range ports {
			wg.Add(1)
			go task.fingerprintx(&wg, asset, port)
		}
	}

	wg.Wait()
	return err
}

func (task *PortScan) nmap(addresses map[string][]string) error {
	parser := func(line string) {
		re := regexp.MustCompile(`(\d+)/tcp\s+\w+\s+\w+`)
		match := re.FindStringSubmatch(line)
		if len(match) > 1 {
			addresses[task.Asset.Name] = append(addresses[task.Asset.Name], match[1])
		}
	}

	cmd := exec.Command("nmap", "-Pn", "--host-timeout", "1m", "--max-retries", "0", task.Asset.Name, "-6")
	return task.XYZ.Execute(cmd, parser)
}

func (task *PortScan) masscan(addresses map[string][]string, rate string) error {
	parser := func(line string) {
		re := regexp.MustCompile(`(\d+)\/tcp on (\d+\.\d+\.\d+\.\d+)`)
		match := re.FindStringSubmatch(line)
		if len(match) == 3 {
			ip := match[2]
			addresses[ip] = append(addresses[ip], match[1])
		}
	}

	cmd := exec.Command("masscan", "-p-", fmt.Sprintf("--rate=%s", rate), task.Asset.Name)
	return task.XYZ.Execute(cmd, parser)
}

func (task *PortScan) fingerprintx(wg *sync.WaitGroup, asset model.Asset, port string) {
	defer wg.Done()

	var result map[string]any

	cmd := exec.Command("fingerprintx", "-t", fmt.Sprintf("%s:%s", asset.Name, port), "--json")
	task.XYZ._Execute(cmd, func(line string) {
		json.Unmarshal([]byte(line), &result)
	})

	task.Job.Stream <- asset.Attribute("port", port)

	protocol, ok := result["protocol"].(string)
	if !ok || protocol == "" {
		return
	}

	task.Job.Stream <- asset.Attribute(strings.ToLower(protocol), port)
	task.Job.Stream <- asset.Attribute("protocol", strings.ToLower(protocol))
}

func split(cidr string) []string {
	var subnets []string

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return subnets
	}

	size, _ := ipNet.Mask.Size()
	if size > 24 {
		return []string{cidr}
	}

	ranges := 1 << (24 - size)
	ip := ipNet.IP

	for i := 0; i < ranges; i++ {
		ip := net.IPv4(ip[0], ip[1], ip[2]+byte(i), 0)
		subnets = append(subnets, fmt.Sprintf("%s/24", ip))
	}

	return subnets
}
