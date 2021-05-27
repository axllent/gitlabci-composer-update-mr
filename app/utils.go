package app

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// EnvString will return an environment variable, else a default
func envString(key, dflt string) string {
	if os.Getenv(key) != "" {
		return strings.TrimSpace(os.Getenv(key))
	}
	return dflt
}

// EnvInt will return an environment int, else a default
func envInt(key string, dflt int) int {
	val := dflt
	if os.Getenv(key) != "" {
		i, err := strconv.Atoi(os.Getenv(key))
		if err == nil && i > 0 {
			val = i
		}
	}
	return val
}

// EnvCSVSlice will return an environment CSV slice, else a default
func envCSVSlice(key string, dflt []string) []string {
	if os.Getenv(key) != "" {
		return strings.Split(os.Getenv(key), ",")
	}
	return dflt
}

// EnvTrue will return an environment boolean, else a default
func envTrue(key string, dflt bool) bool {
	if os.Getenv(key) != "" {
		options := make(map[string]bool)
		options["true"] = true
		options["yes"] = true
		options["y"] = true
		options["1"] = true
		options["false"] = false
		options["no"] = false
		options["n"] = false
		options["0"] = false
		str := strings.ToLower(os.Getenv(key))
		val, ok := options[str]
		if ok {
			return val
		}
	}
	return dflt
}

// IsFile returns if a path is a file
func isFile(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) || !info.Mode().IsRegular() {
		return false
	}

	return true
}

// Which locates a binary in the current $PATH.
// It will append ".exe" to the filename if the platform is Windows.
func which(binName string) (string, error) {
	return exec.LookPath(binName)
}

// // PrettyPrint for debugging
// func prettyPrint(i interface{}) {
// 	s, _ := json.MarshalIndent(i, "", "\t")
// 	fmt.Println(string(s))
// }
