# grun

**grun** is a command-line tool for evaluating Go expressions. It was inspired by
[runner](https://github.com/stevedonovan/runner) which does the same for Rust.

It constructs a little Go program printing out the expression and runs it using `go run`.
This is normally fast but the first invocation may take a little while since
it runs `go list all` first to make a package map.

```sh
grun$ grun -e 'strings.Split("hello world"," ")'
[hello world]
```
You can choose to display the results in JSON, which is good for presenting data
structures like slices and maps neatly, although not so useful for
types that don't have public fields.

```sh
grun$ grun -j -e 'strings.Split("hello world"," ")'
[
  "hello",
  "world"
]
```

Go can be a bit verbose so there are some shortcuts

- `S.`: `strings.`
- `M.`: `math.`
- `C.`: `strconv.`
- `R(rx)`: `regexp.MustCompile(rx)`

```sh
grun$ grun -e 'R(`[a-z]+`).MatchString("hello")'
true
grun$ grun -e 'M.Sin(1.2)'
0.9320390859672264
```

Expressions may of course return multiple results, commonly for error returns:

```sh
grun$ grun -e 'os.Open("nada.txt")'
<nil> open nada.txt: no such file or directory
grun$ grun -j -e 'os.Open("nada.txt")'
null
 {
  "Op": "open",
  "Path": "nada.txt",
  "Err": 2
}
```

We resolve nested packages like `io/ioutil` using the package map:

```sh
grun$ cat tmp.txt
hello dolly
grun$ grun -e 'ioutil.ReadFile("tmp.txt")'
[104 101 108 108 111 32 100 111 108 108 121 10] <nil>
```

(These are not _always_ unique but it's good enough for now)

There are limits to what a single Go expresion can do, so the expression may be
preceded by one or more statements:

```sh
grun$ grun -e 's,_ := ioutil.ReadFile("tmp.txt"); string(s)'
hello dolly
```

Regular error handling is a _little clumsy_ for one-liners, so there is an implicit
shortcut. If there's a second var assigned, _and_ it starts with 'e', then we put in
a `log.Fatal` check:

```sh
grun$ grun -e 's,e := ioutil.ReadFile("nada.txt"); string(s)'
2021/04/21 16:03:07 open nada.txt: no such file or directory
exit status 1
2021/04/21 16:03:07 exit status 1
```
No self-respecting programming language designer would actually _implement_ such a feature,
but **grun** is about using Go in a very informal way.

It may feel more natural to actually edit these lines in a file:

```sh
grun$ cat read.go
// read.go
s,e := ioutil.ReadFile("nada.txt")
fmt.Println(string(s))

grun$ grun -f read.go
2021/04/21 16:17:24 open nada.txt: no such file or directory
exit status 1
2021/04/21 16:17:24 exit status 1
```
These are not really .go files, of course - the imports and boilerplate are implicit.

You _can_ include a proper Go file using `-i`:

```sh
grun$ cat answer.go
package main

func answer() int {
	return 42
}
grun$ grun -e 'answer()' -i answer.go
42
```

Finally, after importing new packages that you would like to make available to **grun**,
you will need to run `grun -r` to refresh the package list.
