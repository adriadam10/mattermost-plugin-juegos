package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mattermost/mattermost/server/public/plugin"
)

// Several people (and the ticker) change the same keys at once: every write is
// a compare-and-set, retried on a clash.
const kvRetries = 8

var errNotFound = errors.New("not found")

func loadJSON[T any](api plugin.API, key string) (*T, []byte, error) {
	raw, appErr := api.KVGet(key)
	if appErr != nil {
		return nil, nil, appErr
	}
	v := new(T)
	if raw == nil {
		return v, nil, nil
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return nil, nil, fmt.Errorf("decode %s: %w", key, err)
	}
	return v, raw, nil
}

// updateJSON applies change to the value at key and saves it. With a nil
// previous value change gets a zero T, unless mustExist. An error from change
// is returned as is and nothing is saved.
func updateJSON[T any](api plugin.API, key string, mustExist bool, change func(*T) error) (before, after T, err error) {
	for range kvRetries {
		v, old, err := loadJSON[T](api, key)
		if err != nil {
			return before, after, err
		}
		if old == nil && mustExist {
			return before, after, errNotFound
		}
		if err := json.Unmarshal(orNull(old), &before); err != nil {
			return before, after, err
		}
		if err := change(v); err != nil {
			return before, *v, err
		}
		updated, err := json.Marshal(v)
		if err != nil {
			return before, after, err
		}
		ok, appErr := api.KVCompareAndSet(key, old, updated)
		if appErr != nil {
			return before, after, appErr
		}
		if ok {
			return before, *v, nil
		}
		var zero T
		before = zero
	}
	return before, after, fmt.Errorf("%s kept changing while saving it", key)
}

func orNull(raw []byte) []byte {
	if raw == nil {
		return []byte("null")
	}
	return raw
}
