# snap_git
`sanp_git` automatically git checkout (for temporary branch) and commit and push and notify to slack.

Highly recommended using for **private** repository. (otherwise, a secret key is pushed unexpectedly for public repository).

Currently, `GITHUB_ACCESS_TOKEN` is required for checking if target repository is private.  