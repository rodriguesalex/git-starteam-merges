package main

import (
	"fmt"
	"regexp"
	"strings"
)

// Example V1 labels:
//
// namePrefix == 033.00-QA1 (strip -QA1)
// namePrefix == 11.00 (try 011.00)
// namePrefix == 019.00 && l.Name == 019_Mainline.059_ProdFix.026.00

func (v *View) FindV1MergeLabel(namePrefix string) (Label, bool) {
	for _, l := range v.Labels {
		name := l.Name
		if i := strings.Index(name, "_"); i != -1 {
			name = name[:i]
		}
		if strings.HasPrefix(name, namePrefix) {
			return l, true
		}
	}

	dashIndex := strings.Index(namePrefix, "-")
	dotIndex := strings.Index(namePrefix, ".")
	switch {
	case dashIndex != -1:
		return v.FindV1MergeLabel(namePrefix[:dashIndex])
	case dotIndex != -1:
		return v.FindV1MergeLabel(namePrefix[:dotIndex])
	case len(namePrefix) < 3:
		return v.FindV1MergeLabel("0" + namePrefix)
	}
	return Label{}, false
}

var v1_1Pattern = regexp.MustCompile(`^(\d{3}\.\d\d)_([A-Z][a-zA-Z0-9_]+)[._](\d{3}\.\d\d) ?$`)

func (v Views) V1Label(view *View, label Label) error {
	var sourceViewName, sourceLabelName, targetLabelName string

	subs := v1_1Pattern.FindStringSubmatch(label.Name)
	if subs != nil {
		targetLabelName = subs[1]
		sourceViewName = subs[2]
		sourceLabelName = subs[3]
	} else {
		fields := strings.Split(label.Name, "_")
		if len(fields) == 1 {
			return nil // not a merge
		}
		sourceFields := strings.SplitN(fields[1], ".", 2)
		if len(sourceFields) < 2 {
			return fmt.Errorf("Invalid V1 merge label %s", label)
		}
		targetLabelName = fields[0]
		sourceViewName = sourceFields[0]
		sourceLabelName = sourceFields[1]
	}

	sourceView, ok := v.FindView(sourceViewName)
	if !ok {
		return fmt.Errorf("V1 merge source view %q not found: %s",
			sourceViewName, label)
	}
	sourceLabel, ok := sourceView.FindV1MergeLabel(sourceLabelName)
	if !ok {
		return fmt.Errorf("V1 merge source label %q.%q not found: %s",
			sourceView.Name, sourceLabelName, label)
	}
	targetLabel, ok := view.FindV1MergeLabel(targetLabelName)
	if !ok {
		return fmt.Errorf("V1 merge target label %q.%q not found: %s",
			view.Name, targetLabelName, label)
	}
	merge := Merge{Source: sourceLabel, Target: targetLabel}
	sourceView.AddOutMerge(merge)
	view.AddInMerge(merge)
	return nil
}
