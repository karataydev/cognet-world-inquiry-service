package model

type WordSearchResult struct {
	Word      string `json:"word"`
	Language  string `json:"language"`
	ConceptID string `json:"concept_id"`
}

type WordSuggestionResponse struct {
	Word         string       `json:"word"`
	ConceptID    string       `json:"concept_id"`
	LanguageInfo LanguageInfo `json:"language_info"`
}

type ChainWord struct {
	Word         string       `json:"word"`
	Translit1    string       `json:"translit1"`
	LanguageInfo LanguageInfo `json:"language_info"`
}

type CognateChain struct {
	Chain []ChainWord `json:"chain"`
}

type CognateChainResponse struct {
	ConceptID string         `json:"concept_id"`
	Chains    []CognateChain `json:"chains"`
}
