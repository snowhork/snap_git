# snapshooting
`snapshooting` periodically git checkout (for temporary branch) and commit and push and notify to slack.

Highly recommended using for **private** repository. (otherwise, some secret keys will be pushed unexpectedly for public repository).

Currently, `GITHUB_ACCESS_TOKEN` is required for checking if target repository is private.

## requirements

* macOS
* `go (>= 1.16.0)` (for install) (If needed, run `brew install go`)
* `GITHUB_ACCESS_TOKEN` (get from your github accounts)
* `SLACK_WEBHOOK_URL` (get from your slack workspace settings, incoming webhook)

## installation

```sh
go install github.com/snowhork/snapshooting@latest
```

## settings
Create file `~/.snapshooting.yml` and edit it like this:

```yml
interval: 1800 # 30 minutes
repos:
  - local_path: "/path/to/local/secret-repository"
    github:
      owner: snowhork
      repo: secret-repository
      
  - local_path: "/path/to/local/another-repository"
    github:
      owner: some-organization
      repo: another-repository
GITHUB_ACCESS_TOKEN: xxxx
SLACK_WEBHOOK_URL: yyyy
```

## auto launch
Sorry, automatically launching is not supported.

But, it is practical to add to `.bash_profile` or `.zsh_profile` like this:

```sh
if [[ $(ps aux | grep snapshooting | grep -v grep | wc -l) -eq 0 ]]
   then
       snapshooting 
fi
```

## limitation
* only used for github private repository
* only notified to slack channel