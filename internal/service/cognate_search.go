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
	GetWordSuggestions(ctx context.Context, prefix string) ([]WordSuggestion, error)
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

func (cs *cognateSearch) GetWordSuggestions(ctx context.Context, prefix string) ([]WordSuggestion, error) {
	if len(prefix) < 2 {
		return []WordSuggestion{}, nil
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
	suggestions := make([]WordSuggestion, 0)

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

		suggestions = append(suggestions, WordSuggestion{
			Word:      word,
			Language:  parts[1],
			ConceptID: parts[2],
		})

		if len(suggestions) >= limit {
			break
		}
	}

	return suggestions, nil
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
