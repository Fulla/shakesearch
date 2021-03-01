package main

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockText = `This is a text for testing package main.

I am skipping a line to start another paragraph.
This is still in the same paragraph.



Here we start again.`

func TestParagraphs(t *testing.T) {
	ps := newParagraphs(mockText)
	expected := []*Paragraph{
		&Paragraph{id: 0, from: 0, to: 40},
		&Paragraph{id: 1, from: 42, to: 127},
		&Paragraph{id: 2, from: 131, to: 151},
	}

	assert.Len(t, ps.paragraphs, len(expected))
	for i, p := range ps.paragraphs {
		log.Println(mockText[p.from:p.to])
		assert.EqualValues(t, p, expected[i])
	}
}
