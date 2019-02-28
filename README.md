# ZYLISP

*A Go Lisp*

[![][logo]][logo-large]

[![Build Status][travis-badge]][travis]
[![Tag][tag-badge]][tag]
[![Go version][go-v]](.travis.yml)


## Intro

First, some quick examples...

Note that parenthesis always mean a function call or native Lisp form, and
function calls always use outer-parentheses.

Traditional lisp style:

```lisp
// tail recursion; tail-call optimization works, so this won't overflow the stack.
zylisp> (defn factTc [n accum]
        (cond (== n 0) accum
          (let [newn (- n 1)
                newaccum (* accum n)]
            (factTc newn newaccum))))
zylisp> (factTc 11 1) // compute factorial of 11, aka 11! aka 11*10*9*8*7*6*5*4*3*2
(factTc 11 1)
39916800
zylisp>
```

An optional infix syntax is layered on top. The infix syntax is a subset of
Go. Anything inside curly braces is infix. Outer parenthesis are still always
used for function calls. The zylisp REPL is in infix mode by default to
facilitate math.

```lisp
// show off the infix parser
zylisp> x := 3; y := 5; if x + y == 8 { (println "we add up") } else { (println "wat?" ) }
we add up
zylisp>
```


## Motivation


### An Embeddable Scripting Language

*Quickly create a mini-language to drive your project*

ZYLISP is an embeddable scripting language. It is a modernized Lisp with an
object-oriented flavor, and provides an interpreter and REPL
(Read-Eval-Print-Loop; that is, it comes with a command line interactive
interface).


### DSLs

ZYLISP allows you to create a Domain Specific Language to drive
your program with minimal fuss and maximum convenience.

It is written in Go and plays nicely with Go programs
and Go structs, using reflection to instantiate trees of Go structs
from the scripted configuration. These data structures are native
Go, and Go methods will run on them at compiled-Go speed.

Because it speaks JSON and Msgpack fluently, ZYLISP is ideally suited for driving
complex configurations and providing projects with a domain specific
language customized to your problem domain.

The example snippets in the tests/*.zy provide many examples.
The full [documentation can be found in the Wiki](https://github.com/zylisp/zylisp/wiki).
ZYLISP blends traditional and new. While the s-expression syntax
defines a Lisp, ZYLISP borrows some syntax from Clojure,
and some (notably the for-loop style) directly from the Go/C tradition.


## Installation

```bash
$ go get github.com/zylisp/zylisp/cmd/zylisp
```


## Development

```bash
$ git clone git@github.com:zylisp/zylisp.git
$ cd zylisp
$ export GOPATH=$GOPATH:`pwd`
$ export PATH=$PATH:`pwd`/bin
$ make
```


## Features

New features are tracked in the [CHANGELOG.md](CHANGELOG.md). Language details
are available on the [wiki](https://github.com/zylisp/zylisp/wiki). Features in
development are tracked on Github ([features](https://github.com/zylisp/zylisp/issues?q=is%3Aopen+is%3Aissue+label%3Afeature), [epics](https://github.com/zylisp/zylisp/issues?q=is%3Aopen+is%3Aissue+label%3Aepic)).


## The Name

ZYLISP is a contraction of Zygomys Lisp (the project from which it was forked); Zygomys is in turn is a contraction of Zygogeomys, [a genus of pocket gophers. The Michoacan pocket gopher (Zygogeomys trichopus) finds its natural habitat in high-altitude forests.](https://en.wikipedia.org/wiki/Michoacan_pocket_gopher)


## License

Copyright (c) 2016-2018, The zygomys authors.

Copyright (c) 2019, The ZYLISP authors.

Two-clause BSD, see LICENSE file.

## Authors

* Glisp - Howard Mao
* Zygomys - Jason E. Aten, Ph.D.
* ZYLISP - Duncan McGreggor


<!-- Named page links below: /-->

[logo]: media/images/logo-1-250x.png
[logo-large]: media/images/logo-1.png
[travis]: https://travis-ci.org/zylisp/zylisp
[travis-badge]: https://travis-ci.org/zylisp/zylisp.png?branch=master
[tag-badge]: https://img.shields.io/github/tag/zylisp/zylisp.svg
[tag]: https://github.com/zylisp/zylisp/tags
[go-v]: https://img.shields.io/badge/Go-1.12-blue.svg
