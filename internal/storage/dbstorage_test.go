package storage

import (
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestDBStorage_WithDBRetry(t *testing.T) {
	type testCase struct {
		name           string
		failCount      int
		retriable      bool
		expectAttempts int
		expectError    bool
	}

	testCases := []testCase{
		{
			name:           "non-retriable_1_ko",
			failCount:      1,
			retriable:      false,
			expectAttempts: 1,
			expectError:    true,
		},
		{
			name:           "retriable_2_ok",
			failCount:      2,
			retriable:      true,
			expectAttempts: 3,
			expectError:    false,
		},
		{
			name:           "retriable_6_ko",
			failCount:      5,
			retriable:      true,
			expectAttempts: 4,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attempts := 0

			err := withDBRetry(func() error {
				attempts++
				if attempts <= tc.failCount {
					if tc.retriable {
						return &pgconn.PgError{Code: pgerrcode.ConnectionFailure}
					}
					return errors.New("non-retriable error")
				}
				return nil
			}, tc.name)

			assert.Equal(t, tc.expectAttempts, attempts, "не совпадает количество попыток")
			if tc.expectError {
				assert.Error(t, err, "ожидалась ошибка")
			} else {
				assert.NoError(t, err, "ошибки быть не должно")
			}
		})
	}
}
