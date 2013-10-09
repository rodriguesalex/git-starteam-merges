package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// Flags
var (
	inFile = flag.String("in", "", "Input file")
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage: %s [options]
%s matches up StarTeam merge labels to source and target view.

The input file of StarTeam labels can be generated with the
org.sync.LabelDumper program from
https://github.com/patrick-higgins/git-starteam/tree/label-dumper

Writes a CSV file to stdout with source view, source label, target
view, and target label columns.
`, os.Args[0], os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

type Label struct {
	Name string
	*View
}

func (l Label) String() string {
	return fmt.Sprintf("%q.%q", l.View.Name, l.Name)
}

type Merge struct {
	Source Label
	Target Label
}

type View struct {
	Name      string
	Base      *View  // nil if no parent view
	BaseLabel *Label // nil if not label-based
	Labels    []Label
	OutMerges []Merge // source is this view
	InMerges  []Merge // target is this view
}

func (v *View) AddLabel(name string) {
	v.Labels = append(v.Labels, Label{Name: name, View: v})
}

func (v *View) FindLabel(name string) (Label, bool) {
	for _, l := range v.Labels {
		if l.Name == name {
			return l, true
		}
	}
	return Label{}, false
}

func (v *View) AddInMerge(merge Merge) {
	v.InMerges = append(v.InMerges, merge)
}

func (v *View) AddOutMerge(merge Merge) {
	v.OutMerges = append(v.OutMerges, merge)
}

func (v View) String() string {
	baseName := "nil"
	baseLabel := "nil"
	if v.Base != nil {
		baseName = v.Base.Name
	}
	if v.BaseLabel != nil {
		baseLabel = v.BaseLabel.Name
	}
	s := fmt.Sprintf("View: %s\n\tBase: %s\n\tBase label: %s\n\tLabels:\n", v.Name, baseName, baseLabel)
	for _, l := range v.Labels {
		s += "\t\t" + l.Name + "\n"
	}
	s += "\tIn merges:\n"
	for _, m := range v.InMerges {
		s += fmt.Sprintf("\t\t%s <- %s %s\n", m.Target.Name, m.Source.View.Name, m.Source.Name)
	}
	s += "\tOut merges:\n"
	for _, m := range v.OutMerges {
		s += fmt.Sprintf("\t\t%s -> %s %s\n", m.Source.Name, m.Target.View.Name, m.Target.Name)
	}
	return s
}

type Views struct {
	m map[string]*View
}

func (v Views) FindView(name string) (*View, bool) {
	view, ok := v.m[strings.ToLower(name)]
	return view, ok
}

func (v *Views) AddLabel(viewName, labelName string) {
	view, ok := v.FindView(viewName)
	if !ok {
		view = &View{Name: viewName}
		v.m[strings.ToLower(viewName)] = view
	}
	view.AddLabel(labelName)
}

func (v *Views) SetBase(viewName, baseViewName, baseLabelName string) error {
	view, ok := v.FindView(viewName)
	if !ok {
		return fmt.Errorf("Cannot set base for unknown view %q", viewName)
	}
	baseView, ok := v.FindView(baseViewName)
	if !ok {
		return fmt.Errorf("Base view %q of %q cannot be found", baseViewName, viewName)
	}
	view.Base = baseView
	label, ok := baseView.FindLabel(baseLabelName)
	if !ok {
		return fmt.Errorf("Base label %q on %q for %q not found", baseLabelName, baseViewName, viewName)
	}
	view.BaseLabel = &label
	return nil
}

func (v *Views) FindMerges() error {
	for _, view := range v.m {
		for _, label := range view.Labels {
			var err error
			switch {
			case strings.Contains(label.Name, "Merge-to-"):
				continue
			case strings.HasPrefix(label.Name, "MergeTo_"):
				continue
			case strings.HasPrefix(label.Name, "MergeFrom_"):
				err = v.V3Label(view, label)
			case strings.HasPrefix(label.Name, "Merge-from-"):
				err = v.V2Label(view, label)
			default:
				err = v.V1Label(view, label)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
	return nil
}

func (v Views) String() string {
	s := ""
	for _, view := range v.m {
		s += view.String()
	}
	return s
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

	in, err := os.Open(*inFile)
	if err != nil {
		log.Fatalf("Could not open %s: %v", *inFile, err)
	}

	views := Views{make(map[string]*View)}

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "::")
		switch fields[0] {
		case "L":
			views.AddLabel(fields[1], fields[2])
		case "LB":
			err = views.SetBase(fields[1], fields[2], fields[3])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
	if err = scanner.Err(); err != nil {
		log.Fatalf("Error reading %s: %v", *inFile, err)
	}

	err = views.FindMerges()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Find merges:", err)
	}

	//fmt.Print(views)

	csvOut := csv.NewWriter(os.Stdout)
	csvOut.Write([]string{
		"source view",
		"source label",
		"target view",
		"target label",
	})
	for _, view := range views.m {
		for _, inMerge := range view.InMerges {
			csvOut.Write([]string{
				inMerge.Source.View.Name,
				inMerge.Source.Name,
				inMerge.Target.View.Name,
				inMerge.Target.Name,
			})
		}
	}
	csvOut.Flush()
	err = csvOut.Error()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
