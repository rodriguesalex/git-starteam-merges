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

type TokenHandler interface {
	HandleToken(*lex.Lexer, io.Writer) error
}

type TagHarvester struct {
	TagToMark map[string]int
}

func (t *TagHarvester) HandleToken(l *lex.Lexer, o io.Writer) error {
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
		t.TagToMark[tag] = mark

		io.WriteString(o, l.Line())
		l.Consume()

	default:
		io.WriteString(o, l.Line())
		l.Consume()
	}
	return nil
}

type MergeAdder struct {
	MarkToParentMark map[int]int
	Commits          []*gitexport.Commit
}

func (m *MergeAdder) HandleToken(l *lex.Lexer, o io.Writer) error {
	switch l.Token() {
	case lex.CommitTok:
		parser := gitexport.NewLexerParser(l)
		commit, err := parser.Commit()
		if err != nil {
			return err
		}
		parentMark := int(0)
		ok := false
		if commit.Mark != 0 {
			if parentMark, ok = m.MarkToParentMark[commit.Mark]; ok {
				commit.Merge = append(commit.Merge, fmt.Sprintf(":%d", parentMark))
			}
		}
		m.Commits = append(m.Commits, commit)

	case lex.ResetTok:
		if err := m.WriteCommits(o); err != nil {
			return err
		}
		io.WriteString(o, l.Line())
		l.Consume()

	default:
		io.WriteString(o, l.Line())
		l.Consume()
	}
	return nil
}

func (m *MergeAdder) WriteCommits(o io.Writer) error {
	if len(m.Commits) == 0 {
		return nil
	}

	g := Graph{
		Vertices: make([]Vertex, len(m.Commits)+1),
	}

	for _, c := range m.Commits {
		if c.Mark == 0 {
			c.Write(o)
		} else {
			var mark int
			if c.From != "" {
				fmt.Sscanf(c.From, ":%d", &mark)
				g.AddEdge(c.Mark, mark)
				//fmt.Fprintf(temp, "%d %d\n", mark, c.Mark)
			}
			for _, merge := range c.Merge {
				fmt.Sscanf(merge, ":%d", &mark)
				g.AddEdge(c.Mark, mark)
				//fmt.Fprintf(temp, "%d %d\n", mark, c.Mark)
			}
		}
	}

	marks, err := TopoSort(g)
	if err != nil {
		return fmt.Errorf("Could not sort commits: %v", err)
	}

NextMark:
	for _, mark := range marks {
		// TODO: use indexed lookup
		for _, c := range m.Commits {
			if c.Mark == mark {
				c.Write(o)
				continue NextMark
			}
		}
	}

	m.Commits = nil

	return nil
}

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

func filter(l *lex.Lexer, o io.Writer, handler TokenHandler) {
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
			if err := handler.HandleToken(l, o); err != nil {
				fmt.Fprintf(os.Stderr, "%d: error: %s\n", l.LineNumber(), err)
				os.Exit(1)
			}
		}
	}
}

// reads the CSV file and returns a mapping from merge target to source tag.
func readCSV(file string) (map[string]string, error) {
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

func commitParents(parentTags map[string]string, tagMarks map[string]int) (map[int]int, error) {
	parentMarks := make(map[int]int)
	for k, v := range parentTags {
		km, kok := tagMarks[k]
		vm, vok := tagMarks[v]
		if kok && vok {
			parentMarks[km] = vm
		} else {
			fmt.Fprintf(os.Stderr, "Missing tag: %v=%v", k, v)
		}
	}
	return parentMarks, nil
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

	temp, err := ioutil.TempFile("", "gsm")
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not create temp file: ", err)
		os.Exit(1)
	}

	l := lex.New(os.Stdin)
	tagHarvester := &TagHarvester{TagToMark: make(map[string]int)}
	filter(l, temp, tagHarvester)

	_, err = temp.Seek(0, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not seek to beginning of temp file: ", err)
		os.Exit(1)
	}

	parentTags, err := readCSV(*inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", *inFile, err)
		os.Exit(1)
	}

	parentMarks, err := commitParents(parentTags, tagHarvester.TagToMark)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", *inFile, err)
		os.Exit(1)
	}

	//fmt.Fprintf(os.Stderr, "tagToMark: %#v\nparentTags: %#v\nparentMarks: %#v\n", tagHarvester.TagToMark, parentTags, parentMarks)

	mergeAdder := &MergeAdder{MarkToParentMark: parentMarks}

	l = lex.New(temp)
	filter(l, os.Stdout, mergeAdder)
}
