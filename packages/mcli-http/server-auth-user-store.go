package mclihttp

import (
	"errors"
	"fmt"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_interface "mcli/packages/mcli-interface"
)

type UserStore struct {
	kvStore          mcli_interface.KVStorer
	CollectionPrefix string
	UserCache        map[string]*Credential
}

func NewUserStore(kvstore mcli_interface.KVStorer, collectionPrefix string) *UserStore {
	if collectionPrefix == "" {
		collectionPrefix = "userlist"
	}
	return &UserStore{kvStore: kvstore, CollectionPrefix: collectionPrefix}
}

func (us *UserStore) SetUser(u mcli_interface.Credentialer) error {
	user, ok := u.(*Credential)
	if !ok {
		return fmt.Errorf("method SetUser error - wrong input parameter for *Credential")
	}

	rawUserExist, _, exists := us.GetUser(user.Username)
	var userExist *Credential
	// password processing
	if exists {
		userExist, ok = rawUserExist.(*Credential)
		if !ok {
			return fmt.Errorf("existing user type mismatch")
		}
		user.Password = userExist.Password
	} else {
		hashedPassword, err := mcli_crypto.HashPassword(user.Password)
		if err != nil {
			return fmt.Errorf("error hashing password: %v", err)
		}
		user.Password = hashedPassword
	}

	// TODO: make user cache in memory with capacity and ttl

	err := us.kvStore.SetRecord(user.Username, user, us.CollectionPrefix)
	if err != nil {
		return err
	}
	return nil
}

func (us *UserStore) GetUser(username string) (mcli_interface.Credentialer, error, bool) {
	// TODO: make user cache in memory with capacity and ttl and search hear first
	user := Credential{}
	userRaw, err, ok := us.kvStore.GetRecord(username, us.CollectionPrefix)
	// fmt.Println("UserStore GetUser:", userRaw, err, ok)
	if err != nil || !ok {
		return &user, err, false
	}
	err = us.kvStore.GetUnMarshal()(userRaw, &user)
	if err != nil {
		return &user, err, false
	}
	return &user, nil, ok
}

func (us *UserStore) SetPassword(username, password string, expired bool) error {
	var err error
	var changed bool = false
	iUserExist, _, _ := us.GetUser(username)
	userExist, ok := iUserExist.(*Credential)
	if !ok {
		return fmt.Errorf("user %s does not exist", username)
	}
	if !(len(password) == 0 || password == "do_not_change") {
		hashedPassword, err := mcli_crypto.HashPassword(password)
		if err != nil {
			return fmt.Errorf("error hashing password: %v", err)
		}
		userExist.Password = hashedPassword
		changed = true
	}
	changed = changed || userExist.Expired != expired
	if changed {
		userExist.Expired = expired
		err = us.kvStore.SetRecord(userExist.Username, userExist, us.CollectionPrefix)
		if err != nil {
			return err
		}
	}
	return nil
}

func (us *UserStore) CheckPassword(username, password string) (bool, error) {
	iUser, err, ok := us.GetUser(username)
	// fmt.Println("UserStore CheckPassword:", user, err, ok)
	if err != nil || !ok {
		return false, err
	}
	user, ok := iUser.(*Credential)
	if !ok {
		return false, errors.New("check user password: type mismatch of existing record")
	}

	return mcli_crypto.CheckHashedPassword(user.Password, password)
}

func (us *UserStore) GetUsers(pattern string) (map[string]mcli_interface.Credentialer, error) {
	result := make(map[string]mcli_interface.Credentialer)
	users, err := us.kvStore.GetRecords(pattern, us.CollectionPrefix)
	if err != nil {
		return nil, err
	}

	for _, userRaw := range users {
		user := Credential{}
		err = us.kvStore.GetUnMarshal()(userRaw, &user)
		if err != nil {
			return nil, err
		}
		result[user.Username] = &user
	}

	return result, nil
}
