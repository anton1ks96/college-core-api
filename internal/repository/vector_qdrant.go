package repository

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/domain"
	qdrantpkg "github.com/anton1ks96/college-core-api/pkg/database/qdrant"
	"github.com/qdrant/go-client/qdrant"
)

type VectorQdrantRepository struct {
	client     *qdrant.Client
	collection string
}

func NewVectorRepository(cfg *config.Config, q *qdrantpkg.Qdrant) *VectorQdrantRepository {
	return &VectorQdrantRepository{
		client:     q.Client,
		collection: cfg.Qdrant.Collection,
	}
}

func (r *VectorQdrantRepository) EnsureCollection(ctx context.Context, vectorSize uint64) error {
	exists, err := r.client.CollectionExists(ctx, r.collection)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	if !exists {
		err = r.client.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: r.collection,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     vectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
	}

	return nil
}

func (r *VectorQdrantRepository) UpsertChunks(
	ctx context.Context,
	datasetID string,
	version int,
	title string,
	chunks []domain.ChunkData,
	vectors [][]float32,
) (int, error) {
	points := make([]*qdrant.PointStruct, 0, len(chunks))

	for i, ch := range chunks {
		pid := pointID(datasetID, version, ch.Index)

		points = append(points, &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(pid),
			Vectors: qdrant.NewVectorsDense(vectors[i]),
			Payload: qdrant.NewValueMap(map[string]any{
				"dataset_id": datasetID,
				"version":    version,
				"chunk_id":   ch.Index,
				"title":      title,
				"text":       ch.Text,
			}),
		})
	}

	_, err := r.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: r.collection,
		Points:         points,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to upsert points: %w", err)
	}

	return len(points), nil
}

func (r *VectorQdrantRepository) Search(
	ctx context.Context,
	datasetID string,
	version int,
	queryVector []float32,
	k uint64,
) ([]domain.SearchHit, error) {
	filter := &qdrant.Filter{
		Must: []*qdrant.Condition{
			qdrant.NewMatch("dataset_id", datasetID),
			qdrant.NewMatchInt("version", int64(version)),
		},
	}

	scored, err := r.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: r.collection,
		Query:          qdrant.NewQueryDense(queryVector),
		Filter:         filter,
		Limit:          qdrant.PtrOf(k),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search points: %w", err)
	}

	hits := make([]domain.SearchHit, 0, len(scored))
	for _, sp := range scored {
		hit := domain.SearchHit{
			Score: sp.Score,
		}

		if v, ok := sp.Payload["dataset_id"]; ok {
			hit.DatasetID = v.GetStringValue()
		}
		if v, ok := sp.Payload["version"]; ok {
			hit.Version = int(v.GetIntegerValue())
		}
		if v, ok := sp.Payload["chunk_id"]; ok {
			hit.ChunkID = int(v.GetIntegerValue())
		}
		if v, ok := sp.Payload["title"]; ok {
			hit.Title = v.GetStringValue()
		}
		if v, ok := sp.Payload["text"]; ok {
			hit.Text = v.GetStringValue()
		}

		hits = append(hits, hit)
	}

	return hits, nil
}

func (r *VectorQdrantRepository) DeleteByDatasetID(ctx context.Context, datasetID string) error {
	filter := &qdrant.Filter{
		Must: []*qdrant.Condition{
			qdrant.NewMatch("dataset_id", datasetID),
		},
	}

	_, err := r.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: r.collection,
		Points:         qdrant.NewPointsSelectorFilter(filter),
	})
	if err != nil {
		return fmt.Errorf("failed to delete points for dataset %s: %w", datasetID, err)
	}

	return nil
}

// pointID генерирует детерминированный uint64 ID точки из dataset_id, version и chunk_id
// Берёт MD5-хеш строки "dataset_id:version:chunk_id" и использует первые 6 байт (48 бит) как число
func pointID(datasetID string, version, chunkID int) uint64 {
	sum := md5.Sum([]byte(fmt.Sprintf("%s:%d:%d", datasetID, version, chunkID)))

	var id uint64
	for i := 0; i < 6; i++ {
		id = (id << 8) | uint64(sum[i])
	}

	return id
}
