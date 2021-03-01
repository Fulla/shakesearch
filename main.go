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
)

const (
	PHRMODE = "phr"
	AWMODE  = "aw"
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
	simplifiedDat := simplifyText(dat)
	s.CompleteWorks = string(dat)
	s.simplifiedCompleteWorks = string(simplifiedDat)
	s.SuffixArray = suffixarray.New(simplifiedDat)
	s.searchParagraphs = newParagraphs(s.simplifiedCompleteWorks)
	s.resultParagraphs = newParagraphs(s.CompleteWorks)
	return nil
}

func (s *Searcher) SearchPhrase(query []byte) map[int]bool {
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
	return pIds
}

func (s *Searcher) SearchAllWords(query []byte) map[int]bool {
	// returns the set of ids for the paragraphs containing
	// all words in the query (even if they are not contiguous)
	words := bytes.Split(query, []byte(" "))
	partChans := make(chan map[int]bool, len(words))
	var partials []map[int]bool
	for _, word := range words {
		go func(w []byte) {
			idxs := s.SearchPhrase(w)
			partChans <- idxs
		}(word)
	}
	for part := range partChans {
		partials = append(partials, part)
		if len(partials) == len(words) {
			close(partChans)
		}
	}
	return intersection(partials)
}

func (s *Searcher) ResultText(paragraphId int) string {
	p := s.resultParagraphs.Get(paragraphId)
	pText := s.CompleteWorks[p.from:p.to]
	return pText
}

func (s *Searcher) Search(query string, mode string) []string {
	simplifiedQuery := simplifyText([]byte(query))
	results := []string{}
	var paragraphs map[int]bool
	if mode == AWMODE {
		paragraphs = s.SearchAllWords(simplifiedQuery)
	} else {
		paragraphs = s.SearchPhrase(simplifiedQuery)
	}
	for id := range paragraphs {
		text := s.ResultText(id)
		results = append(results, text)
	}
	return results
}
