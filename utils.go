package main

import (
	"fmt"
	"regexp"
)

type Paragraph struct {
	id   int
	from int
	to   int
}

func (p *Paragraph) Includes(index int) bool {
	return p.from <= index && index <= p.to
}

type Paragraphs struct {
	paragraphs []*Paragraph
}

func (ps *Paragraphs) Initialize(text string) {
	re, _ := regexp.Compile(`[\n\r]{2,}`)
	matches := re.FindAllStringIndex(text, -1)
	var paragraphs []*Paragraph
	pStart := 0
	for i, indexes := range matches {
		p := &Paragraph{
			id:   i,
			from: pStart,
			to:   indexes[0],
		}
		pStart = indexes[1]
		paragraphs = append(paragraphs, p)
	}
	lastP := &Paragraph{
		id:   len(matches),
		from: pStart,
		to:   len(text),
	}
	paragraphs = append(paragraphs, lastP)
	ps.paragraphs = paragraphs
}

func (ps *Paragraphs) ParagraphForTextIndex(index int) (int, error) {
	// returns the id of the paragraph that includes the text
	// index being passed as input
	for _, p := range ps.paragraphs {
		if p.Includes(index) {
			return p.id, nil
		}
	}
	return -1, fmt.Errorf("No paragraph found for text idx %d", index)
}

func (ps *Paragraphs) Get(id int) *Paragraph {
	if id >= len(ps.paragraphs) {
		return nil
	}
	return ps.paragraphs[id]
}

func newParagraphs(text string) *Paragraphs {
	ps := &Paragraphs{}
	ps.Initialize(text)
	return ps
}

func allContain(sets []map[int]bool, elem int) bool {
	for _, s := range sets {
		if !s[elem] {
			return false
		}
	}
	return true
}

func intersection(sets []map[int]bool) map[int]bool {
	if len(sets) < 1 {
		return map[int]bool{}
	}
	if len(sets) == 1 {
		return sets[0]
	}
	base := sets[0]
	others := sets[1:]
	intersec := make(map[int]bool)
	for elem := range base {
		if !allContain(others, elem) {
			continue
		}
		intersec[elem] = true
	}
	return intersec
}
