package main

import (
	"fmt"
	"github.com/shibukawa/configdir"
	"strings"
)

const packageFile = "go.list"

func Packages(rebuild bool) (map[string]string, error) {
	configDirs := configdir.New("grun", "grun")
	cache := configDirs.QueryCacheFolder()
	if !cache.Exists(packageFile) {
		rebuild = true
	}
	packages := make(map[string]string)
	if rebuild {
		stdout, stderr, e := Exec("go", "list", "-e", "all")
		if e != nil {
			return nil, fmt.Errorf("go list failed %w: %s", e, stderr)
		}
		lines := strings.Split(stdout, "\n")
		for _, line := range lines {
			parts := strings.Split(line, "/")
			if parts[0] == "golang.org" || parts[0] == "cmd" {
				continue
			}
			if len(parts) > 1 && !contains(parts, "internal") {
				packages[parts[len(parts)-1]] = line
			}
		}
		fmt.Println("rebuilt package list", len(packages))
		var sb strings.Builder
		for k, v := range packages {
			_, e := fmt.Fprintln(&sb, k, v)
			if e != nil {
				return nil, e
			}
		}
		e = cache.WriteFile(packageFile, []byte(sb.String()))
		if e != nil {
			return nil, e
		}
	} else {
		res, e := cache.ReadFile(packageFile)
		if e != nil {
			return nil, e
		}
		for _, line := range strings.Split(string(res), "\n") {
			parts := strings.Split(line, " ")
			if len(parts) > 1 {
				packages[parts[0]] = parts[1]
			}
		}
	}
	return packages, nil
}
