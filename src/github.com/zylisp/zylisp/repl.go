package zylisp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/zylisp/zcore"
)

var precounts map[string]int
var postcounts map[string]int

type Zlisp struct {
	*zcore.Zlisp
}

func CountPreHook(env *zcore.Zlisp, name string, args []zcore.Sexp) {
	precounts[name] += 1
}

func CountPostHook(env *zcore.Zlisp, name string, retval zcore.Sexp) {
	postcounts[name] += 1
}

func getLine(reader *bufio.Reader) (string, error) {
	line := make([]byte, 0)
	for {
		linepart, hasMore, err := reader.ReadLine()
		if err != nil {
			return "", err
		}
		line = append(line, linepart...)
		if !hasMore {
			break
		}
	}
	return string(line), nil
}

// NB at the moment this doesn't track comment and strings state,
// so it will fail if unbalanced '(' are found in either.
func isBalanced(str string) bool {
	parens := 0
	squares := 0

	for _, c := range str {
		switch c {
		case '(':
			parens++
		case ')':
			parens--
		case '[':
			squares++
		case ']':
			squares--
		}
	}

	return parens == 0 && squares == 0
}

var continuationPrompt = "... "

func (pr *Prompter) getExpressionOrig(reader *bufio.Reader) (readin string, err error) {
	line, err := getLine(reader)
	if err != nil {
		return "", err
	}

	for !isBalanced(line) {
		fmt.Printf(continuationPrompt)
		nextline, err := getLine(reader)
		if err != nil {
			return "", err
		}
		line += "\n" + nextline
	}
	return line, nil
}

// liner reads Stdin only. If noLiner, then we read from reader.
func (pr *Prompter) getExpressionWithLiner(env *zcore.Zlisp, reader *bufio.Reader, noLiner bool) (readin string, xs []zcore.Sexp, err error) {

	var line, nextline string

	if noLiner {
		fmt.Printf(pr.prompt)
		line, err = getLine(reader)
	} else {
		line, err = pr.Getline(nil)
	}
	if err != nil {
		return "", nil, err
	}

	err = zcore.UnexpectedEnd
	var x []zcore.Sexp

	// test parse, but don't load or generate bytecode
	env.Parser.ResetAddNewInput(bytes.NewBuffer([]byte(line + "\n")))
	x, err = env.Parser.ParseTokens()
	//P("\n after ResetAddNewInput, err = %v. x = '%s'\n", err, SexpArray(x).SexpString())

	if len(x) > 0 {
		xs = append(xs, x...)
	}

	for err == zcore.ErrMoreInputNeeded || err == zcore.UnexpectedEnd || err == zcore.ResetRequested {
		if noLiner {
			fmt.Printf(continuationPrompt)
			nextline, err = getLine(reader)
		} else {
			nextline, err = pr.Getline(&continuationPrompt)
		}
		if err != nil {
			return "", nil, err
		}
		env.Parser.NewInput(bytes.NewBuffer([]byte(nextline + "\n")))
		x, err = env.Parser.ParseTokens()
		if len(x) > 0 {
			for i := range x {
				if x[i] == zcore.SexpEnd {
					zcore.P("found an SexpEnd token, omitting it")
					continue
				}
				xs = append(xs, x[i])
			}
		}
		switch err {
		case nil:
			line += "\n" + nextline
			zcore.Q("no problem parsing line '%s' into '%s', proceeding...\n", line, (&zcore.SexpArray{Val: x, Env: env}).SexpString(nil))
			return line, xs, nil
		case zcore.ResetRequested:
			continue
		case zcore.ErrMoreInputNeeded:
			continue
		default:
			return "", nil, fmt.Errorf("Error on line %d: %v\n", env.Parser.Lexer.Linenum(), err)
		}
	}
	return line, xs, nil
}

func processDumpCommand(env *zcore.Zlisp, args []string) {
	if len(args) == 0 {
		env.DumpEnvironment()
	} else {
		err := env.DumpFunctionByName(args[0])
		if err != nil {
			fmt.Println(err)
		}
	}
}

