package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/patrick-higgins/gitexport"
	"github.com/patrick-higgins/gitexport/lex"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Flags
var (
	inFile = flag.String("in", "", "Input CSV file")
)

var targetToSourceMergeTag map[string]string
var tagToMark map[string]int = make(map[string]int)

type filterFunc func(*lex.Lexer, io.Writer) error

func passData(l *lex.Lexer, o io.Writer) {
	data, err := l.ConsumeData()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%d: error: %s\n", l.LineNumber(), l.Error())
		os.Exit(1)
	}
	fmt.Fprintf(o, "data %d\n", len(data))
	_, err = o.Write(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%d: error: %s\n", l.LineNumber(), l.Error())
		os.Exit(1)
	}
}

func filter(l *lex.Lexer, o io.Writer, fn filterFunc) {
	for {
		switch l.Token() {
		case lex.EOFTok:
			return
		case lex.ErrTok:
			fmt.Fprintf(os.Stderr, "%d: error: %s\n", l.LineNumber(), l.Error())
			os.Exit(1)
		case lex.InvalidTok:
			fmt.Fprintf(os.Stderr, "%d: invalid command: %s\n", l.LineNumber(), l.Line())
			os.Exit(1)

		case lex.DataTok:
			passData(l, o)

		default:
			err := fn(l, o)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%d: error: %s\n", l.LineNumber(), err)
				os.Exit(1)
			}
		}
	}
}

func passOne(l *lex.Lexer, o io.Writer) error {
	switch l.Token() {
	case lex.TagTok:
		tag := l.Field(1)
		io.WriteString(o, l.Line())
		l.Consume()
		var mark int
		_, err := fmt.Sscanf(l.Line(), "from :%d\n", &mark)
		if err != nil {
			return fmt.Errorf("invalid 'from' line: %s", l.Line())
		}
		tagToMark[tag] = mark

		io.WriteString(o, l.Line())
		l.Consume()

	default:
		io.WriteString(o, l.Line())
		l.Consume()
	}
	return nil
}

func passTwo(l *lex.Lexer, o io.Writer) error {
	switch l.Token() {
	case lex.CommitTok:
		parser := gitexport.NewLexerParser(l)
		commit, err := parser.Commit()
		if err != nil {
			return err
		}
		/*
			if commit.Mark != 0 {
				if parent, ok := parentMark[commit.Mark]; ok {
					commit.Merge = append(commit.Merge, parent)
				}
			}
		*/
		commit.Write(o)

	default:
		io.WriteString(o, l.Line())
		l.Consume()
	}
	return nil
}

func commitParents(file string) (map[string]string, error) {
	in, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	csvIn := csv.NewReader(in)
	// skip headers
	_, err = csvIn.Read()
	if err != nil {
		return nil, err
	}

	parents := make(map[string]string)
	for {
		record, err := csvIn.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		sourceTag := record[0] + "." + record[1]
		targetTag := record[2] + "." + record[3]
		parents[targetTag] = sourceTag
	}
	return parents, nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `usage: %s [options]
%s is a git-fast-export filter which reads StarTeam merge labels
from a CSV file and uses them to add the merge parents to commits.
`, os.Args[0], os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	var errs []string
	if *inFile == "" {
		errs = append(errs, "-in is required")
	}
	if len(errs) > 0 {
		fmt.Fprintln(os.Stderr, strings.Join(errs, "\n"))
		flag.Usage()
	}

	var err error
	targetToSourceMergeTag, err = commitParents(*inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", *inFile, err)
	}

	temp, err := ioutil.TempFile("", "gsm")
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not create temp file: ", err)
	}

	// First pass builds tagToMark entries
	l := lex.New(os.Stdin)
	filter(l, temp, passOne)

	_, err = temp.Seek(0, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not seek to beginning of temp file: ", err)
	}

	// Second pass writes merge entries using tagToMark entries
	l = lex.New(temp)
	filter(l, os.Stdout, passTwo)
}
