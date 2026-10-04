package cacherepository

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
)

const (
	obiSettingsKey    = "obi:settings"
	obiSettingsGenKey = "obi:settings:gen"
)

// OBISettingsRepo caches the settings the agents run OBI by, which every agent
// reads every 30 seconds: from Redis, not the database.
//
// An entry carries the generation it was read at. Every change to the
// settings moves the generation, once committed, and drops the entry: an agent
// that read the database before the change and caches what it read after it
// caches it under the old generation, which is not taken.
type OBISettingsRepo interface {
	// Get answers the cached settings of the current generation, nil when there
	// are none; and that generation, to Set what is read from the database
	// with.
	Get(ctx context.Context) (*entity.OBIAgentSettings, int64, error)
	// Set caches settings read from the database at a generation, for ttl.
	Set(ctx context.Context, settings *entity.OBIAgentSettings, gen int64, ttl time.Duration) error
	// Invalidate moves the generation and drops the entry: the settings
	// changed. It is called once the change is committed.
	Invalidate(ctx context.Context) error
}

type obiSettingsRepo struct {
	client rediscache.Client
}

func NewOBISettingsRepo(client rediscache.Client) OBISettingsRepo {
	return &obiSettingsRepo{client: client}
}

// obiSettingsEntry is the cached value: the settings, and the generation they
// were read at.
type obiSettingsEntry struct {
	Gen      int64                    `json:"gen"`
	Settings *entity.OBIAgentSettings `json:"settings"`
}

func (repo *obiSettingsRepo) Get(ctx context.Context) (*entity.OBIAgentSettings, int64, error) {
	values, err := repo.client.MGet(ctx, obiSettingsGenKey, obiSettingsKey).Result()
	if err != nil {
		return nil, 0, hperrors.Wrap(err)
	}
	gen, err := obiSettingsGen(values[0])
	if err != nil {
		return nil, 0, err
	}
	s, ok := values[1].(string)
	if !ok {
		return nil, gen, nil
	}
	var entry obiSettingsEntry
	if err = json.Unmarshal([]byte(s), &entry); err != nil {
		// Unreadable - written by another version - as if there were none.
		return nil, gen, nil //nolint:nilerr // a miss: the settings are read from the database
	}
	if entry.Gen != gen || entry.Settings == nil {
		// Read before the last change: as if there were none.
		return nil, gen, nil
	}
	return entry.Settings, gen, nil
}

// obiSettingsGen is the generation as MGET answers it: none is 0.
func obiSettingsGen(value any) (int64, error) {
	s, ok := value.(string)
	if !ok {
		return 0, nil
	}
	gen, err := strconv.ParseInt(s, 10, 64)
	return gen, hperrors.Wrap(err)
}

func (repo *obiSettingsRepo) Set(
	ctx context.Context,
	settings *entity.OBIAgentSettings,
	gen int64,
	ttl time.Duration,
) error {
	data, err := json.Marshal(obiSettingsEntry{Gen: gen, Settings: settings})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(repo.client.Set(ctx, obiSettingsKey, data, ttl).Err())
}

func (repo *obiSettingsRepo) Invalidate(ctx context.Context) error {
	_, err := repo.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Incr(ctx, obiSettingsGenKey)
		pipe.Del(ctx, obiSettingsKey)
		return nil
	})
	return hperrors.Wrap(err)
}
