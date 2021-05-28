package app

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
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
func which(binName string) (string, error) {
	return exec.LookPath(binName)
}

// HashFileSHA1 will return the SHA1 hash of a file
func hashFileSHA1(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnSHA1String string

	//Open the filepath passed by the argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnSHA1String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new SHA1 hash interface to write to
	hash := sha1.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnSHA1String, err
	}

	//Get the 20 bytes hash
	hashInBytes := hash.Sum(nil)[:20]

	//Convert the bytes to a string
	returnSHA1String = hex.EncodeToString(hashInBytes)

	return returnSHA1String, nil
}

// // PrettyPrint for debugging
// func prettyPrint(i interface{}) {
// 	s, _ := json.MarshalIndent(i, "", "\t")
// 	fmt.Println(string(s))
// }
