package postgresql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/defskela/SocialNetwork/internal/config"
)

func TestNewClient(t *testing.T) {
	cfg := config.MustLoadPath("../../../configs/local.yaml")
	cfg.Postgres.Host = "localhost"

	ctx := context.Background()
	pool, err := NewClient(ctx, 3, &cfg.Postgres)
	assert.NoError(t, err)
	assert.NotNil(t, pool)

	if pool != nil {
		err = pool.Ping(ctx)
		assert.NoError(t, err)
		pool.Close()
	}
}

func TestDoWithTries(t *testing.T) {
	t.Run("Success immediately", func(t *testing.T) {
		attempts := 0
		err := DoWithTries(func() error {
			attempts++
			return nil
		}, 3, time.Millisecond)

		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("Success after retry", func(t *testing.T) {
		attempts := 0
		err := DoWithTries(func() error {
			attempts++
			if attempts < 2 {
				return errors.New("fail")
			}
			return nil
		}, 3, time.Millisecond)

		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("Fail all retries", func(t *testing.T) {
		attempts := 0
		err := DoWithTries(func() error {
			attempts++
			return errors.New("always fail")
		}, 3, time.Millisecond)

		assert.Error(t, err)
		assert.Equal(t, "always fail", err.Error())
		assert.Equal(t, 3, attempts)
	})
}
