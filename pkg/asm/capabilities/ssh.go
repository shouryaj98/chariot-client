package capabilities

import (
	"fmt"
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
	"net"
	"net/url"
	"strings"
	"time"
)

type SSH struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

type Vulnerability func(url *url.URL) (bool, string)

var checks = map[string]Vulnerability{
	"CVE-2024-6387": CheckRegreSSHion,
}

func NewSSH(job model.Job) model.Capability {
	return &SSH{Asset: job.Target, Job: job}
}

func (task *SSH) Match() bool {
	return task.Asset.Is("ssh")
}

func (task *SSH) Invoke() error {
	url, _ := url.Parse(task.Asset.Name)
	for cve, check := range checks {
		vulnerable, banner := check(url)
		if vulnerable {
			risk := model.NewRisk(task.Asset, cve)
			risk.Status = model.TriageHigh

			task.Job.Stream <- risk
			task.Job.Stream <- risk.Proof([]byte(banner))
			task.Job.Stream <- risk.Attribute("asset", task.Asset.Name)
		}
	}

	return nil
}

func CheckRegreSSHion(url *url.URL) (bool, string) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", url.Hostname(), url.Port()), 5*time.Second)
	if err != nil {
		return false, ""
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	conn.Write([]byte("SSH-2.0-OpenSSH\r\n"))

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return false, ""
	}

	response := string(buffer[:n])
	versions := []string{
		"SSH-2.0-OpenSSH_8.5p1",
		"SSH-2.0-OpenSSH_8.6p1",
		"SSH-2.0-OpenSSH_8.7p1",
		"SSH-2.0-OpenSSH_8.8p1",
		"SSH-2.0-OpenSSH_8.9p1",
		"SSH-2.0-OpenSSH_9.0p1",
		"SSH-2.0-OpenSSH_9.1p1",
		"SSH-2.0-OpenSSH_9.2p1",
		"SSH-2.0-OpenSSH_9.3p1",
		"SSH-2.0-OpenSSH_9.4p1",
		"SSH-2.0-OpenSSH_9.5p1",
		"SSH-2.0-OpenSSH_9.6p1",
		"SSH-2.0-OpenSSH_9.7p1",
	}
	patchedVersions := []string{
		"SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.10",
		"SSH-2.0-OpenSSH_9.3p1 Ubuntu-3ubuntu3.6",
		"SSH-2.0-OpenSSH_9.6p1 Ubuntu-3ubuntu13.3",
		"SSH-2.0-OpenSSH_9.6p1 Ubuntu-3ubuntu13.4",
		"SSH-2.0-OpenSSH_9.3p1 Ubuntu-1ubuntu3.6",
		"SSH-2.0-OpenSSH_9.2p1 Debian-2+deb12u3",
		"SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u3",
		"SSH-2.0-OpenSSH_9.7p1 Debian-7",
		"SSH-2.0-OpenSSH_9.6 FreeBSD-20240701",
		"SSH-2.0-OpenSSH_9.7 FreeBSD-20240701",
	}
	for _, version := range patchedVersions {
		if strings.Contains(response, version) {
			return false, response
		}
	}
	for _, version := range versions {
		if strings.Contains(response, version) {
			return true, response
		}
	}
	return false, response
}
