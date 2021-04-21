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

func dedupStrings(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
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