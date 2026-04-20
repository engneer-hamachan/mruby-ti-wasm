//go:build !tinygo && !(js && wasm)

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"ti/base"
	"ti/builtin"
	"ti/cmd"
	"ti/context"
	"ti/eval"
	"ti/lexer"
	"ti/lexer/reader"
	"ti/loader"
	"ti/parser"
)

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

	if len(base.TSignatures) > 0 && flags.IsDefineAllInfo {
		cmd.PrintAllDefinitionsForLsp(p)
	}

	if len(base.TSignatures) > 0 && flags.IsSuggest {
		cmd.PrintSuggestionsForLsp(p)
	}

	if flags.IsHover {
		cmd.PrintHover(p)
	}

	if flags.IsExtends {
		cmd.PrintTargetClassExtends()
		os.Exit(0)
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

func preload(round string, flags *cmd.ExecuteFlags) {
	for _, preloadFile := range loader.GetPreloadFiles() {
		if fp, err := os.Open(preloadFile); err == nil {
			br := bufio.NewReader(fp)
			p := getParser(br, preloadFile)

			cmd.ApplyParserFlags(&p)

			evaluationLoop(p, flags, round, true)

			fp.Close()
		}
	}
}

func main() {
	if err := builtin.LoadFromFS(tiConfigFS); err != nil {
		panic(err)
	}

	var br *bufio.Reader
	var file string

	cmd.ValidateArgs()
	flags := cmd.BuildFlags()

	if flags.IsHelp {
		cmd.PrintHelp()
		return
	}

	if flags.IsVersion {
		cmd.PrintVersion()
		return
	}

	if flags.IsAllType {
		cmd.PrintAllTypes()
		return
	}

	var stdinContent []byte
	if flags.IsStdin {
		stdinContent, _ = io.ReadAll(os.Stdin)
	}

	for _, round := range context.GetRounds() {
		if flags.IsStdin {
			file = "stdin.rb"
			br = bufio.NewReader(bytes.NewReader(stdinContent))
		} else {
			file = cmd.GetTargetFile()
			fp, _ := os.Open(file)
			br = bufio.NewReader(fp)
		}

		p := getParser(br, file)
		cmd.ApplyParserFlags(&p)

		cleanSimpleIdentifires()

		preload(round, flags)
		evaluationLoop(p, flags, round, false)
	}
}
