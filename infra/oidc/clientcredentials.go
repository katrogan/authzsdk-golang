package oidc

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"userclouds.com/infra/ucerr"
)

// ClientCredentialsTokenSource encapsulates parameters required to issue a Client Credentials OIDC request and return a token
type ClientCredentialsTokenSource struct {
	TokenURL        string
	ClientID        string
	ClientSecret    string
	CustomAudiences []string
	SubjectJWT      string // optional, ID Token for a UC user if this access token is being created on their behalf
}

// GetToken issues a request to an OIDC-compliant token endpoint to perform
// the Client Credentials flow in exchange for an access token.
func (ccts ClientCredentialsTokenSource) GetToken() (string, error) {
	query := url.Values{}
	// TODO: move common OIDC values into constants
	query.Add("grant_type", "client_credentials")
	query.Add("client_id", ccts.ClientID)
	query.Add("client_secret", ccts.ClientSecret)
	for _, aud := range ccts.CustomAudiences {
		query.Add("audience", aud)
	}
	if ccts.SubjectJWT != "" {
		query.Add("subject_jwt", ccts.SubjectJWT)
	}
	req, err := http.NewRequest(http.MethodPost, ccts.TokenURL, strings.NewReader(query.Encode()))
	if err != nil {
		return "", ucerr.Wrap(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// TODO: re-use client?
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", ucerr.Wrap(err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		var oauthe ucerr.OAuthError
		if err := json.NewDecoder(resp.Body).Decode(&oauthe); err != nil {
			return "", ucerr.Wrap(err)
		}
		oauthe.Code = resp.StatusCode
		return "", ucerr.Wrap(oauthe)
	}
	var tresp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tresp); err != nil {
		return "", ucerr.Wrap(err)
	}
	return tresp.AccessToken, nil
}
