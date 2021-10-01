package app

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
)

var (
	// Config struct
	Config struct {
		// ComposerPath binary path
		ComposerPath string

		// ComposerLockFile file path
		ComposerLockFile string

		// GitPath binary path
		GitPath string

		// Repo is the directory where the repository is
		RepoDir string

		// GitUser username
		GitUser string

		// GitEmail username
		GitEmail string

		// GitBranch branch name
		GitBranch string

		// MRBranch is the branch name for the merge request
		MRBranch string
	}
)

// BuildConfig will ensure the correct parameters are set
func BuildConfig() {
	errors := []error{}
	var err error

	composerVersion := fmt.Sprintf("composer-%d", envInt("COMPOSER_MR_COMPOSER_VERSION", 2))

	Config.ComposerPath, err = which(composerVersion)
	if err != nil {
		errors = append(errors, fmt.Errorf("\"%s\" not found", composerVersion))
	}

	Config.GitPath, err = which("git")
	if err != nil {
		errors = append(errors, fmt.Errorf("\"git\" not found"))
	}

	Config.RepoDir, err = filepath.Abs(Config.RepoDir)
	if err != nil {
		errors = append(errors, err)
	}

	Config.ComposerLockFile = path.Join(Config.RepoDir, "composer.lock")
	if !isFile(Config.ComposerLockFile) {
		errors = append(errors, fmt.Errorf("%s not found", Config.ComposerLockFile))
	}

	t := time.Now().Local().Format("20060102030405")

	Config.MRBranch = fmt.Sprintf("%scomposer-update-%s", envString("COMPOSER_MR_BRANCH_PREFIX", ""), t)

	if len(errors) == 0 {
		// test if project's merge requests are accessible
		if !isMREnabled() {
			errors = append(errors, fmt.Errorf("merge requests not enabled for %s, or API user doesn't have access to project", os.Getenv("CI_PROJECT_PATH")))
		}
	}

	if len(errors) > 0 {
		fmt.Println("\n==========\nError:")
		for _, err := range errors {
			fmt.Printf("- %v\n", err)
		}
		fmt.Println("==========")
		os.Exit(1)
	}
}
