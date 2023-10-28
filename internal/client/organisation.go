package client

import (
	"github.com/google/uuid"
	"time"
)

var organisationData = map[string]CreateOrganisationRes{}

func (q qernalAPIClient) CreateOrganisation(req CreateOrganisationReq) (CreateOrganisationRes, error) {
	id := uuid.NewString()

	organisation := CreateOrganisationRes{
		ID:     id,
		Name:   req.Name,
		UserID: uuid.NewString(),
		Date: struct {
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}{CreatedAt: time.Now().String(), UpdatedAt: time.Now().String()},
	}
	organisationData[id] = organisation
	return organisation, nil
}

func (q qernalAPIClient) ReadOrganisation(id string) (CreateOrganisationRes, error) {
	return organisationData[id], nil
}

type CreateOrganisationRes struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	UserID string `json:"user_id"`
	Date   struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
}

type CreateOrganisationReq struct {
	Name string `json:"name"`
}
