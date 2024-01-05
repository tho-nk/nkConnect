// internal/app/client.go
package app

import (
	"gopkg.in/oauth2.v3/store"
)

// Client represents a custom client model that includes scopes.
type Client struct {
	ID                string
	Secret            string
	Domain            string
	ApplicationScopes map[*Application][]string
}

// GetID implements the oauth2.ClientInfo interface.
func (c *Client) GetID() string {
	return c.ID
}

// GetSecret implements the oauth2.ClientInfo interface.
func (c *Client) GetSecret() string {
	return c.Secret
}

// GetDomain implements the oauth2.ClientInfo interface.
func (c *Client) GetDomain() string {
	return c.Domain
}

// GetUserID implements the oauth2.ClientInfo interface.
func (c *Client) GetUserID() string {
	// You might return some default user ID or an empty string depending on your use case.
	return ""
}

var clientStoreInstance *store.ClientStore

func newClientStore() *store.ClientStore {
	return store.NewClientStore()
}

// GetApplicationStore returns the singleton instance of ApplicationStore.
func GetClientStore() *store.ClientStore {
	if clientStoreInstance == nil {
		clientStoreInstance = newClientStore()
	}
	return clientStoreInstance
}
