package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func ComposerUpdate() (string, error) {
	return run(Config.ComposerPath, "update", "--no-progress")
}

func ParseComposerLock() (ComposerLock, error) {
	var v = ComposerLock{}

	jsonFile, err := os.Open(Config.ComposerLockFile)
	if err != nil {
		return v, err
	}

	b, err := ioutil.ReadAll(jsonFile)
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
	var prelookup = make(map[string]Package)

	var diff = ComposerDiff{}
	diff.Checksum = post.Checksum
	// var postlookup = make(map[string]Package)

	for _, p := range pre.Packages {
		prelookup[p.Name] = p
	}

	for _, p := range pre.PackagesDev {
		prelookup[p.Name] = p
	}

	newpackages := append(post.Packages, post.PackagesDev...)

	// var p = []ComposerDiffPackage{}

	for _, post := range newpackages {
		pre, ok := prelookup[post.Name]
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

		delete(prelookup, pre.Name)
	}

	// deleted packages
	for _, del := range prelookup {
		dp := ComposerDiffPackage{}
		dp.Name = del.Name
		dp.PreVersion = del.Version
		dp.URL = repoURL(del.Source.URL)
		diff.Packages = append(diff.Packages, dp)
	}

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
