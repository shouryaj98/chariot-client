package capabilities

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/gocolly/colly/v2"
)

type Crawler struct {
	Job    model.Job
	Asset  model.Asset
	Header string
	XYZ
}

func NewCrawler(job model.Job) model.Capability {
	return &Crawler{Asset: job.Target, Job: job, Header: hash(job.Username), XYZ: NewXYZ()}
}

func (task *Crawler) Match() bool {
	webserver := strings.HasPrefix(task.Asset.Name, "http")
	return webserver && task.Asset.Is(model.ActiveHigh)
}

func (task *Crawler) Invoke() error {
	crawl := task.Crawler()
	acceptsInput := task.FindInputFields(crawl)
	containsSecrets := task.FindSecrets(crawl)

	crawl.Visit(task.Asset.Name)

	acceptsInput()
	containsSecrets()

	return nil
}

func (task *Crawler) Crawler() *colly.Collector {
	regex := regexp.MustCompile("^https?://(?:[a-z0-9-]+\\.)*" + regexp.QuoteMeta(task.Asset.DNS))
	allowed := func(link string, mime string) bool {
		return regex.Match([]byte(link)) && (strings.HasPrefix(mime, "text/") || mime == "" || strings.HasSuffix(link, "/"))
	}

	c := colly.NewCollector(colly.URLFilters(regex), colly.UserAgent(fmt.Sprintf("Chariot: %s", task.Header)), colly.MaxBodySize(1024*1024))
	c.SetRequestTimeout(1 * time.Second)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if allowed(link, mime.TypeByExtension(path.Ext(link))) {
			if v, err := c.HasVisited(link); err == nil && !v {
				c.Visit(link)
			}
		}
	})
	c.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
		if !regex.MatchString(req.URL.String()) {
			return colly.ErrNoURLFiltersMatch
		}
		return nil
	})

	return c
}

func (task *Crawler) FindInputFields(c *colly.Collector) func() {
	var urls []string

	c.OnHTML("input", func(e *colly.HTMLElement) {
		params, _ := url.ParseQuery(e.Request.URL.RawQuery)
		if len(params) != 0 {
			for key := range params {
				params.Set(key, "FUZZ")
			}
			e.Request.URL.RawQuery = params.Encode()
			urls = append(urls, e.Request.URL.String())
		}
	})

	return func() {
		if slices.Sort(urls); len(urls) == 0 {
			return
		}

		scan := exec.Command("dalfox", strings.Fields("pipe --timeout 3 --report-format json --only-poc=v --format json --skip-bav --silence")...)
		scan.Stdin = strings.NewReader(strings.Join(slices.Compact(urls), "\n"))

		var results []struct {
			Param     string `json:"param"`
			Message   string `json:"message_str"`
			MessageID int    `json:"message_id"`
			Data      string `json:"data"`
		}

		task.XYZ._Execute(scan, func(line string) {
			if err := json.Unmarshal([]byte(line), &results); err == nil {
				for _, result := range results {
					if result.MessageID == 0 {
						continue
					}
					risk := model.NewRisk(task.Asset, fmt.Sprintf("xss-identified:%s", result.Param))
					risk.Status = model.TriageMedium

					if proof, err := json.MarshalIndent(result, "", "\t"); err == nil {
						task.Job.Stream <- risk
						task.Job.Stream <- risk.Proof([]byte(proof))
					}
				}
			}
		})
	}
}

func (task *Crawler) FindSecrets(c *colly.Collector) func() {
	type secretFinding struct {
		Rule    string   `json:"rule"`
		Matches []string `json:"matches"`
	}

	type Finding struct {
		RuleName   string `json:"rule_name"`
		NumMatches int    `json:"num_matches"`
		Matches    []struct {
			Provenance []struct {
				Path string `json:"path"`
			} `json:"provenance"`
			Snippet struct {
				Matching string `json:"matching"`
			} `json:"snippet"`
		} `json:"matches"`
	}

	tmp, _ := os.MkdirTemp("/tmp", fmt.Sprintf("crawler.%d", time.Now().UnixMilli()))
	crawl, store := tmp+"/crawl/", tmp+"/store/"
	cache := make(map[string]map[string]*secretFinding)
	size := 0
	count := 0

	c.OnResponse(func(r *colly.Response) {
		if !strings.HasPrefix(http.DetectContentType(r.Body), "text/") {
			return
		}

		path := base64.StdEncoding.EncodeToString([]byte(r.Request.URL.Path))
		dir := fmt.Sprintf("%s%s", crawl, r.Request.URL.Host)
		if err := os.MkdirAll(dir, os.ModePerm); err == nil && size+len(r.Body) < 10*1024*1024*1024 { // 10 GiB
			size += len(r.Body)
			os.WriteFile(fmt.Sprintf("%s/%s", dir, path), r.Body, os.ModePerm)
		}
	})

	extractMeta := func(prov string) (string, string) {
		parts := strings.Split(strings.TrimPrefix(prov, crawl), "/")
		if len(parts) == 2 {
			if location, err := base64.StdEncoding.DecodeString(parts[1]); err == nil {
				return parts[0], string(location)
			}
		}
		return "", ""
	}

	parser := func(line string) {
		var npLine Finding
		json.Unmarshal([]byte(line), &npLine)

		for _, match := range npLine.Matches {
			for _, provenance := range match.Provenance {
				dns, location := extractMeta(provenance.Path)
				if dns == "" || location == "" {
					continue
				}

				matched := match.Snippet.Matching
				if _, ok := cache[dns]; !ok {
					cache[dns] = make(map[string]*secretFinding)
				}
				if _, ok := cache[dns][matched]; !ok {
					cache[dns][matched] = &secretFinding{Rule: npLine.RuleName}
					count++
				}
				cache[dns][matched].Matches = append(cache[dns][matched].Matches, location)
			}
		}
	}

	return func() {
		defer os.RemoveAll(tmp)

		scan := exec.Command("noseyparker", "scan", "--datastore", store, crawl)
		report := exec.Command("noseyparker", "report", "--datastore", store, "-f", "jsonl")
		task.XYZ._Execute(scan, func(line string) {})
		task.XYZ.Execute(report, parser)

		if data, err := json.MarshalIndent(cache, "", "\t"); err == nil && len(cache) > 0 {
			risk := model.NewRisk(task.Asset, "web-secrets")
			risk.Status = model.TriageMedium

			task.Job.Stream <- risk
			task.Job.Stream <- risk.Proof(data)
		}
	}
}
