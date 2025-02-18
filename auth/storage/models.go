// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package storage

import (
	"database/sql/driver"
	"fmt"
	"time"

	uuid "github.com/gofrs/uuid"
	null "gopkg.in/guregu/null.v4"
)

type IdentityStatus string

const (
	IdentityStatusActive  IdentityStatus = "active"
	IdentityStatusBlocked IdentityStatus = "blocked"
)

func (e *IdentityStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = IdentityStatus(s)
	case string:
		*e = IdentityStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for IdentityStatus: %T", src)
	}
	return nil
}

type NullIdentityStatus struct {
	IdentityStatus IdentityStatus `json:"identityStatus"`
	Valid          bool           `json:"valid"` // Valid is true if IdentityStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullIdentityStatus) Scan(value interface{}) error {
	if value == nil {
		ns.IdentityStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.IdentityStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullIdentityStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.IdentityStatus), nil
}

func AllIdentityStatusValues() []IdentityStatus {
	return []IdentityStatus{
		IdentityStatusActive,
		IdentityStatusBlocked,
	}
}

type AccessToken struct {
	Hash       string    `db:"hash" json:"hash"`
	IdentityID uuid.UUID `db:"identity_id" json:"identityId"`
	SessionID  uuid.UUID `db:"session_id" json:"sessionId"`
	Data       []byte    `db:"data" json:"data"`
	RevokedAt  null.Time `db:"revoked_at" json:"revokedAt"`
	ExpiresAt  time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
}

type Credential struct {
	ID             uuid.UUID `db:"id" json:"id"`
	IdentityID     uuid.UUID `db:"identity_id" json:"identityId"`
	CredentialHash string    `db:"credential_hash" json:"credentialHash"`
	Type           string    `db:"type" json:"type"`
	ExpiredAt      null.Time `db:"expired_at" json:"expiredAt"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
}

type Identity struct {
	ID        uuid.UUID      `db:"id" json:"id"`
	Identity  string         `db:"identity" json:"identity"`
	UserID    uuid.UUID      `db:"user_id" json:"userId"`
	Status    IdentityStatus `db:"status" json:"status"`
	Data      []byte         `db:"data" json:"data"`
	UpdatedAt time.Time      `db:"updated_at" json:"updatedAt"`
	CreatedAt time.Time      `db:"created_at" json:"createdAt"`
}

type RefreshToken struct {
	Hash      string    `db:"hash" json:"hash"`
	SessionID uuid.UUID `db:"session_id" json:"sessionId"`
	Data      []byte    `db:"data" json:"data"`
	RevokedAt null.Time `db:"revoked_at" json:"revokedAt"`
	UsedAt    null.Time `db:"used_at" json:"usedAt"`
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type Session struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"userId"`
	IdentityID uuid.UUID `db:"identity_id" json:"identityId"`
	Data       []byte    `db:"data" json:"data"`
	ExpiresAt  time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
}
