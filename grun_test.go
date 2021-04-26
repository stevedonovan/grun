package main

import (
	"strings"
	"testing"
)

func compareResults(t *testing.T, desired, stdout, stderr string) {
	if stdout != desired {
		t.Errorf("%s: expected '%v', got '%v' (%v)",t.Name(),desired,stdout,stderr)
	}
}

func grunExpr(expr string, json bool) (string,string) {
	args := []string{}
	if json {
		args = append(args,"-J")
	}
	args = append(args,"-e",expr)
	stdout,stderr,_ := Exec("grun", args...)
	stdout = strings.TrimSpace(stdout)
	return stdout,stderr
}

func TestExprSimple(t *testing.T) {
	stdout,stderr := grunExpr(`strings.Split("hello dolly"," ")`, false)
	compareResults(t, "[hello dolly]", stdout, stderr)
}

func TestExprSimpleJson(t *testing.T) {
	stdout,stderr := grunExpr(`strings.Split("hello dolly"," ")`, true)
	compareResults(t, `["hello","dolly"]`, stdout, stderr)
}

func TestExprPackageShortcuts(t *testing.T) {
	stdout,stderr := grunExpr(`S.Split(S.TrimSpace(" hello dolly ")," ")`, false)
	compareResults(t, "[hello dolly]", stdout, stderr)
}

func TestExprRegexShortcuts(t *testing.T) {
	stdout,stderr := grunExpr(`R("^[a-z]\\d+").MatchString("z234 hey")`, false)
	compareResults(t, "true", stdout, stderr)
}

func TestExprWithStatements(t *testing.T) {
	stdout,stderr := grunExpr(`rx := R("^[a-z]\\d+"); rx.MatchString("k243 ")`, false)
	compareResults(t, "true", stdout, stderr)
}

func TestExprWithMultipleReturns(t *testing.T) {
	stdout,stderr := grunExpr(`m := R("^[a-z]\\d+").MatchString;  m("k243 "), m(" v50")`, false)
	compareResults(t, "true false", stdout, stderr)
}
