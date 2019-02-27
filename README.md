# ZYLISP

[![][logo]][logo-large]

*a Go Lisp*

[![Build Status][travis-badge]][travis]
[![Tag][tag-badge]][tag]
[![Go version][go-v]](.travis.yml)


Quick examples...

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


### Quickly create a mini-language to drive your project

ZYLISP is an embeddable scripting language. It is a modernized Lisp with an
object-oriented flavor, and provides an interpreter and REPL
(Read-Eval-Print-Loop; that is, it comes with a command line interactive
interface).


## Why Use ZYLISP?

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


### Installation

```bash
$ go get github.com/zylisp/zylisp/cmd/zylisp
```


### Development

```bash
$ git clone git@github.com:zylisp/zylisp.git
$ cd zylisp
$ export GOPATH=$GOPATH:`pwd`
$ export PATH=$PATH:`pwd`/bin
$ make
```


### Features in ZYLISP 5.1.1

Not your average parentheses...

 * [x] package mechanism that supports modularity and isolation of scripts/packages/libraries from each other. [See tests/package.zy for examples.](https://github.com/glycerine/zylisp/blob/master/tests/package.zy)
 * [x] NaN handing that matches typical expectations/Go's answers.
 * [x] struct defintion and type checking. [See `tests/declare.zy` for examples.](https://github.com/glycerine/zylisp/blob/master/tests/declare.zy)
 * [x] Readable nested method calls: `(a.b.c.Fly)` calls method `Fly` on object `c` that lives within objects `a` and `b`.
 * [x] Use `zylisp` to configure trees of Go structs, and then run methods on them at natively-compiled speed (since you are calling into Go code).
 * [x] sandbox-able environment; try `zylisp -sandbox` and see the NewGlispSandbox() function.
 * [x] `emacs/zylisp.el` emacs mode provides one-keypress stepping through code.
 * [x] Command-line editing, with tab-complete for keywords (courtesy of https://github.com/peterh/liner)
 * [x] JSON and Msgpack interop: serialization and deserialization
 * [x] `(range key value hash_or_array (body))` range loops act like Go for-range loops: iterate through hashes or arrays.
 * [x] `(for [(initializer) (test) (advance)] (body))` for-loops match those in C and Go. Both `(break)` and `(continue)` are available for additional loop control, and can be labeled to break out of nested loops.
 * [x] Raw bytes type `(raw string)` lets you do zero-copy `[]byte` manipulation.
 * [x] Record definitions `(defmap)` make configuration a breeze.
 * [x] Files can be recursively sourced with `(source "path-string")`.
 * [x] Go style raw string literals, using `` `backticks` ``, can contain newlines and `"` double quotes directly. Easy templating.
 * [x] Easy to extend. See the `repl/random.go`, `repl/regexp.go`, and `repl/time.go` files for examples.
 * [x] Clojure-like threading `(-> hash field1: field2:)` and `(:field hash)` selection.
 * [x] Lisp-style macros for your DSL.
 * [x] optional infix notation within `{}` curly braces. Expressions typed at the REPL are assumed to be infix (wrapped in {} implicitly), enhancing the REPL experience for dealing with math.


### Additional Features

 * [x] Go-style comments, both `/*block*/` and `//through end-of-line.`
 * [x] ZYLISP is a small Go library, easy to integrate and use/extend.
 * [x] Float (float64), Int (int64), Char, String, Symbol, List, Array, and Hash datatypes builtin.
 * [x] Arithmetic (`+`, `-`, `*`, `/`, `mod`, `**`)
 * [x] Shift Operators (`sll`, `srl`, `sra`)
 * [x] Bitwise operations (`bitAnd`, `bitOr`, `bitXor`)
 * [x] Comparison operations (`<`, `>`, `<=`, `>=`, `==`, `!=`)
 * [x] Short-circuit boolean operators (`and` and `or`)
 * [x] Conditionals (`cond`)
 * [x] Lambdas (`fn`)
 * [x] Bindings (`def`, `defn`, `let`, `letseq`)
 * [x] Standalone and embedable REPL.
 * [x] Tail-call optimization
 * [x] Go API
 * [x] Macro System with macexpand `(macexpand (yourMacro))` makes writing/debugging macros easier.
 * [x] Syntax quoting -- with caret `^()` instead of backtick.
 * [x] Backticks used for raw multiline strings, as in Go.
 * [x] Lisp-expression quoting uses `%` (not `'`; which delimits runes as in Go).
 * [x] Channel and goroutine support
 * [x] Full closures with lexical scope.

[See the wiki for lots of details and a full description of the ZYLISP language.](https://github.com/zylisp/zylisp/wiki).


### The Name

ZYLISP is a contraction of Zygomys Lisp (the project from which it was forked); Zygomys is in turn is a contraction of Zygogeomys, [a genus of pocket gophers. The Michoacan pocket gopher (Zygogeomys trichopus) finds its natural habitat in high-altitude forests.](https://en.wikipedia.org/wiki/Michoacan_pocket_gopher)


### License

Copyright (c) 2016-2018, The zygomys authors.

Copyright (c) 2019, The ZYLISP authors.

Two-clause BSD, see LICENSE file.

### Authors

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
