package service

import (
	"cognet-world-inquiry-service/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

type CognateSearch interface {
	GetWordSuggestions(ctx context.Context, prefix string) ([]model.WordSuggestionResponse, error)
	FindCognateChains(ctx context.Context, conceptID, word, lang string) (*model.CognateChainResponse, error)
	FindByConceptID(ctx context.Context, conceptID string) ([]model.Cognate, error)
}

type WordSuggestion struct {
	Word      string `json:"word"`
	Language  string `json:"language"`
	ConceptID string `json:"concept_id"` // Added ConceptID to suggestion
}

type cognateSearch struct {
	redisClient *redis.Client
}

func NewCognateSearch(redisClient *redis.Client) CognateSearch {
	return &cognateSearch{
		redisClient: redisClient,
	}
}

type coordinateTracker struct {
	used map[string]int // key: "lat,lng", value: count
}

func newCoordinateTracker() *coordinateTracker {
	return &coordinateTracker{
		used: make(map[string]int),
	}
}

func (ct *coordinateTracker) getAdjustedCoordinates(original []float64) []float64 {
	key := fmt.Sprintf("%f,%f", original[0], original[1])
	count := ct.used[key]
	ct.used[key]++

	// If this is the first word at these coordinates, return original
	if count == 0 {
		return original
	}

	// Define offset patterns (in degrees)
	// These patterns form a circular pattern around the original point
	offsets := [][2]float64{
		{1, 0},       // right
		{-1, 0},      // left
		{0, 1},       // up
		{0, -1},      // down
		{0.5, 0.5},   // up-right
		{-0.5, 0.5},  // up-left
		{0.5, -0.5},  // down-right
		{-0.5, -0.5}, // down-left
	}

	// Get the offset pattern based on count
	offsetIndex := (count - 1) % len(offsets)
	offset := offsets[offsetIndex]

	return []float64{
		original[0] + offset[0],
		original[1] + offset[1],
	}
}

func (cs *cognateSearch) getLanguageInfo(ctx context.Context, langCode string) (model.LanguageInfo, error) {
	data, err := cs.redisClient.Get(ctx, fmt.Sprintf("lang:%s", langCode)).Result()
	if err != nil {
		return model.LanguageInfo{}, fmt.Errorf("failed to get language info: %w", err)
	}

	var langInfo model.LanguageInfo
	if err := json.Unmarshal([]byte(data), &langInfo); err != nil {
		return model.LanguageInfo{}, fmt.Errorf("failed to unmarshal language info: %w", err)
	}

	return langInfo, nil
}

func (cs *cognateSearch) GetWordSuggestions(ctx context.Context, prefix string) ([]model.WordSuggestionResponse, error) {
	if len(prefix) < 2 {
		return []model.WordSuggestionResponse{}, nil
	}

	prefix = strings.ToLower(prefix)
	limit := 10

	// Get matches from prefix index
	matches, err := cs.redisClient.SMembers(ctx, fmt.Sprintf("prefix:%s", prefix)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch suggestions: %w", err)
	}

	// Process matches and remove duplicates
	seen := make(map[string]bool)
	suggestions := make([]model.WordSuggestionResponse, 0)

	for _, match := range matches {
		// match format: "word|language|conceptID"
		parts := strings.Split(match, "|")
		if len(parts) != 3 {
			continue
		}

		word := parts[0]
		if seen[word] {
			continue
		}
		seen[word] = true

		langInfo, _ := cs.getLanguageInfo(ctx, parts[1])

		suggestions = append(suggestions, model.WordSuggestionResponse{
			Word:         word,
			LanguageInfo: langInfo,
			ConceptID:    parts[2],
		})

		if len(suggestions) >= limit {
			break
		}
	}

	return suggestions, nil
}

func (cs *cognateSearch) buildChains(cognates []model.Cognate, ctx context.Context) ([]model.CognateChain, error) {
	// Create a map of connections and store original cognates
	connections := make(map[string]map[string]model.Cognate)
	cognateData := make(map[string]model.Cognate) // Store original cognate data

	for _, cog := range cognates {
		// Initialize maps if they don't exist
		if connections[cog.Lang1+":"+cog.Word1] == nil {
			connections[cog.Lang1+":"+cog.Word1] = make(map[string]model.Cognate)
		}
		if connections[cog.Lang2+":"+cog.Word2] == nil {
			connections[cog.Lang2+":"+cog.Word2] = make(map[string]model.Cognate)
		}

		// Store bidirectional connections
		connections[cog.Lang1+":"+cog.Word1][cog.Lang2+":"+cog.Word2] = cog
		connections[cog.Lang2+":"+cog.Word2][cog.Lang1+":"+cog.Word1] = cog

		// Store original cognate data for both words
		cognateData[cog.Lang1+":"+cog.Word1] = cog
		cognateData[cog.Lang2+":"+cog.Word2] = cog
	}

	processed := make(map[string]bool)
	var chains []model.CognateChain
	coordTracker := newCoordinateTracker()

	// Helper function to create chain word
	createChainWord := func(word, lang string, key string) (model.ChainWord, error) {
		langInfo, err := cs.getLanguageInfo(ctx, lang)
		if err != nil {
			return model.ChainWord{}, err
		}

		adjustedCoords := coordTracker.getAdjustedCoordinates(langInfo.Coordinates)
		adjustedLangInfo := langInfo
		adjustedLangInfo.Coordinates = adjustedCoords

		// Get the original cognate data
		cognate, exists := cognateData[key]
		var translit string
		if exists {
			if key == cognate.Lang1+":"+cognate.Word1 {
				translit = cognate.Translit1
			} else if key == cognate.Lang2+":"+cognate.Word2 {
				translit = cognate.Translit2
			}
		}

		return model.ChainWord{
			Word:         word,
			Translit1:    translit,
			LanguageInfo: adjustedLangInfo,
		}, nil
	}

	// Helper function to build chain recursively
	var buildChainRecursive func(startWord string, visited map[string]bool) []model.ChainWord
	buildChainRecursive = func(startWord string, visited map[string]bool) []model.ChainWord {
		if visited[startWord] {
			return nil
		}
		visited[startWord] = true

		parts := strings.Split(startWord, ":")
		if len(parts) != 2 {
			return nil
		}

		lang, word := parts[0], parts[1]
		var chain []model.ChainWord

		// Create word for current node with original cognate data
		wordObj, err := createChainWord(word, lang, startWord)
		if err == nil {
			chain = append(chain, wordObj)
		}

		// Explore connections
		for nextWord := range connections[startWord] {
			if !visited[nextWord] {
				subChain := buildChainRecursive(nextWord, visited)
				if len(subChain) > 0 {
					chain = append(chain, subChain...)
				}
			}
		}

		return chain
	}

	// Find starting points
	for startWord := range connections {
		if !processed[startWord] {
			visited := make(map[string]bool)
			chain := buildChainRecursive(startWord, visited)

			// Mark all words in chain as processed
			for _, word := range chain {
				processed[word.LanguageInfo.Code+":"+word.Word] = true
			}

			if len(chain) > 0 {
				chains = append(chains, model.CognateChain{Chain: chain})
			}
		}
	}

	return chains, nil
}

// FindByConceptID returns all cognates for a concept ID
func (cs *cognateSearch) FindByConceptID(ctx context.Context, conceptID string) ([]model.Cognate, error) {
	jsonStrings, err := cs.redisClient.LRange(ctx, fmt.Sprintf("concept:%s", conceptID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cognates: %w", err)
	}

	if len(jsonStrings) == 0 {
		return []model.Cognate{}, nil
	}

	cognates := make([]model.Cognate, 0, len(jsonStrings))
	for _, jsonStr := range jsonStrings {
		var cognate model.Cognate
		if err := json.Unmarshal([]byte(jsonStr), &cognate); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cognate: %w", err)
		}
		cognates = append(cognates, cognate)
	}

	return cognates, nil
}

func (cs *cognateSearch) FindCognateChains(ctx context.Context, conceptID, word, lang string) (*model.CognateChainResponse, error) {
	jsonStrings, err := cs.redisClient.LRange(ctx, fmt.Sprintf("concept:%s", conceptID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cognates: %w", err)
	}

	var cognates []model.Cognate
	for _, jsonStr := range jsonStrings {
		var cognate model.Cognate
		if err := json.Unmarshal([]byte(jsonStr), &cognate); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cognate: %w", err)
		}
		cognates = append(cognates, cognate)
	}

	chains, err := cs.buildChains(cognates, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build chains: %w", err)
	}

	// If word and language are provided, filter for specific chain
	if word != "" && lang != "" {
		var filteredChains []model.CognateChain
		for _, chain := range chains {
			// Check if this chain contains the specified word and language
			containsTarget := false
			for _, chainWord := range chain.Chain {
				if chainWord.Word == word && chainWord.LanguageInfo.Code == lang {
					containsTarget = true
					break
				}
			}
			if containsTarget {
				filteredChains = append(filteredChains, chain)
				break // We only need the first matching chain
			}
		}
		chains = filteredChains
	}

	return &model.CognateChainResponse{
		ConceptID: conceptID,
		Chains:    chains,
	}, nil
}
