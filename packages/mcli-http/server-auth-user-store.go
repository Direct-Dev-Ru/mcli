package mclihttp

import (
	"encoding/json"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_store "mcli/packages/mcli-store"
)

type UserStore struct {
	kvStore          mcli_store.KVStorer
	CollectionPrefix string
}

func NewUserStore(kvstore mcli_store.KVStorer, collectionPrefix string) *UserStore {
	if collectionPrefix == "" {
		collectionPrefix = "userlist"
	}
	return &UserStore{kvStore: kvstore, CollectionPrefix: collectionPrefix}
}

func (us *UserStore) GetUser(username string) (*Credential, error, bool) {
	user := Credential{}
	userRaw, err, ok := us.kvStore.GetRecord(username)
	// fmt.Println("UserStore GetUser:", userRaw, err, ok)
	if err != nil || !ok {
		return &user, err, false
	}
	err = json.Unmarshal([]byte(userRaw), &user)
	if err != nil {
		return &user, err, false
	}
	return &user, nil, ok
}

func (us *UserStore) SetPassword(username, password string) error {
	return nil
}

func (us *UserStore) CheckPassword(username, password string) (bool, error) {
	user, err, ok := us.GetUser(username)
	// fmt.Println("UserStore CheckPassword:", user, err, ok)
	if err != nil || !ok {
		return false, err
	}
	return mcli_crypto.CheckHashedPassword(user.Password, password)
}

func (us *UserStore) GetAllUsers(pattern string) ([]*Credential, error) {

	return nil, nil
}
