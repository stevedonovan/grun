package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type Data struct {
	Imports []string
	Lines   []string
	Expr    string
	Json    bool
	Flat    bool
	Args    bool
}

const templ = `package main
import (
{{range .Imports}}
	"{{- .}}"
{{end}}
)

{{- if .Json}}
func vararg(vv ...interface{}) []interface{} {
	return vv
}
func marshal(data interface{}) (string,error) {
	var b bytes.Buffer
    enc := json.NewEncoder(&b)
    enc.SetEscapeHTML(false)
	{{if not .Flat}}
	enc.SetIndent("","  ")
	{{end}}
    if e := enc.Encode(data); e != nil {
		return "", e
	}
    return b.String(),nil
}
{{- end}}

type M = map[string]interface{}
type S = []interface{}

func main() {
	{{- if .Args}}
	args := os.Args[1:]
	{{end}}
	{{range .Lines}}
	{{- .}}
	{{end}}
	{{- if .Expr}}
	{{- if .Json}}
	vals := vararg({{.Expr}})
	for _,val := range vals {
		b,e := marshal(&val)
		if e == nil {
			fmt.Print(string(b))
		} else {
			fmt.Println("cannot convert",val,"to JSON: ",e)
		}
	}
	{{else}}
	fmt.Println({{.Expr}})
	{{end}}
	{{end}}
}
`

var (
	expr      = flag.String("e", "", "expression to evaluate")
	verbose   = flag.Bool("v", false, "verbose mode")
	json_out  = flag.Bool("j", false, "Pretty JSON output")
	json_flat = flag.Bool("J", false, "Flat JSON output")
	file      = flag.String("f", "", "file to run")
	rebuild   = flag.Bool("r", false, "rebuild package list")
	infile    = flag.String("i", "", "Go file to link in")
)

var (
	tmpDir, _ = ioutil.TempDir("", "grun")
	tmpFile   = filepath.Join(tmpDir, "tmp.go")
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

var (
	substitutions = map[string]string{
		"S": "strings",
		"M": "math",
		"C": "strconv",
	}
)

func main() {
	flag.Parse()
	remainingArgs := flag.Args()
	if *rebuild {
		_, e := Packages(true)
		check(e)
		return
	}
	if *json_flat {
		*json_out = true
	}
	rx := regexp.MustCompile

	if *expr == "" && *file == "" {
		log.Fatal("provide expression with -e or file to run with -f")
	}
	if *file != "" {
		contents, e := ioutil.ReadFile(*file)
		if e != nil {
			log.Fatal("cannot read ", *file, " error ", e)
		}
		// line comments must begin the line
		slashSlash := rx(`^\s*//.*`)
		lines := filter(strings.Split(string(contents),"\n"),func(s string) bool {
			return ! slashSlash.MatchString(s)
		})
		*expr = strings.Join(lines,";")
	}
	packages, e := Packages(false)
	check(e)
	// pull out all string literals as vars so the substitutions don't zap them
	lines := []string{}
	kount := 0
	final := rx(`"[^"]*"`).ReplaceAllStringFunc(*expr, func(s string) string {
		kount++
		svar := "_s" + strconv.Itoa(kount)
		lines = append(lines, svar+":="+s)
		return svar
	})
	// R(<regex>) is a shortcut!
	final = rx(`\bR\(`).ReplaceAllString(final, "regexp.MustCompile(")
	// single-letter abbreviations for common packages
	final = rx(`\b[A-Z]\.`).ReplaceAllStringFunc(final, func(s string) string {
		s = strings.Trim(s, ".")
		return substitutions[s] + "."
	})
	uses_args := rx(`\bargs\b`).MatchString(final)
	expression := final
	auto_log := false
	receivers := []string{}
	// check if there are statements before the expression, and extract as lines...
	if strings.Contains(final, ";") {
		lines = append(lines, strings.Split(final, ";")...)
		last := len(lines) - 1
		expression = strings.TrimSpace(lines[last])
		extra := lines[0:last]
		// trim, and automatically insert error checks if assignment looks like an error
		assignVarError := rx(`([a-z]\w*)(,([a-z_]\w*))?\s*:=`)
		lines = []string{}
		for _, l := range extra {
			l = strings.TrimSpace(l)
			matches := assignVarError.FindStringSubmatch(l)
			lines = append(lines, l)
			if len(matches) > 0 {
				// important that any declared vars are not considered packages!
				receivers = append(receivers, matches[1]+".")
				// If an error is returned, handle it implicitly...
				err := matches[3]
				if strings.HasPrefix(err, "e") {
					lines = append(lines, fmt.Sprintf("if %s != nil { log.Fatal(%s) }", err, err))
					auto_log = true
				}
			}
		}
	}

	mods := rx(`\b[a-z]\w+\.`).FindAllString(final, -1)
	mods = append(mods, "fmt.")
	if uses_args {
		mods = append(mods, "os.")
	}
	if *json_out {
		mods = append(mods, "json.", "bytes.")
	}
	if auto_log {
		mods = append(mods, "log.")
	}
	mods = dedupStrings(mods)
	mods = removeStrings(mods, receivers)
	if *verbose {
		fmt.Println("modules", mods)
	}

	imports := []string{}
	for _, m := range mods {
		m = strings.Trim(m, ".")
		full, ok := packages[m]
		if ok {
			m = full
		}
		imports = append(imports, m)
	}
	data := Data{
		Imports: imports,
		Expr:    expression,
		Lines:   lines,
		Json:    *json_out,
		Flat:    *json_flat,
		Args:    uses_args,
	}
	tmpl, err := template.
		New("test").
		Parse(templ)
	check(err)
	f, e := os.Create(tmpFile)
	check(e)
	check(tmpl.Execute(f, data))
	check(f.Close())
	args := []string{"run", tmpFile}
	if *infile != "" {
		dest := filepath.Join(tmpDir, filepath.Base(*infile))
		e := copyFile(*infile, dest)
		check(e)
		args = append(args, dest)
		defer os.Remove(dest)
	}
	args = append(args, "--")
	args = append(args, remainingArgs...)
	stdout, stderr, e := Exec("go", args...)
	fmt.Print(stdout)
	fmt.Fprint(os.Stderr, stderr)
	check(e)
}
