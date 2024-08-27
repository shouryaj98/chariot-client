package capabilities

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/PuerkitoBio/goquery"
)

var domainRegex = regexp.MustCompile(`\b((?:[\w-]+\.)+com)\b`)

const (
	HTML_API   = "https://www.sec.gov"
	JSON_API   = "https://data.sec.gov"
	USER_AGENT = "research@praetorian.com"
)

type filingsData struct {
	AccessionNumber    []string `json:"accessionNumber"`
	FilingDate         []string `json:"filingDate"`
	AcceptanceDateTime []string `json:"acceptanceDateTime"`
	Form               []string `json:"form"`
	PrimaryDocument    []string `json:"primaryDocument"`
}

type submissionFiling struct {
	Recent filingsData `json:"recent"`
}

type submissionsResult struct {
	Filings submissionFiling `json:"filings"`
}

type Filing struct {
	AccessionNumber string `json:"accessionNumber"`
	FilingDate      string `json:"filingDate"`
	Form            string `json:"form"`
	PrimaryDocument string `json:"primaryDocument"`
	Text            string `json:"text"`
}

type Edgar struct {
	Stream chan interface{}
	Asset  model.Asset
	XYZ
}

func NewEdgar(job model.Job) model.Capability {
	return &Edgar{Asset: job.Target, Stream: job.Stream, XYZ: NewXYZ()}
}

func (task *Edgar) Match() bool {
	return task.Asset.Is("tld") && !task.Asset.System()
}

func (task *Edgar) Invoke() error {
	return task.Adjacent(func(f Filing) bool {
		today := time.Now().Format("2006-01-02")
		return strings.HasPrefix(f.FilingDate, today)
	})
}

func (task *Edgar) Register() error {
	return task.Adjacent(func(f Filing) bool {
		return strings.Contains(f.Form, "10-K") && strings.HasPrefix(f.FilingDate, "201")
	})
}

func (task *Edgar) Adjacent(filterFunc func(Filing) bool) error {
	company := strings.Split(task.Asset.DNS, ".")[0]
	domains := make(map[string]struct{})

	ciks, err := searchCIK(company)
	if err != nil {
		return err
	}

	for _, cik := range ciks {
		filings, err := getFilingsForCIK(cik)
		if err != nil {
			return err
		}

		filtered := make([]*Filing, 0)
		for _, f := range filings {
			if filterFunc(*f) {
				filtered = append(filtered, f)
			}
		}

		err = downloadFilings(filtered, cik)
		if err != nil {
			continue
		}

		for _, f := range filtered {
			if !strings.Contains(strings.ToUpper(f.Text), strings.ToUpper(task.Asset.DNS)) {
				continue
			}

			matches := domainRegex.FindAllString(f.Text, -1)
			for _, match := range matches {
				split := strings.Split(strings.ToLower(match), ".")
				match := split[len(split)-2] + "." + split[len(split)-1]
				domains[match] = struct{}{}
			}
		}
	}

	isForbidden := func(domain string) bool {
		for _, d := range []string{"dfinsolutions"} {
			if strings.Contains(domain, d) {
				return true
			}
		}
		return false
	}
	for domain := range domains {
		if !isForbidden(domain) {
			asset := model.NewAsset(domain, domain)
			asset.Status = model.FrozenLow
			task.Stream <- asset
		}
	}

	return nil
}

func downloadFilings(fs []*Filing, cik string) error {
	stripped := strings.TrimLeft(cik, "0")
	for _, f := range fs {
		u, err := url.Parse(HTML_API)
		if err != nil {
			return err
		}

		accessionNumber := strings.ReplaceAll(f.AccessionNumber, "-", "")
		u = u.JoinPath("Archives", "edgar", "data", stripped, accessionNumber, f.PrimaryDocument)
		body, err := _request("GET", u.String(), nil, "User-Agent", USER_AGENT)
		if err != nil {
			return err
		}

		f.Text = string(body)
	}

	return nil
}

func searchCIK(company string) ([]string, error) {
	u, err := url.Parse(HTML_API)
	if err != nil {
		return nil, err
	}

	u = u.JoinPath("/cgi-bin/cik_lookup")
	data := url.Values{}
	data.Set("company", company)
	body, err := _request("POST", u.String(), []byte(data.Encode()), "User-Agent", USER_AGENT)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var ciks []string
	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		row := s.Find("tr")
		row.Find("a").Each(func(i int, s *goquery.Selection) {
			if !strings.Contains(s.Text(), "Perform another Company-CIK Lookup") {
				ciks = append(ciks, s.Text())
			}
		})
	})

	return ciks, nil
}

func getFilingsForCIK(cik string) ([]*Filing, error) {
	u, err := url.Parse(JSON_API)
	if err != nil {
		return nil, err
	}

	format := fmt.Sprintf("CIK%s.json", cik)
	u = u.JoinPath("/submissions", format)
	sr, err := request[submissionsResult]("GET", u.String(), nil, "User-Agent", USER_AGENT)
	if err != nil {
		return nil, err
	}

	return invertSubmissions(*sr), nil
}

func invertSubmissions(sr submissionsResult) []*Filing {
	f := sr.Filings.Recent
	var filings []*Filing

	for i := 0; i < len(f.AcceptanceDateTime); i++ {
		fil := &Filing{
			AccessionNumber: f.AccessionNumber[i],
			FilingDate:      f.FilingDate[i],
			Form:            f.Form[i],
			PrimaryDocument: f.PrimaryDocument[i],
		}

		filings = append(filings, fil)
	}

	return filings
}
