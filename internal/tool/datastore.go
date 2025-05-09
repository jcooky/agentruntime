package tool

import (
	"context"
	"sync"
)

type (
	CallData struct {
		Name      string `json:"name"`
		Arguments any    `json:"request"`
		Result    any    `json:"result"`
	}
	CallDataStore struct {
		callData []CallData
	}
	callDataStoreContextKeyType string
)

var (
	callDataStoreContextKey  = callDataStoreContextKeyType("ctx.callDataStore")
	lockCallDataStoreContext sync.Mutex
)

func WithEmptyCallDataStore(ctx context.Context) context.Context {
	return context.WithValue(ctx, callDataStoreContextKey, &CallDataStore{})
}

func appendCallData(ctx context.Context, callData CallData) {
	lockCallDataStoreContext.Lock()
	defer lockCallDataStoreContext.Unlock()

	var store *CallDataStore
	if v := ctx.Value(callDataStoreContextKey); v != nil {
		if v, ok := v.(*CallDataStore); ok {
			store = v
		}
	}
	if store == nil {
		return
	}

	store.callData = append(store.callData, callData)
}

func GetCallData(ctx context.Context) []CallData {
	store, ok := ctx.Value(callDataStoreContextKey).(*CallDataStore)
	if !ok {
		return nil
	}

	return store.callData
}
