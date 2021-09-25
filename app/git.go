package app

import (
	"fmt"
	"os"
	"regexp"

	"github.com/xanzy/go-gitlab"
)

var (
	gitClient  *gitlab.Client
	gitIsSetup bool
)

// SwitchBranch will switch to a branch
func SwitchBranch(branch string) error {
	fmt.Println("Pulling latest changes from", branch)
	if out, err := runQuiet(Config.GitPath, "checkout", branch); err != nil {
		fmt.Println(out)
		return err
	}
	if out, err := runQuiet(Config.GitPath, "pull", "--rebase"); err != nil {
		fmt.Println(out)
		return err
	}

	return nil
}

// CreateMergeBranch creates the merge branch using git
func CreateMergeBranch() error {
	if err := gitSetup(); err != nil {
		return err
	}
	if out, err := runQuiet(Config.GitPath, "checkout", "-b", Config.MRBranch); err != nil {
		fmt.Println(out)
		return err
	}
	if out, err := runQuiet(Config.GitPath, "add", "."); err != nil {
		fmt.Println(out)
		return err
	}
	if out, err := runQuiet(Config.GitPath, "commit", "-m", "$ composer update"); err != nil {
		fmt.Println(out)
		return err
	}
	if out, err := runQuiet(Config.GitPath, "push", "origin", Config.MRBranch); err != nil {
		fmt.Println(out)
		return err
	}

	return nil
}

// DeleteOriginBranch will delete a branch from origin
func deleteOriginBranch(branch string) error {
	if err := gitSetup(); err != nil {
		return err
	}

	fmt.Printf("Deleting older branch/MR: %s\n", branch)
	if out, err := runQuiet(Config.GitPath, "push", "origin", ":"+branch); err != nil {
		fmt.Println(out)
		return err
	}

	return nil
}

// GitSetup will set up the local git instance wth the user.name, user.email
// and update remote url for committing
func gitSetup() error {
	if gitIsSetup {
		return nil
	}
	if _, err := run(Config.GitPath, "config", "user.name", Config.GitUser); err != nil {
		return err
	}

	if _, err := run(Config.GitPath, "config", "user.email", Config.GitEmail); err != nil {
		return err
	}

	if getAPIToken() != "" &&
		os.Getenv("CI_REPOSITORY_URL") != "" {
		var re = regexp.MustCompile(`^https:\/\/gitlab-ci-token:(.*)@(.*)`)
		var str = os.Getenv("CI_REPOSITORY_URL")

		match := re.FindStringSubmatch(str)
		originURL := fmt.Sprintf("https://gitlab-ci-token:%s@%s",
			getAPIToken(),
			match[2],
		)

		if _, err := run(Config.GitPath, "remote", "set-url", "origin", originURL); err != nil {
			fmt.Println("Error setting remote")
			return err
		}
	}

	gitIsSetup = true

	return nil
}
