package client

type qernalAPIClient struct {
	host  string
	token string
}

type QernalAPIClient interface {
	CreateOrganisation(req CreateOrganisationReq) (CreateOrganisationRes, error)
	ReadOrganisation(id string) (CreateOrganisationRes, error)
}

func New(host string, token string) (QernalAPIClient, error) {
	return qernalAPIClient{}, nil
}
