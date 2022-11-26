package mclicrypto

import "time"

type SecretEntry struct {
	Name        string
	Login       string
	Secret      string
	Description string
	CreatedAt   time.Time
}
