package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func Exec(exe string, args ...string) (stdout string, stderr string, err error) {
	cmd := exec.Command(exe, args...)
	stdoutf, _ := cmd.StdoutPipe()
	stderrf, _ := cmd.StderrPipe()
	err = cmd.Start()
	if err != nil {
		return
	}
	out_bytes, _ := ioutil.ReadAll(stdoutf)
	stdout = string(out_bytes)
	err_bytes, _ := ioutil.ReadAll(stderrf)
	stderr = string(err_bytes)
	err = cmd.Wait()
	return
}

func dedupStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// good old https://gobyexample.com/collection-functions
func filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func removeStrings(slice, strings []string) []string {
	return filter(slice,func(s string) bool {
		return index(strings,s) == -1
	})
}

func contains(haystack []string, needle string) bool {
	return index(haystack, needle) >= 0
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}