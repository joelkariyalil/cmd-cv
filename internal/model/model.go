package model

import "time"

type CreateRequest struct {
	Message string `json:"message"`
}

type DeleteRequest struct {
	UID string `json:"uid"`
}

type RetrieveResponse struct {
	Message        string    `json:"message"`
	CreatedAt      time.Time `json:"created_at,omitzero"`
	LastAccessedAt time.Time `json:"last_accessed_at,omitzero"`
	ExpiresAt      time.Time `json:"expires_at,omitzero"`
}
