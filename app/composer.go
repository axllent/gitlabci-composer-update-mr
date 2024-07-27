// Package app is the main application
package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// ComposerUpdate will update composer
func ComposerUpdate() (string, error) {
	args := []string{"update", "--no-progress"}

	for _, f := range Config.ComposerFlags {
		args = append(args, f)
	}

	return run(Config.ComposerPath, args...)
}

// ParseComposerLock parses a composer lock file
func ParseComposerLock() (ComposerLock, error) {
	var v = ComposerLock{}

	jsonFile, err := os.Open(Config.ComposerLockFile)
	if err != nil {
		return v, err
	}

	b, err := io.ReadAll(jsonFile)
	if err != nil {
		return v, err
	}

	if err := jsonFile.Close(); err != nil {
		return v, err
	}

	if err := json.Unmarshal(b, &v); err != nil {
		return v, err
	}

	v.Checksum, err = fileHash(Config.ComposerLockFile)
	if err != nil {
		return v, err
	}

	return v, nil
}

// CompareDiffs will return a ComposerDiff struct for parsing
func CompareDiffs(pre, post ComposerLock) ComposerDiff {
	var preLookup = make(map[string]Package)

	var diff = ComposerDiff{}
	diff.Checksum = post.Checksum

	for _, p := range pre.Packages {
		preLookup[p.Name] = p
	}

	for _, p := range pre.PackagesDev {
		preLookup[p.Name] = p
	}

	newPackages := append(post.Packages, post.PackagesDev...)

	for _, post := range newPackages {
		pre, ok := preLookup[post.Name]
		dp := ComposerDiffPackage{}
		dp.Name = post.Name
		dp.PostVersion = post.Version
		dp.URL = repoURL(post.Source.URL)

		if ok {
			// new version
			if pre.Version != post.Version {
				dp.PreVersion = pre.Version
				dp.CompareURL = compareURL(post.Source.URL, pre.Version, post.Version)
				diff.Packages = append(diff.Packages, dp)
			}
		} else {
			// new package
			diff.Packages = append(diff.Packages, dp)
		}

		delete(preLookup, pre.Name)
	}

	// deleted packages
	for _, del := range preLookup {
		dp := ComposerDiffPackage{}
		dp.Name = del.Name
		dp.PreVersion = del.Version
		dp.URL = repoURL(del.Source.URL)
		diff.Packages = append(diff.Packages, dp)
	}

	// we will add to this if there are packages
	diff.CommitMessage = Config.GitCommitTitle

	if len(diff.Packages) == 0 {
		return diff
	}

	// generate markdown description
	description := "## Updated Composer Packages\n\n"
	description += "Checksum: " + diff.Checksum
	description += "\n\n### Changes\n\n"
	for _, p := range diff.Packages {
		name := fmt.Sprintf("- [%s](%s): ", p.Name, p.URL)
		version := fmt.Sprintf("`%s...REMOVED`\n", p.PreVersion)
		if p.PreVersion != "" && p.PostVersion != "" {
			if p.CompareURL != "" {
				version = fmt.Sprintf("[`%s...%s`](%s)\n", p.PreVersion, p.PostVersion, p.CompareURL)
			} else {
				version = fmt.Sprintf("`%s...%s`\n", p.PreVersion, p.PostVersion)
			}
		} else if p.PostVersion != "" {
			version = fmt.Sprintf("`NEW...%s`\n", p.PostVersion)
		}
		description += name + version
	}
	diff.Description = description

	// append to the git commit message
	diff.CommitMessage += "\n"
	for _, p := range diff.Packages {
		version := fmt.Sprintf("%s...REMOVED", p.PreVersion)
		if p.PreVersion != "" && p.PostVersion != "" {
			version = fmt.Sprintf("%s...%s", p.PreVersion, p.PostVersion)
		} else if p.PostVersion != "" {
			version = fmt.Sprintf("NEW...%s", p.PostVersion)
		}
		diff.CommitMessage += "\n" + p.Name + ": " + version
	}

	return diff
}

func repoURL(uri string) string {
	uri = strings.TrimRight(uri, ".git")

	var re = regexp.MustCompile(`^git@`)
	uri = re.ReplaceAllString(uri, "https://")

	return uri
}

func compareURL(uri, pre, post string) string {
	uri = strings.TrimRight(uri, ".git")

	var re = regexp.MustCompile(`^git@`)
	uri = re.ReplaceAllString(uri, "https://")

	if strings.HasPrefix(uri, "https://github.com/") {
		uri = fmt.Sprintf("%s/compare/%s...%s", strings.TrimRight(uri, "/"), url.QueryEscape(pre), url.QueryEscape(post))
	} else if strings.HasPrefix(uri, "https://gitlab.") {
		uri = fmt.Sprintf("%s/compare/%s...%s", strings.TrimRight(uri, "/"), url.QueryEscape(pre), url.QueryEscape(post))
	} else if strings.HasPrefix(uri, "https://bitbucket.") {
		uri = fmt.Sprintf("%s/branches/compare/%s%%0D%s", strings.TrimRight(uri, "/"), url.QueryEscape(pre), url.QueryEscape(post))
	} else {
		return ""
	}

	return uri
}
