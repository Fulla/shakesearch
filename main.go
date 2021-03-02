package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	PHRMODE  = "phr"
	AWMODE   = "aw"
	MAXLINES = 10
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks           string
	simplifiedCompleteWorks string
	SuffixArray             *suffixarray.Index
	searchParagraphs        *Paragraphs
	resultParagraphs        *Paragraphs
	sections                *Sections
}

type TextResponse struct {
	Text []string `json:"text"`
	Work string   `json:"work"`
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		mode := r.URL.Query().Get("m")
		if mode == "" {
			mode = AWMODE
		}
		results := searcher.Search(query[0], mode)
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func simplifyText(input []byte) []byte {
	symbolsRe, _ := regexp.Compile(`[\',;.-:]`)
	noSymbols := symbolsRe.ReplaceAll(input, []byte(""))
	lowerCase := bytes.ToLower(noSymbols)
	return lowerCase
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %+v", err)
	}
	dat = bytes.TrimSpace(dat)
	dat = bytes.ReplaceAll(dat, []byte("\r\n"), []byte("\n"))
	dat = bytes.ReplaceAll(dat, []byte("\r"), []byte("\n"))
	dat = bytes.ReplaceAll(dat, []byte("’"), []byte("'"))
	simplifiedDat := simplifyText(dat)
	s.CompleteWorks = string(dat)
	s.simplifiedCompleteWorks = string(simplifiedDat)
	s.SuffixArray = suffixarray.New(simplifiedDat)
	s.searchParagraphs = newParagraphs(s.simplifiedCompleteWorks)
	s.resultParagraphs = newParagraphs(s.CompleteWorks)
	s.sections = newSections(s.CompleteWorks)
	return nil
}

func (s *Searcher) SearchPhrase(query []byte) (map[int]bool, [][]byte) {
	// returns the list of ids for the paragraphs containing
	// the exact phrase in query
	idxs := s.SuffixArray.Lookup(query, -1)
	pIds := make(map[int]bool, 0)
	for _, idx := range idxs {
		pId, err := s.searchParagraphs.ParagraphForTextIndex(idx)
		if err != nil {
			log.Panic(err)
			continue
		}
		pIds[pId] = true
	}
	return pIds, [][]byte{query}
}

func (s *Searcher) AsyncSearchWord(w []byte, result chan map[int]bool,
	done chan struct{}) {

	idxs, _ := s.SearchPhrase(w)
	select {
	case result <- idxs:
	case <-done:
		return
	}
}

func (s *Searcher) SearchAllWords(query []byte) (map[int]bool, [][]byte) {
	// returns the set of ids for the paragraphs containing
	// all words in the query (even if they are not contiguous)
	words := bytes.Split(query, []byte(" "))
	partChans := make(chan map[int]bool, len(words))
	done := make(chan struct{})
	var partials []map[int]bool
	for _, word := range words {
		go s.AsyncSearchWord(word, partChans, done)
	}
	for part := range partChans {
		if len(part) < 1 {
			// early return when no intersection is possible
			close(done)
			return map[int]bool{}, words
		}
		partials = append(partials, part)
		if len(partials) == len(words) {
			close(partChans)
		}
	}
	return intersection(partials), words
}

func searchIndexInLines(searchIdx int, lines []string) int {
	lineIdx := 0
	for i, line := range lines {
		lineIdx = lineIdx + len(line)
		if lineIdx > searchIdx {
			return i
		}
	}
	return len(lines)
}

func (s *Searcher) ResultText(paragraphId int, words [][]byte) TextResponse {
	p := s.resultParagraphs.Get(paragraphId)
	pText := s.CompleteWorks[p.from:p.to]
	w, _ := s.sections.FindWorkByTextIndex(p.from)
	title := ""
	if w != nil {
		title = w.title
	}
	textLines := strings.Split(pText, "\n")
	midWordIdx, err := searchInText(pText, words)
	if err != nil {
		log.Print(err)
		midWordIdx = []int{0, 5}
	}
	lineForWord := searchIndexInLines(midWordIdx[0], textLines)
	if len(textLines) > MAXLINES {
		textLines = balanceLines(textLines, lineForWord)
	}
	textLines = append([]string{"[...]"}, textLines...)
	textLines = append(textLines, "[...]")
	res := TextResponse{
		Text: textLines,
		Work: title,
	}
	return res
}

func (s *Searcher) Search(query string, mode string) []TextResponse {
	simplifiedQuery := simplifyText([]byte(query))
	var results []TextResponse
	var words [][]byte
	var paragraphs map[int]bool
	if mode == AWMODE {
		paragraphs, words = s.SearchAllWords(simplifiedQuery)
	} else {
		paragraphs, words = s.SearchPhrase(simplifiedQuery)
	}
	for id := range paragraphs {
		text := s.ResultText(id, words)
		results = append(results, text)
	}
	return results
}
