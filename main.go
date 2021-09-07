package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/google/go-github/v38/github"

	"github.com/pkg/errors"
)

func do(repo string, slackUrl string, now time.Time) error {
	if err := doGit(repo, now); err != nil {
		if errors.Is(err, NoChangedError) {
			log.Printf("[%s] No Changed.\n", repoBaseName(repo))
			return nil
		}

		if errors.Is(err, UntrackedFileExistsError) {
			log.Printf("[%s] %v\n", repoBaseName(repo), err)
			if err := doSlack(slackUrl,
				"warning",
				fmt.Sprintf("[%s] Snapshot not executed at %s", repoBaseName(repo), timeFormat(now)),
				fmt.Sprintf("%v", err),
			); err != nil {
				return err
			}
			return nil
		}

		// unexpected
		log.Printf("[%s] %+v\n", repoBaseName(repo), err)
		if err := doSlack(slackUrl,
			"danger",
			fmt.Sprintf("[%s] Unexpcted error at %s", repoBaseName(repo), timeFormat(now)),
			fmt.Sprintf("%+v", err),
		); err != nil {
			return err
		}

		return err
	}

	log.Printf("[%s] Pushed successfully.\n", repoBaseName(repo))
	if err := doSlack(slackUrl,
		"good",
		fmt.Sprintf("[%s] Snapshot executed at %s", repoBaseName(repo), timeFormat(now)),
		"looks good :)",
	); err != nil {
		return err
	}

	return nil
}

func timeFormat(t time.Time) string {
	return fmt.Sprintf("%s (%d)", t.Format("2006/01/02 15:04:05"), t.Unix())
}

func repoBaseName(repoPath string) string {
	return filepath.Base(repoPath)
}

type Setting struct {
	Repos    []RepoSetting `yaml:"repos"`
	Interval int           `yaml:"interval"`
}

type RepoSetting struct {
	LocalPath string `yaml:"local_path"`
	GitHub    struct {
		Owner string `yaml:"owner"`
		Repo  string `yaml:"repo"`
	} `yaml:"github"`
}

func readSetting(settingPath string) (Setting, error) {
	f, err := os.Open(settingPath)
	if err != nil {
		return Setting{}, errors.WithStack(err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return Setting{}, errors.WithStack(err)
	}

	var configs Setting
	if err := yaml.Unmarshal(content, &configs); err != nil {
		return Setting{}, errors.WithStack(err)
	}

	return configs, nil
}

const settingFileName = ".snap_git.yml"

func main() {
	slackUrl := os.Getenv("SLACK_WEBHOOK_URL")
	if slackUrl == "" {
		log.Fatalf("SLACK_WEBHOOK_URL must be set")
	}

	var (
		githubClient *github.Client
	)
	{
		accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
		if accessToken == "" {
			log.Fatalf("GITHUB_ACCESS_TOKEN must be set")
		}

		githubClient = newGithubClient(accessToken)
	}

	var (
		setting Setting
	)
	{
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("%+v", err)
		}

		reposFilePath := path.Join(home, settingFileName)

		if _, err := os.Stat(reposFilePath); err != nil {
			log.Printf("%+v", err)
			log.Fatalf("%s must be exist. List up local repository directory paths in the file", reposFilePath)
		}

		setting, err = readSetting(reposFilePath)
		if err != nil {
			log.Fatalf("%+v", err)
		}

		if len(setting.Repos) == 0 {
			log.Fatalf("%s is empty", reposFilePath)
		}

		for _, repo := range setting.Repos {
			if ok, err := gitExists(repo.LocalPath); !ok {
				log.Fatalf("%s %+v", repo.LocalPath, err)
			}

			if repoBaseName(repo.LocalPath) != repo.GitHub.Repo {
				log.Fatalf("`%s` (local repo name) is different from `%s` (github repo name). Is it mistake?", repoBaseName(repo.LocalPath), repo.GitHub.Repo)
			}

			if ok, err := isGithubPrivate(githubClient, repo.GitHub.Owner, repo.GitHub.Repo); err != nil {
				log.Fatalf("%s %+v", repo.GitHub.Repo, err)
			} else if !ok {
				log.Fatalf("%s is public!", repo.GitHub.Repo)
			}
		}
	}

	for _, repo := range setting.Repos {
		log.Printf("watching: %s \n", repo.LocalPath)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.Tick(time.Duration(setting.Interval) * time.Second)
	log.Println("watching start")

	for {
		select {
		case <-sigChan:
			log.Println("snap_git: interrupt. terminating...")
			os.Exit(0)
		case <-ticker:
			for _, repo := range setting.Repos {
				now := time.Now()
				log.Printf("[%s] Snapshotting...", repoBaseName(repo.LocalPath))

				if err := do(repo.LocalPath, slackUrl, now); err != nil {
					log.Printf("[%s] unexpected error happened: %v", repoBaseName(repo.LocalPath), err)
				}
			}
		}
	}
}
