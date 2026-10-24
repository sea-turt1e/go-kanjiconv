package kanjiconv

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// SupportedGoVersions lists the Go releases this module is verified with.
// They are the first releases of each line that distrust the compromised
// eMudhra intermediate CA certificates
// (https://emudhra.we-have-a-problem.com/). Later patch releases of the same
// line are also accepted.
var SupportedGoVersions = []string{"go1.27.3", "go1.26.10"}

// GoVersionWarning returns a warning message when the running binary was
// built with a Go toolchain other than one of SupportedGoVersions (or a later
// patch release of the same line). It returns "" when the toolchain is
// supported.
func GoVersionWarning() string {
	return goVersionWarning(runtime.Version())
}

func goVersionWarning(v string) string {
	if isSupportedGoVersion(v) {
		return ""
	}
	return fmt.Sprintf(
		"warning: go-kanjiconv was built with %s, which is not a supported Go version (%s). "+
			"Rebuild with one of these to pick up the fix for the compromised eMudhra intermediate CAs: "+
			"https://emudhra.we-have-a-problem.com/",
		v, strings.Join(SupportedGoVersions, ", "))
}

func isSupportedGoVersion(v string) bool {
	line, patch, ok := splitGoVersion(v)
	if !ok {
		return false
	}
	for _, s := range SupportedGoVersions {
		sLine, sPatch, _ := splitGoVersion(s)
		if line == sLine && patch >= sPatch {
			return true
		}
	}
	return false
}

// splitGoVersion splits a release version such as "go1.27.3" into its
// release line ("go1.27") and patch number (3). Pre-releases ("go1.27rc1"),
// development builds ("devel ...") and versions with build suffixes
// ("go1.27.3 X:boringcrypto") are reported as not ok.
func splitGoVersion(v string) (line string, patch int, ok bool) {
	if !strings.HasPrefix(v, "go") || strings.Count(v, ".") != 2 {
		return "", 0, false
	}
	i := strings.LastIndex(v, ".")
	patch, err := strconv.Atoi(v[i+1:])
	if err != nil {
		return "", 0, false
	}
	return v[:i], patch, true
}
