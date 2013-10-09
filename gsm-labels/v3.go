package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Example V3 labels:
//
// Target::MergeFrom_Source.082.00_Source2.053.00	<- Source::082.00_Source2.053.00
// Target::MergeFrom_Source.056.00					<- Source::056.00
// Target::MergeFrom_Source.84_Source2.58			<- Source::84_Source2.58
// Target::MergeFrom_Source_10032013-181835			<- Source::MergeTo_10032013-181835

func (v *View) FindV3MergeLabel(dateTime string) (Label, bool) {
	var label Label
	found := 0
	for _, l := range v.Labels {
		if strings.HasSuffix(l.Name, dateTime) {
			found++
			label = l
		}
	}
	if found > 1 {
		fmt.Fprintf(os.Stderr, "ERROR: more than one date-time merge label found %q.%q\n", v.Name, dateTime)
		return Label{}, false
	}
	if found == 1 {
		return label, true
	}
	return label, false
}

var v3Pattern = regexp.MustCompile(`^MergeFrom_([A-Z][a-zA-Z0-9_]+)_(\d+-\d+)$`)

func (v Views) V3Label(view *View, label Label) error {
	subs := v3Pattern.FindStringSubmatch(label.Name)
	if subs == nil {
		return fmt.Errorf("Invalid V3 merge label %s", label)
	}

	sourceView, ok := v.FindView(subs[1])
	if !ok {
		return fmt.Errorf("V3 merge source view %q not found: %s", subs[1], label)
	}

	sourceLabel, ok := sourceView.FindV3MergeLabel(subs[2])
	if !ok {
		return fmt.Errorf("V3 merge source label %q %q not found: %s", sourceView.Name, subs[2], label)
	}

	merge := Merge{Source: sourceLabel, Target: label}
	sourceView.AddOutMerge(merge)
	view.AddInMerge(merge)
	return nil
}
