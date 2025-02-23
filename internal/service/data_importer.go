package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"cognet-world-inquiry-service/internal/model"

	"github.com/redis/go-redis/v9"
)

type DataImporter interface {
	ImportFromReader(ctx context.Context, reader *bufio.Reader) error
	GetImportStatus() string
	ClearDatabase(ctx context.Context) error
}

type dataImporter struct {
	redisClient *redis.Client
	status      string
}

func generatePrefixes(word string) []string {
	word = strings.ToLower(word) // Normalize to lowercase for better searching
	var prefixes []string
	runes := []rune(word) // Handle UTF-8 characters properly

	// Generate prefixes starting from minimum 2 characters
	for i := 2; i <= len(runes); i++ {
		prefixes = append(prefixes, string(runes[:i]))
	}
	return prefixes
}

func NewDataImporter(redisClient *redis.Client) DataImporter {
	return &dataImporter{
		redisClient: redisClient,
		status:      "ready",
	}
}

func (d *dataImporter) ImportFromReader(ctx context.Context, reader *bufio.Reader) error {
	d.status = "importing"
	defer func() { d.status = "ready" }()

	// Skip header
	if _, err := reader.ReadString('\n'); err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	pipeline := d.redisClient.Pipeline()
	batchSize := 1000
	count := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading line: %w", err)
		}

		fields := strings.Split(strings.TrimSpace(line), "\t")
		if len(fields) < 5 {
			continue
		}

		cognate := model.Cognate{
			ConceptID: fields[0],
			Lang1:     fields[1],
			Word1:     fields[2],
			Lang2:     fields[3],
			Word2:     fields[4],
		}

		if len(fields) > 5 {
			cognate.Translit1 = fields[5]
		}
		if len(fields) > 6 {
			cognate.Translit2 = fields[6]
		}

		// Convert to JSON for storage
		jsonData, err := json.Marshal(cognate)
		if err != nil {
			return fmt.Errorf("failed to marshal cognate: %w", err)
		}

		// 1. Store complete cognate data
		pipeline.RPush(ctx, fmt.Sprintf("concept:%s", cognate.ConceptID), jsonData)

		// 2. Create prefix indices for both words
		wordInfo1 := fmt.Sprintf("%s|%s", cognate.ConceptID, cognate.Lang1)
		wordInfo2 := fmt.Sprintf("%s|%s", cognate.ConceptID, cognate.Lang2)

		// Generate and store prefixes for Word1
		for _, prefix := range generatePrefixes(cognate.Word1) {
			prefixKey := fmt.Sprintf("prefix:%s", prefix)
			pipeline.SAdd(ctx, prefixKey, fmt.Sprintf("%s|%s|%s", cognate.Word1, cognate.Lang1, cognate.ConceptID))
		}

		// Generate and store prefixes for Word2
		for _, prefix := range generatePrefixes(cognate.Word2) {
			prefixKey := fmt.Sprintf("prefix:%s", prefix)
			pipeline.SAdd(ctx, prefixKey, fmt.Sprintf("%s|%s|%s", cognate.Word2, cognate.Lang2, cognate.ConceptID))
		}

		// 3. Store full word indices (for exact matches)
		pipeline.SAdd(ctx, fmt.Sprintf("word:%s", cognate.Word1), wordInfo1)
		pipeline.SAdd(ctx, fmt.Sprintf("word:%s", cognate.Word2), wordInfo2)

		count++

		// Execute pipeline in batches
		if count%batchSize == 0 {
			if _, err := pipeline.Exec(ctx); err != nil {
				return fmt.Errorf("failed to execute pipeline: %w", err)
			}
			pipeline = d.redisClient.Pipeline()
		}
	}

	// Execute remaining commands
	if count%batchSize != 0 {
		if _, err := pipeline.Exec(ctx); err != nil {
			return fmt.Errorf("failed to execute final pipeline: %w", err)
		}
	}

	// Store import metadata
	metadata := map[string]interface{}{
		"total_records": count,
		"status":        "completed",
		"timestamp":     time.Now().Unix(),
	}

	metadataJSON, _ := json.Marshal(metadata)
	if err := d.redisClient.Set(ctx, "import:metadata", metadataJSON, 0).Err(); err != nil {
		return fmt.Errorf("failed to store metadata: %w", err)
	}

	return nil
}

func (d *dataImporter) GetImportStatus() string {
	return d.status
}

func (d *dataImporter) ClearDatabase(ctx context.Context) error {
	return d.redisClient.FlushAll(ctx).Err()
}
