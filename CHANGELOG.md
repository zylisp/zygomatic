## Changes in ZYLISP 6.0.0

* Project rename
* Code split into separate packages
* Function registration cleanup


## Features in ZYLISP 5.1.1

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


## Additional Features

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

