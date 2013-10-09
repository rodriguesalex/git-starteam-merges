package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Flags
var (
	inFile = flag.String("in", "", "Input CSV file")
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage: %s [options]
%s is a git-filter-branch parent-filter which reads StarTeam merge labels
from a CSV file and uses them to create the parents.
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

	commit := os.Getenv("GIT_COMMIT")
	//fmt.Fprintln(os.Stderr, "GIT_COMMIT:", commit)

	tags, err := getTags(commit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "getTags error:", err)
		os.Exit(1)
	}

	parents, err := commitParents(*inFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "commitParents error:", err)
		os.Exit(1)
	}

	var tagRefs []string
	for _, tag := range tags {
		if parent, ok := parents[tag]; ok {
			ref, err := revParse(parent + "^{commit}")
			if err != nil {
				fmt.Fprintf(os.Stderr, "couldn't parse rev %s: %v\n", tag, err)
				continue
			}
			tagRefs = append(tagRefs, "-p "+ref)
			//fmt.Fprintf(os.Stderr, "added parent tag %s for %s\n", parent, tag)
		} else {
			//fmt.Fprintln(os.Stderr, "no parent found for", tag)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		switch {
		case len(tagRefs) == 0:
			fmt.Println(scanner.Text())
		case scanner.Text() == "":
			fmt.Println(strings.Join(tagRefs, " "))
		default:
			fmt.Println(scanner.Text() + " " + strings.Join(tagRefs, " "))
		}
	}
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

func revParse(ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", ref)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func getTags(commit string) ([]string, error) {
	cmd := exec.Command("git", "tag", "--points-at", commit)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	if out.Len() == 0 {
		return nil, nil
	}
	return strings.Split(out.String(), "\n"), nil
}
