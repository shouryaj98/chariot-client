package capabilities

import (
	"context"
	"regexp"
	"time"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/google/go-github/v61/github"
)

type Git struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

func NewGit(job model.Job) model.Capability {
	return &Git{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (task *Git) Match() bool {
	return task.Asset.Is("github")
}

func (task *Git) Invoke() error {
	token := task.Job.Config["secret"]
	if pat, ok := task.Asset.Config["pat"]; ok {
		token = pat
	}
	client := github.NewClient(nil).WithAuthToken(token)

	re := regexp.MustCompile(`^(https://)?github\.com/([^/]+)/?$`)
	org := task.Asset.Name
	matches := re.FindStringSubmatch(org)
	org = matches[len(matches)-1]
	task.contributors(client, org)

	user, _, _ := client.Users.Get(context.Background(), org)
	isOrganization := user.GetType() == "Organization"

	opt := github.ListOptions{PerPage: 100}
	for {
		var repos []*github.Repository
		var resp *github.Response
		var err error

		if isOrganization {
			repos, resp, err = client.Repositories.ListByOrg(context.Background(), org, &github.RepositoryListByOrgOptions{ListOptions: opt})
		} else if _, ok := task.Asset.Config["pat"]; ok {
			repos, resp, err = client.Repositories.ListByAuthenticatedUser(context.Background(), &github.RepositoryListByAuthenticatedUserOptions{
				ListOptions: opt,
			})
		} else {
			repos, resp, err = client.Repositories.ListByUser(context.Background(), org, &github.RepositoryListByUserOptions{ListOptions: opt})
		}

		for _, repo := range repos {
			if !repo.GetFork() {
				asset := model.NewAsset(repo.GetHTMLURL(), *repo.Name)
				asset.Config = task.Asset.Config
				asset.Status = model.ActiveLow
				task.Job.Stream <- asset
				task.exposures(asset, client, repo)
			}
		}

		if err != nil || resp.NextPage == 0 {
			return err
		}
		opt.Page = resp.NextPage
	}
}

func (a *Git) Secret() string {
	return "/pats/github"
}

func (task *Git) contributors(client *github.Client, org string) {
	opt := github.ListOptions{PerPage: 100}
	for {
		users, resp, err := client.Organizations.ListMembers(context.Background(), org, &github.ListMembersOptions{ListOptions: opt})
		for _, user := range users {
			task.contributor(client, user.GetLogin())
		}
		if err != nil || resp.NextPage == 0 {
			break
		}
		users, resp, err = client.Organizations.ListMembers(context.Background(), org, &github.ListMembersOptions{ListOptions: opt})
		opt.Page = resp.NextPage
	}
}

func (task *Git) contributor(client *github.Client, login string) {
	opt := github.ListOptions{PerPage: 100}
	for {
		repos, resp, err := client.Repositories.ListByUser(context.Background(), login, &github.RepositoryListByUserOptions{ListOptions: opt})
		for _, repo := range repos {
			if !repo.GetFork() && repo.GetOwner().GetLogin() == login {
				task.Job.Stream <- model.NewAsset(repo.GetHTMLURL(), *repo.Name)
			}
		}
		if err != nil || resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
}

func (task *Git) exposures(asset model.Asset, client *github.Client, repo *github.Repository) {
	if repo.GetPrivate() {
		return
	}

	if time.Since(repo.GetCreatedAt().Time) <= 24*time.Hour {
		task.Job.Stream <- model.NewRisk(asset, "newly-created-public-repo")
	}

	if events, _, err := client.Activity.ListRepositoryEvents(context.Background(), repo.GetOwner().GetLogin(), repo.GetName(), &github.ListOptions{PerPage: 50}); err == nil {
		for _, event := range events {
			if event.GetType() != "PublicEvent" {
				continue
			}
			if time.Since(event.GetCreatedAt().Time) <= 24*time.Hour {
				risk := model.NewRisk(asset, "private-repo-newly-made-public")

				task.Job.Stream <- risk
				task.Job.Stream <- risk.Proof(event.GetRawPayload())
			}
		}
	}
}
