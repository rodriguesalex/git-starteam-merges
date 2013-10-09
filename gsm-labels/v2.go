package main

import (
	"fmt"
	"os"
	"strings"
)

// Example V2 labels:
//
// Target::Merge-from-Source-091620131444 <- Source::131.00_Merge-to-Target-091620131444
// Target::Merge-from-Source-091620131556 <- Source::Merge-to-Target-091620131556

func (v *View) FindV2MergeLabel(mergeTo *View, date string) (Label, bool) {
	var label Label
	found := 0
	for _, l := range v.Labels {
		if strings.Contains(l.Name, mergeTo.Name+"-"+date) {
			found++
			label = l
		}
	}
	switch found {
	case 0:
		fmt.Fprintln(os.Stderr, "No V2 merge label found:", mergeTo, date)
		return Label{}, false
	case 1:
		return label, true
	}
	fmt.Fprintln(os.Stderr, "More than one V2 merge label found:", mergeTo, date)
	return Label{}, false
}

func (v Views) V2Label(view *View, label Label) error {
	fields := strings.Split(label.Name, "Merge-from-")
	if len(fields) != 2 {
		return fmt.Errorf("Invalid V2 merge label %s", label)
	}

	sourceFields := strings.SplitN(fields[1], "-", 2)
	if len(sourceFields) < 2 {
		return fmt.Errorf("Invalid V2 merge label %q %s", fields[1], label)
	}
	sourceView, ok := v.FindView(sourceFields[0])
	if !ok {
		return fmt.Errorf("V2 merge source view %q not found: %s", sourceFields[0], label)
	}
	sourceLabel, ok := sourceView.FindV2MergeLabel(view, sourceFields[1])
	if !ok {
		return fmt.Errorf("V2 merge source label %q.%q not found: %s", sourceView.Name, sourceFields[1], label)
	}
	merge := Merge{Source: sourceLabel, Target: label}
	sourceView.AddOutMerge(merge)
	view.AddInMerge(merge)
	return nil
}
