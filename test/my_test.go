package test

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	redisCli "github.com/redis/go-redis/v9"
	redispkg "otel-eino-example/pkg/redis"
	"strconv"
	"testing"
)

func TestRetriever(t *testing.T) {
	ctx := context.Background()
	r, err := newRetriever(ctx)
	if err != nil {
		panic(err)
	}
	resp, err := r.Retrieve(ctx, "Eino是什么?")
	if err != nil {
		panic(err)
	}
	if len(resp) == 0 {
		fmt.Println("empty response")
	}
	for _, doc := range resp {
		t.Logf("doc: %+v", doc)
	}
}

// newRetriever component initialization function of node 'RedisRetriever' in graph 'EinoAgent'
func newRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	redisClient := redisCli.NewClient(&redisCli.Options{
		Addr:     "localhost:6379",
		Protocol: 2,
	})
	config := &redis.RetrieverConfig{
		Client:       redisClient,
		Index:        fmt.Sprintf("%s%s", redispkg.RedisPrefix, redispkg.IndexName),
		Dialect:      2,
		ReturnFields: []string{redispkg.ContentField, redispkg.MetadataField, redispkg.DistanceField},
		TopK:         8,
		VectorField:  redispkg.VectorField,
		DocumentConverter: func(ctx context.Context, doc redisCli.Document) (*schema.Document, error) {
			resp := &schema.Document{
				ID:       doc.ID,
				Content:  "",
				MetaData: map[string]any{},
			}
			for field, val := range doc.Fields {
				if field == redispkg.ContentField {
					resp.Content = val
				} else if field == redispkg.MetadataField {
					resp.MetaData[field] = val
				} else if field == redispkg.DistanceField {
					distance, err := strconv.ParseFloat(val, 64)
					if err != nil {
						continue
					}
					resp.WithScore(1 - distance)
				}
			}

			return resp, nil
		},
	}
	embeddingIns11, err := newEmbedding(ctx)
	if err != nil {
		return nil, err
	}
	config.Embedding = embeddingIns11
	rtr, err = redis.NewRetriever(ctx, config)
	if err != nil {
		return nil, err
	}
	return rtr, nil
}

func newEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	config := &openai.EmbeddingConfig{
		Model:   "Qwen/Qwen3-Embedding-8B",
		APIKey:  "sk-gtfdhlgslcalhsawczszderfcwtjfqzwxweljkexdeparsbu",
		BaseURL: "https://api.siliconflow.cn/v1",
	}
	eb, err = openai.NewEmbedder(ctx, config)
	if err != nil {
		return nil, err
	}
	return eb, nil
}