func Repl(env *zcore.Zlisp, cfg *zcore.ZlispConfig) {
	var reader *bufio.Reader
	if cfg.NoLiner {
		// reader is used if one wishes to drop the liner library.
		// Useful for not full terminal env, like under test.
		reader = bufio.NewReader(os.Stdin)
	}

	if cfg.Trace {
		// debug tracing
		env.DebugExec = true
	}

	if !cfg.Quiet {
		fmt.Println(ReplBanner)
		if cfg.Sandboxed {
			fmt.Printf("ZYLISP version: %s, [sandbox mode]\n", Version())
		} else {
			fmt.Printf("ZYLISP version: %s\n", Version())
		}
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Println(ReplHelp)
	}
	var pr *Prompter // can be nil if noLiner
	if !cfg.NoLiner {
		pr = NewPrompter(cfg.Prompt)
		defer pr.Close()
	} else {
		pr = &Prompter{prompt: cfg.Prompt}
	}
	infixSym := env.MakeSymbol("infix")

	for {
		line, exprsInput, err := pr.getExpressionWithLiner(env, reader, cfg.NoLiner)
		//zcore.Q("\n exprsInput(len=%d) = '%v'\n line = '%s'\n", len(exprsInput), (&SexpArray{Val: exprsInput}).SexpString(nil), line)
		if err != nil {
			if err == io.EOF {
				fmt.Println(ReplExitMsg)
				os.Exit(0)
			} else {
				fmt.Println(err)
			}
			env.Clear()
			continue
		}

		parts := strings.Split(strings.Trim(line, " "), " ")
		//parts := strings.Split(line, " ")
		if len(parts) == 0 {
			continue
		}
		first := strings.Trim(parts[0], " ")

		if first == ".quit" {
			break
		}

		if first == ".cd" {
			if len(parts) < 2 {
				fmt.Printf("provide directory path to change to.\n")
				continue
			}
			err := os.Chdir(parts[1])
			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			pwd, err := os.Getwd()
			if err == nil {
				fmt.Printf("cur dir: %s\n", pwd)
			} else {
				fmt.Printf("error: %s\n", err)
			}
			continue
		}

		// allow & at the repl to take the address of an expression
		if len(first) > 0 && first[0] == '&' {
			//P("saw & at repl, first='%v', parts='%#v'. exprsInput = '%#v'", first, parts, exprsInput)
			exprsInput = []zcore.Sexp{zcore.MakeList(exprsInput)}
		}

		// allow * at the repl to dereference a pointer and print
		if len(first) > 0 && first[0] == '*' {
			//P("saw * at repl, first='%v', parts='%#v'. exprsInput = '%#v'", first, parts, exprsInput)
			exprsInput = []zcore.Sexp{zcore.MakeList(exprsInput)}
		}

		if first == ".dump" {
			processDumpCommand(env, parts[1:])
			continue
		}

		if first == ".gls" {
			fmt.Printf("\nScopes:\n")
			prev := env.ShowGlobalScope
			env.ShowGlobalScope = true
			err = env.ShowStackStackAndScopeStack()
			env.ShowGlobalScope = prev
			if err != nil {
				fmt.Printf("%s\n", err)
			}
			continue
		}

		if first == ".ls" {
			err := env.ShowStackStackAndScopeStack()
			if err != nil {
				fmt.Println(err)
			}
			continue
		}

		if first == ".verb" {
			zcore.Verbose = !zcore.Verbose
			fmt.Printf("verbose: %v.\n", zcore.Verbose)
			continue
		}

		if first == ".debug" {
			env.DebugExec = true
			fmt.Printf("instruction debugging on.\n")
			continue
		}

		if first == ".undebug" {
			env.DebugExec = false
			fmt.Printf("instruction debugging off.\n")
			continue
		}

		var expr zcore.Sexp
		n := len(exprsInput)
		if n > 0 {
			infixWrappedSexp := zcore.MakeList([]zcore.Sexp{infixSym, &zcore.SexpArray{Val: exprsInput, Env: env}})
			expr, err = env.EvalExpressions([]zcore.Sexp{infixWrappedSexp})
		} else {
			line = ReplLineInfixWrap(line)
			expr, err = env.EvalString(line + " ") // print standalone variables
		}
		switch err {
		case nil:
		case zcore.NoExpressionsFound:
			env.Clear()
			continue
		default:
			fmt.Print(env.GetStackTrace(err))
			env.Clear()
			continue
		}

		if expr != zcore.SexpNull {
			// try to print strings more elegantly!
			switch e := expr.(type) {
			case *zcore.SexpStr:
				if e.Backtick {
					fmt.Printf("`%s`\n", e.S)
				} else {
					fmt.Printf("%s\n", strconv.Quote(e.S))
				}
			default:
				switch sym := expr.(type) {
				case zcore.Selector:
					zcore.Q("repl calling RHS() on Selector")
					rhs, err := sym.RHS(env)
					if err != nil {
						zcore.Q("repl problem in call to RHS() on SexpSelector: '%v'", err)
						fmt.Print(env.GetStackTrace(err))
						env.Clear()
						continue
					} else {
						zcore.Q("got back rhs of type %T", rhs)
						fmt.Println(rhs.SexpString(nil))
						continue
					}
				case *zcore.SexpSymbol:
					if sym.IsDot() {
						resolved, err := zcore.DotGetSetHelper(env, sym.Name(), nil)
						if err != nil {
							fmt.Print(env.GetStackTrace(err))
							env.Clear()
							continue
						}
						fmt.Println(resolved.SexpString(nil))
						continue
					}
				}
				fmt.Println(expr.SexpString(nil))
			}
		}
	}
}

