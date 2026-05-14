//go:build js && wasm

package save

import "syscall/js"

type LocalStorageRepository struct {
	prefix string
}

func NewLocalStorageRepository(prefix string) *LocalStorageRepository {
	return &LocalStorageRepository{prefix: prefix}
}

func (r *LocalStorageRepository) key(slotID string) string {
	return r.prefix + slotID
}

func (r *LocalStorageRepository) Save(slotID string, data []byte) error {
	js.Global().Get("localStorage").Call("setItem", r.key(slotID), string(data))
	return nil
}

func (r *LocalStorageRepository) Load(slotID string) ([]byte, error) {
	val := js.Global().Get("localStorage").Call("getItem", r.key(slotID))
	if val.IsNull() || val.IsUndefined() {
		return nil, ErrNotFound
	}
	return []byte(val.String()), nil
}

func (r *LocalStorageRepository) Delete(slotID string) error {
	js.Global().Get("localStorage").Call("removeItem", r.key(slotID))
	return nil
}

func (r *LocalStorageRepository) Exists(slotID string) bool {
	val := js.Global().Get("localStorage").Call("getItem", r.key(slotID))
	return !val.IsNull() && !val.IsUndefined()
}
