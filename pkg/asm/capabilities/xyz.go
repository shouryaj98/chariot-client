package capabilities

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

type Parser func(string)

type XYZ struct{}

func NewXYZ() XYZ {
	return XYZ{}
}

func (xyz *XYZ) Execute(cmd *exec.Cmd, parser Parser) error {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		slog.Error("failed to start", "command", append([]string{cmd.Path}, cmd.Args...), "error", err)
		return err
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer([]byte{}, 4*1024*1024)
	for scanner.Scan() {
		parser(scanner.Text())
	}
	err = cmd.Wait()
	if err != nil {
		out, _ := io.ReadAll(stderr)
		slog.Error("failed to execute", "command", append([]string{cmd.Path}, cmd.Args...), "error", err, "stderr", string(out))
	}
	return err
}

func (xyz *XYZ) _Execute(cmd *exec.Cmd, parser Parser) error {
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("failed to execute", "command", append([]string{cmd.Path}, cmd.Args...), "error", err, "output", string(output))
		return err
	}
	parser(string(output))
	return nil
}

func (xyz *XYZ) Resolve(host string) []string {
	resolver := &net.Resolver{}

	hits, err := resolver.LookupIPAddr(context.Background(), host)
	if err != nil {
		return []string{}
	}

	var addresses []string
	for _, ip := range hits {
		if ip.IP.IsPrivate() || ip.IP.IsLoopback() {
			continue
		}
		addresses = append(addresses, ip.IP.String())
	}
	return addresses
}

func normalize(s string, name string) string {
	matches := regexp.MustCompile(`(^|\/)([a-zA-Z0-9\-\.]+?)\.?(\#|\?|\/|$)`).FindStringSubmatch(s)
	if s == "" || len(matches) < 3 {
		return name
	}
	return matches[2]
}

func hash(s string) string {
	hasher := md5.New()
	io.WriteString(hasher, s)
	return hex.EncodeToString(hasher.Sum(nil))
}

func request[T any](method, url string, body []byte, headers ...string) (*T, error) {
	data := new(T)
	resp, err := _request(method, url, body, headers...)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(resp, data); err != nil {
		return nil, err
	}

	return data, nil
}

func _request(method, url string, body []byte, headers ...string) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func IsValidWebsite(target string) bool {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Head(target)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	unavailable := resp.StatusCode == http.StatusServiceUnavailable
	redirect := (resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusFound)
	mirror := strings.HasPrefix(resp.Header.Get("Location"), fmt.Sprintf("https://%s", resp.Request.URL.Hostname()))
	http := resp.Request.URL.Scheme == "http"

	return !(unavailable || (redirect && mirror && http))
}

func (xyz *XYZ) Secret() string {
	return ""
}

func (xyz *XYZ) Timeout() int {
	return 10
}
