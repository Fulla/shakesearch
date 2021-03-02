package main

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"
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
	for elem, indexes := range base {
		if !allContain(others, elem) {
			continue
		}
		intersec[elem] = indexes
	}
	return intersec
}

type Work struct {
	title string
	from  int
	to    int
}

func (w *Work) Includes(index int) bool {
	return w.from <= index && index <= w.to
}

type Section struct {
	title string
	from  int
	to    int
}

type Sections struct {
	sections []Section
	works    []Work
}

func (s *Sections) FindWorkByTextIndex(index int) (*Work, error) {
	// returns the work to which it belongs the character at input index
	for _, w := range s.works {
		if w.Includes(index) {
			return &w, nil
		}
	}
	return nil, fmt.Errorf("No work found for text idx %d", index)
}

func getTitles(beginning string) []string {
	parts := strings.Split(beginning, "Contents")
	if len(parts) < 2 {
		return nil
	}
	indice := strings.TrimSpace(parts[1])
	matches := strings.Split(indice, "\n\n")
	var titles []string
	for _, res := range matches {
		title := strings.TrimSpace(res)
		titles = append(titles, title)
	}
	return titles
}

func getWorks(text string, titles []string, shiftIndex int) []Work {
	titlesOpts := strings.Join(titles, "|")
	rexp := fmt.Sprintf("(?P<title>(%s))[\n\r]{2,}", titlesOpts)
	workRe, _ := regexp.Compile(rexp)
	matches := workRe.FindAllStringSubmatchIndex(text, -1)
	var works []Work
	for i, indexes := range matches {
		title := text[indexes[2]:indexes[3]]
		to := len(text)
		if i < len(matches)-1 {
			to = matches[i+1][2]
		}
		p := Work{
			title: title,
			from:  indexes[2] + shiftIndex,
			to:    to + shiftIndex,
		}
		works = append(works, p)
	}
	return works
}

func (s *Sections) Initialize(text string) error {
	parts := strings.SplitN(text, "\n\n\n\n\n\n", 2)
	if len(parts) != 2 {
		return fmt.Errorf("failed to split indice and content. Found %d parts", len(parts))
	}
	indice := parts[0]
	content := parts[1]
	titles := getTitles(indice)
	works := getWorks(content, titles, len(indice)-1)
	s.works = works
	return nil
}

func newSections(text string) *Sections {
	s := &Sections{}
	err := s.Initialize(text)
	if err != nil {
		log.Panic(err)
	}
	return s
}

func maxInt(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func balanceLines(lines []string, targetLine int) []string {
	startLine := maxInt(targetLine-MAXLINES, 0)
	endLine := minInt(targetLine+MAXLINES, len(lines))
	half := int(math.Floor((float64(startLine) + float64(endLine)) / float64(2)))
	return lines[half-MAXLINES/2 : half+(MAXLINES/2)]
}

func searchInText(text string, wds [][]byte) ([]int, error) {
	var words []string
	text = strings.ToLower(text)
	for _, w := range wds {
		words = append(words, string(w))
	}
	rexp := fmt.Sprintf("(%s)", strings.Join(words, "|"))
	re, _ := regexp.Compile(rexp)
	matches := re.FindAllStringIndex(text, -1)
	if len(matches) < 1 {
		return []int{}, fmt.Errorf("Didnt find any of %+v in %s", words, text)
	}
	mid := int(math.Floor(float64(len(matches)) / float64(2)))
	return matches[mid], nil
}