func runScript(env *zcore.Zlisp, fname string, cfg *zcore.ZlispConfig) {
	file, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	err = env.LoadFile(file)
	if err != nil {
		fmt.Println(err)
		if cfg.ExitOnFailure {
			os.Exit(-1)
		}
		return
	}

	_, err = env.Run()
	if cfg.CountFuncCalls {
		fmt.Println("Pre:")
		for name, count := range precounts {
			fmt.Printf("\t%s: %d\n", name, count)
		}
		fmt.Println("Post:")
		for name, count := range postcounts {
			fmt.Printf("\t%s: %d\n", name, count)
		}
	}
	if err != nil {
		fmt.Print(env.GetStackTrace(err))
		if cfg.ExitOnFailure {
			os.Exit(-1)
		}
		Repl(env, cfg)
	}
}

// ReplMain: like main() for a standalone repl, now in library
func ReplMain(cfg *zcore.ZlispConfig) {
	var env *zcore.Zlisp
	if cfg.LoadDemoStructs {
		zcore.RegisterDemoStructs()
	}
	if cfg.Sandboxed {
		env = zcore.NewZlispSandbox()
	} else {
		env = zcore.NewZlisp()
	}
	env.StandardSetup()
	if cfg.LoadDemoStructs {
		// avoid data conflicts by only loading these in demo mode.
		env.ImportDemoData()
	}

	if cfg.CpuProfile != "" {
		f, err := os.Create(cfg.CpuProfile)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		defer pprof.StopCPUProfile()
	}

	precounts = make(map[string]int)
	postcounts = make(map[string]int)

	if cfg.CountFuncCalls {
		env.AddPreHook(CountPreHook)
		env.AddPostHook(CountPostHook)
	}

	if cfg.Command != "" {
		_, err := env.EvalString(cfg.Command)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	runRepl := true
	args := cfg.Flags.Args()
	if len(args) > 0 {
		runRepl = false
		runScript(env, args[0], cfg)
		if cfg.AfterScriptDontExit {
			runRepl = true
		}
	}
	if runRepl {
		Repl(env, cfg)
	}

	if cfg.MemProfile != "" {
		f, err := os.Create(cfg.MemProfile)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		defer f.Close()

		err = pprof.Lookup("heap").WriteTo(f, 1)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}
}

func ReplLineInfixWrap(line string) string {
	return "{" + line + "}"
}
