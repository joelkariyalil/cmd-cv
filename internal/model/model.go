package model

type CreateRequest struct {
	Message string `json:"message"`
}

type DeleteRequest struct {
	UID string `json:"uid"`
}

type RetrieveRequest struct {
	UID string `json:"uid"`
}
