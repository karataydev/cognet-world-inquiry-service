package model

type Cognate struct {
	ConceptID string `json:"concept_id"`
	Lang1     string `json:"lang1"`
	Word1     string `json:"word1"`
	Lang2     string `json:"lang2"`
	Word2     string `json:"word2"`
	Translit1 string `json:"translit1,omitempty"`
	Translit2 string `json:"translit2,omitempty"`
}
