//go:build windows

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/devlights/gord/constants"

	"github.com/devlights/gord"
)

type (
	Args struct {
		dir     string
		text    string
		json    bool
		onlyHit bool
		verbose bool
		debug   bool
	}

	SimpleResult struct {
		Path string `json:"path"`
		Text string `json:"text"`
	}

	DetailResult struct {
		Path string `json:"path"`
		Page int32  `json:"page"`
		Line int32  `json:"line"`
		Text string `json:"text"`
	}
)

var (
	args Args
)

var (
	appLog = log.New(os.Stdout, "", 0)
)

func init() {
	flag.StringVar(&args.dir, "dir", ".", "directory")
	flag.StringVar(&args.text, "text", "", "search text")
	flag.BoolVar(&args.json, "json", false, "output as JSON")
	flag.BoolVar(&args.onlyHit, "only-hit", true, "show ONLY HIT")
	flag.BoolVar(&args.verbose, "verbose", false, "verbose mode")
	flag.BoolVar(&args.debug, "debug", false, "debug mode")
}

func newSimpleResult(path string, text string) *SimpleResult {
	return &SimpleResult{path, text}
}

func newDetailResult(path string, foundRange *gord.Range) *DetailResult {
	text, _ := foundRange.Text()
	page, _ := foundRange.PageNumber()
	line, _ := foundRange.LineNo()

	return &DetailResult{path, page, line, text}
}

func main() {
	flag.Parse()

	if args.text == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if args.dir == "" {
		args.dir = "."
	}

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func abs(p string) string {
	v, _ := filepath.Abs(p)
	return v
}

func genErr(procName string, err error) error {
	return fmt.Errorf("%s failed: %w", procName, err)
}

func run() error {
	quit, _ := gord.InitGord()
	defer quit()

	word, wordRelease, _ := gord.NewGord()
	defer wordRelease()

	_ = word.Silent(false)

	docs, err := word.Documents()
	if err != nil {
		return genErr("word.Documents()", err)
	}

	rootDir := abs(args.dir)
	err = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.Contains(filepath.Base(path), "~$") {
			return nil
		}

		if !strings.HasSuffix(path, ".docx") {
			return nil
		}

		absPath := abs(path)
		doc, docRelease, err := docs.Open(absPath)
		if err != nil {
			return err
		}
		defer docRelease()

		if args.debug {
			appLog.Printf("Document Open: %s", absPath)
		}

		allRange, err := doc.AllRange()
		if err != nil {
			return genErr("doc.AllRange()", err)
		}

		find, err := allRange.Find()
		if err != nil {
			return genErr("range.Find()", err)
		}

		if err := find.Forward(true); err != nil {
			return genErr("find.Forward()", err)
		}

		if err := find.Wrap(constants.WdFindWrapFindStop); err != nil {
			return genErr("find.Wrap()", err)
		}

		if err := find.Format(false); err != nil {
			return genErr("find.Format()", err)
		}

		if err := find.MatchCase(false); err != nil {
			return genErr("find.MatchCase()", err)
		}

		// ワイルドカードを有効にすると MatchCase の設定が無効となるため
		// サーチテキストに * が含まれている場合のみ有効となるようにする.
		//
		// > A wildcard search is always case-sensitive.
		// > You'll notice the same in the interface: if you tick the "Use wildcards" check box,
		// > the "Match case" and the "Find whole words only" check boxes will be disabled;
		//
		// REF: http://www.vbaexpress.com/forum/showthread.php?41816-Match-Case-in-Find-does-not-work
		// REF: https://answers.microsoft.com/en-us/msoffice/forum/all/matchwildcardstrue-renders-matchcase-inoperative/c20685ff-99c8-4334-9e59-c5fb95ff617c
		if strings.Contains(args.text, "*") {
			if err := find.MatchWildcards(true); err != nil {
				return genErr("find.MatchWildcards()", err)
			}
		}

		if err := find.Text(args.text); err != nil {
			return genErr("find.Text()", err)
		}

		found, err := find.Execute()
		if err != nil {
			return genErr("find.Execute() [first time]", err)
		}

		relPath, _ := filepath.Rel(rootDir, absPath)
		if !found {
			if !args.onlyHit {
				err = outSimple(newSimpleResult(relPath, "NO HIT"))
				if err != nil {
					return genErr("outSimple(result)", err)
				}
			}

			return nil
		}

		if !args.verbose {
			err = outSimple(newSimpleResult(relPath, "HIT"))
			if err != nil {
				return genErr("outSimple(result)", err)
			}

			return nil
		}

		foundRange := allRange
		for found {
			err = outDetail(newDetailResult(relPath, foundRange))
			if err != nil {
				return genErr("outDetail(result)", err)
			}

			found, err = find.Execute()
			if err != nil {
				return genErr("find.Execute() [after the second time]", err)
			}

			foundRange = allRange
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func outSimple(result *SimpleResult) error {
	var (
		message = fmt.Sprintf("%s: %s", result.Path, result.Text)
		err     error
	)

	if args.json {
		message, err = toJson(result)
		if err != nil {
			return genErr("toJson(result)", err)
		}
	}

	appLog.Println(message)

	return nil
}

func outDetail(result *DetailResult) error {
	var (
		message = fmt.Sprintf("%s (%3d,%3d): %q", result.Path, result.Page, result.Line, result.Text)
		err     error
	)

	if args.json {
		message, err = toJson(result)
		if err != nil {
			return genErr("toJson(result)", err)
		}
	}

	appLog.Println(message)

	return nil
}

func toJson(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", genErr("json.Marshal(result)", err)
	}

	return string(b), nil
}
