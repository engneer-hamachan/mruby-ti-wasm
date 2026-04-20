//go:build js && wasm

package main

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"os"
	"runtime/debug"
	"syscall/js"
	"ti/base"
	"ti/builtin"
	"ti/cmd"
	"ti/context"
	"ti/eval"
	"ti/lexer"
	"ti/lexer/reader"
	"ti/parser"
	"time"
)

//go:embed .ti-config
var tiConfigFS embed.FS

func getParser(br *bufio.Reader, file string) parser.Parser {
	lr := reader.New(*br)
	l := lexer.New(lr)

	return parser.New(l, file)
}

func setDefineInfos(p *parser.Parser) {
	for _, article := range eval.DefineInfoArticles {
		ctx := article.Ctx

		methodT := article.MethodT
		defineRow := article.DefineRow

		var prefix string

		prefix += "@"
		prefix += p.FileName + ":::"
		prefix += fmt.Sprintf("%d", defineRow)
		prefix += ":::"

		content :=
			base.MakeSignatureContent(prefix, ctx.GetFrame(), ctx.GetClass(), &methodT)

		content += " ["

		switch ctx.IsDefineStatic {
		case true:
			content += "c/"
		default:
			content += "i/"
		}

		readable := [2]bool{ctx.IsPrivate, ctx.IsProtected}

		switch readable {
		case [2]bool{true, false}:
			content += "private"
		case [2]bool{false, true}:
			content += "protected"
		case [2]bool{false, false}:
			content += "public"
		}

		content += "]"

		p.DefineInfos = append(p.DefineInfos, content)
	}
}

func appendSignature() {
	for _, article := range base.TSignatureArticles {
		frame := article.Frame
		class := article.Class
		methodT := article.MethodT
		isStatic := article.IsStatic
		isPrivate := article.IsPrivate
		fileName := article.FileName
		row := article.Row

		key := frame + class + methodT.GetMethodName()
		if isStatic {
			key += "static"
		}

		content := base.MakeSignatureContent(methodT.GetMethodName(), frame, class, &methodT)
		document := base.TSignatureDocument[key]

		sig :=
			base.Sig{
				Method:    methodT.GetMethodName(),
				Detail:    content,
				Frame:     frame,
				Class:     class,
				IsStatic:  isStatic,
				IsPrivate: isPrivate,
				FileName:  fileName,
				Row:       row,
				Document:  document,
			}

		_, ok := base.TSignatures[key]
		if ok && frame == "Builtin" {
			key = fmt.Sprintf("%s__overload_%s", key, content)
		}

		base.TSignatures[key] = sig
	}
}

func evaluationLoop(
	p parser.Parser,
	flags *cmd.ExecuteFlags,
	round string,
	isLoad bool,
) {

	ctx := context.NewContext("", "", round)
	evaluator := eval.Evaluator{}

	p.Errors = []error{}

	for {
		t, err := p.Read()
		if err != nil {
			p.Fatal(ctx, err)
		}

		if t != nil {
			dbg("token:", t.ToString(), "row:", p.Row)
		}

		err = evaluator.Eval(&p, ctx, t)
		if err != nil {
			p.Fatal(ctx, err)
		}

		if t != nil {
			continue
		}

		break
	}

	if round != "check" {
		return
	}

	if isLoad {
		return
	}

	setDefineInfos(&p)
	appendSignature()

	if len(p.DefineInfos) > 0 && flags.IsDefineInfo {
		cmd.PrintDefineInfosForPlugin(p.DefineInfos)
	}

	if len(base.TSignatures) > 0 && flags.IsSuggest {
		cmd.PrintSuggestionsForLsp(p)
	}

	if flags.IsHover {
		cmd.PrintHover(p)
	}

	if len(p.Errors) > 0 {
		cmd.PrintAllErrorsForPlugin(p)
		os.Exit(0)
	}
}

func cleanSimpleIdentifires() {
	for key, value := range base.TFrame {
		if value.IsIdentifierType() && base.ParseFrameKey(key).Variable() == value.ToString() {
			delete(base.TFrame, key)
		}
	}
}

func main() {
	debug.SetGCPercent(-1)

	if js.Global().Get("tiDebug").Truthy() {
		DebugEnabled = true
	}

	if err := builtin.LoadFromFS(tiConfigFS); err != nil {
		panic(err)
	}

	jsCode := js.Global().Get("tiCode")
	codeBytes := make([]byte, jsCode.Get("length").Int())
	js.CopyBytesToGo(codeBytes, jsCode)

	flags := cmd.NewExecuteFlags()

	if js.Global().Get("tiSuggest").Truthy() {
		flags.IsSuggest = true
	}

	if js.Global().Get("tiDefineInfo").Truthy() {
		flags.IsDefineInfo = true
	}

	suggestRow := 0
	if flags.IsSuggest {
		suggestRow = js.Global().Get("tiRow").Int()
	}

	timeout := time.After(1000 * time.Millisecond)
	done := make(chan bool, 1)

	go func() {
		for _, round := range context.GetRounds() {
			dbg("round:", round)
			br := bufio.NewReader(bytes.NewReader(codeBytes))
			p := getParser(br, "input.rb")

			if suggestRow > 0 {
				p.LspTargetRow = suggestRow
			}

			cleanSimpleIdentifires()
			evaluationLoop(p, flags, round, false)
			dbg("round done:", round)
		}
		done <- true
	}()

	select {
	case <-done:
	case <-timeout:
		fmt.Print("")
	}
	os.Exit(0)
}
