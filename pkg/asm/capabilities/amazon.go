package capabilities

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"

	"golang.org/x/sync/errgroup"
)

var regions = []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2", "ap-south-1", "eu-north-1", "eu-west-3", "eu-west-2", "eu-west-1", "ap-northeast-3", "ap-northeast-2", "ap-northeast-1", "ca-central-1", "sa-east-1", "ap-southeast-1", "ap-southeast-2", "eu-central-1"}

type Credentials struct {
	Id    string `json:"AccessKeyId"`
	Key   string `json:"SecretAccessKey"`
	Token string `json:"SessionToken"`
}

type Amazon struct {
	Job     model.Job
	Asset   model.Asset
	Account string
	Region  string
	Credentials
	XYZ
}

func NewAmazon(job model.Job) model.Capability {
	return &Amazon{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (a *Amazon) Send(account string) {
	for _, region := range regions {
		asset := model.NewAsset("amazon", fmt.Sprintf("%s:%s", account, region))
		asset.Config = a.Asset.Config
		a.Job.Stream <- asset
	}
}

func (a *Amazon) Match() bool {
	return a.Asset.Is("amazon") && a.Asset.Name != ""
}

func (a *Amazon) Timeout() int {
	return 45
}

func (a *Amazon) Invoke() error {
	a.Account = a.Asset.Name
	if a.Asset.System() {
		v := strings.Split(a.Asset.Name, ":")
		a.Account, a.Region = v[0], v[1]
	}

	role := "Chariot"
	if provided, ok := a.Asset.Config["role"]; ok {
		role = provided
	}
	cmd := exec.Command("aws", "sts", "assume-role",
		"--role-arn", fmt.Sprintf("arn:aws:iam::%s:role/%s", a.Account, role),
		"--role-session-name", "chariot",
		"--external-id", a.Job.Username,
		"--query", "Credentials",
		"--region", "us-east-1",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to assume role: %s: %s", err, string(out))
	}
	if err = json.Unmarshal(out, &a.Credentials); err != nil {
		return err
	}

	if !a.Asset.System() {
		a.Send(a.Account)
		a.accounts()
		return nil
	}

	a.network()
	a.lambda()
	a.gateway()
	if a.Region == regions[0] {
		a.route53()
	}
	return nil
}

func (a *Amazon) output(asset model.Asset, arn string) {
	a.Job.Stream <- asset
	a.Job.Stream <- asset.Attribute("cloud", arn)
}

func (a *Amazon) lambda() {
	g := errgroup.Group{}
	g.SetLimit(10)
	parser := func(funcName string) {
		cfg, _ := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(a.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(a.Id, a.Key, a.Token)))
		client := lambda.NewFromConfig(cfg)

		input := lambda.GetFunctionUrlConfigInput{
			FunctionName: aws.String(funcName),
			Qualifier:    nil,
		}
		output, err := client.GetFunctionUrlConfig(context.Background(), &input)
		if err != nil {
			slog.Warn("failed to get function url config", "error", err, "name", funcName)
			return
		}
		if output.FunctionUrl != nil {
			funcUrl := strings.Trim(strings.TrimPrefix(*output.FunctionUrl, "https://"), "/")
			arn := fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", a.Region, a.Account, funcName)
			a.output(model.NewAsset(funcUrl, funcName), arn)
		}
	}

	a.exec(parser,
		"lambda", "list-functions",
		"--region", a.Region,
		"--query", "Functions[*].[FunctionName]",
		"--output", "text",
	)
	g.Wait()
}

func (a *Amazon) gateway() {
	parser := func(line string) {
		id := strings.Split(line, "\t")[0]
		stage := strings.Split(line, "\t")[1]
		dns := fmt.Sprintf("%s.execute-api.%s.amazonaws.com/%s", id, a.Region, stage)
		arn := fmt.Sprintf("arn:aws:apigateway:%s::/restapis/%s", a.Region, id)
		a.output(model.NewAsset(dns, id), arn)
	}

	a.exec(parser,
		"apigateway", "get-rest-apis",
		"--region", a.Region,
		"--query", "items[*].[id, name]",
		"--output", "text",
	)
}

// This covers: Elastic IP, ELB, RDS, ECS, EKS, VPC
func (a *Amazon) network() {
	parser := func(line string) {
		ip := strings.Split(line, "\t")[0]
		dns := strings.Split(line, "\t")[1]
		id := strings.Split(line, "\t")[2]
		arn := fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", a.Region, a.Account, id)
		a.output(model.NewAsset(dns, ip), arn)
	}

	a.exec(parser,
		"ec2", "describe-instances",
		"--region", a.Region,
		"--query", "Reservations[*].Instances[?PublicIpAddress != null && PublicDnsName != null].[PublicIpAddress, PublicDnsName, InstanceId]",
		"--output", "text",
	)
}

func (a *Amazon) route53() {
	type ZoneRecords []struct {
		Name            string `json:"Name"`
		Type            string `json:"Type"`
		ResourceRecords []struct {
			Value string `json:"Value"`
		}
	}

	parseZones := func(line string) {
		match := regexp.MustCompile(`(\w+)"`).FindStringSubmatch(line)
		if len(match) < 1 {
			return
		}

		routesCmd := exec.Command(
			"aws",
			"route53", "list-resource-record-sets",
			"--hosted-zone-id", match[1],
			"--query", "ResourceRecordSets[?Type == 'A' || Type == 'AAAA' || Type == 'CNAME']",
		)
		a.session(routesCmd)
		output, err := routesCmd.CombinedOutput()

		arn := fmt.Sprintf("arn:aws:route53::%s:hostedzone/%s", a.Account, match[1])
		if err == nil {
			var zr ZoneRecords
			if err := json.Unmarshal(output, &zr); err != nil {
				return
			}
			for _, record := range zr {
				for _, resource := range record.ResourceRecords {
					a.output(model.NewAsset(normalize(record.Name, record.Name), normalize(resource.Value, resource.Value)), arn)
				}
			}
		} else {
			slog.Error("failed to list resource record sets", "error", err)
		}
	}

	a.exec(parseZones,
		"route53", "list-hosted-zones",
		"--query", "HostedZones[].Id",
		"--output", "json",
	)
}

func (a *Amazon) accounts() {
	a.exec(a.Send,
		"organizations", "list-accounts",
		"--query", "Accounts[*].[Id]",
		"--output", "text",
	)
}

func (a *Amazon) session(cmd *exec.Cmd) *exec.Cmd {
	cmd.Env = append(cmd.Env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", a.Id))
	cmd.Env = append(cmd.Env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", a.Key))
	cmd.Env = append(cmd.Env, fmt.Sprintf("AWS_SESSION_TOKEN=%s", a.Token))
	return cmd
}

func (a *Amazon) exec(parser func(string), args ...string) {
	cmd := a.session(exec.Command("aws", args...))
	if err := a.XYZ.Execute(cmd, parser); err != nil {
		slog.Error("failed to execute aws command", "error", err, "args", cmd.Args)
	}
}
