// internal/app/application.go
package app

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"gopkg.in/oauth2.v3/store"
)

var ClientStore *store.ClientStore

// MyApplication represents an application with associated scopes.
type MyApplication struct {
	ID      string
	Name    string
	Scopes  []string
	Clients []*MyClient
}

// MyClient represents a custom client model that includes scopes.
type MyClient struct {
	ID                string
	Secret            string
	Domain            string
	ApplicationScopes map[*MyApplication][]string
}

// GetID implements the oauth2.ClientInfo interface.
func (c *MyClient) GetID() string {
	return c.ID
}

// GetSecret implements the oauth2.ClientInfo interface.
func (c *MyClient) GetSecret() string {
	return c.Secret
}

// GetDomain implements the oauth2.ClientInfo interface.
func (c *MyClient) GetDomain() string {
	return c.Domain
}

// GetUserID implements the oauth2.ClientInfo interface.
func (c *MyClient) GetUserID() string {
	// You might return some default user ID or an empty string depending on your use case.
	return ""
}

// ApplicationStore is a simple in-memory store for applications.
type ApplicationStore struct {
	mu           sync.Mutex
	applications map[string]MyApplication
}

var applicationStoreInstance *ApplicationStore

// NewApplicationStore creates a new ApplicationStore instance.
func NewApplicationStore() *ApplicationStore {
	return &ApplicationStore{
		applications: make(map[string]MyApplication),
	}
}

// GetApplicationStore returns the singleton instance of ApplicationStore.
func GetApplicationStore() *ApplicationStore {
	if applicationStoreInstance == nil {
		applicationStoreInstance = NewApplicationStore()
	}
	return applicationStoreInstance
}

// RegisterApplication registers a new application.
func (s *ApplicationStore) RegisterApplication(name string, scopes []string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the application with the given name already exists
	if _, exists := s.applications[name]; exists {
		return "", fmt.Errorf("application with name '%s' already exists", name)
	}

	// Generate a unique ID (UUID) for the new application
	appID := uuid.New().String()

	application := MyApplication{
		ID:      appID,
		Name:    name,
		Scopes:  scopes,
		Clients: nil,
	}

	s.applications[name] = application

	return appID, nil
}

// GetApplicationByID retrieves an application by its ID.
func (s *ApplicationStore) GetApplicationByID(appID string) (MyApplication, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, app := range s.applications {
		if app.ID == appID {
			return app, nil
		}
	}

	return MyApplication{}, fmt.Errorf("application not found")
}

// GetApplicationByName retrieves an application by its name.
func (s *ApplicationStore) GetApplicationByName(name string) (MyApplication, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	application, ok := s.applications[name]
	if !ok {
		return MyApplication{}, fmt.Errorf("application not found")
	}

	return application, nil
}

// RegisterClient registers a new client for the given application and scopes.
func (s *ApplicationStore) RegisterClient(application MyApplication, scopes []string) (MyClient, error) {
	// Here, you can generate a client ID and secret, store them, and return the client information.
	// Example:
	clientID := generateClientID(application.Name, scopes)
	clientSecret := generateClientSecret(application.Name, scopes)

	client := MyClient{
		ID:                clientID,
		Secret:            clientSecret,
		Domain:            "http://localhost:9094",
		ApplicationScopes: make(map[*MyApplication][]string),
	}
	client.ApplicationScopes[&application] = scopes

	err := ClientStore.Set(clientID, &client)
	if err != nil {
		return MyClient{}, fmt.Errorf("failed to register client: %v", err)
	}

	return client, nil
}

// generateClientID generates a unique client ID based on application name and scopes.
func generateClientID(applicationName string, scopes []string) string {
	uuidComponent := uuid.New().String()[:4]
	hashedApplication := hashData(applicationName)
	hashedScopes := hashData(strings.Join(scopes, "_"))
	return "client_" + hashedApplication + "_" + hashedScopes + "_" + uuidComponent
}

// generateClientSecret generates a unique client secret based on application name and scopes.
func generateClientSecret(applicationName string, scopes []string) string {
	uuidComponent := uuid.New().String()[:4]
	hashedApplication := hashData(applicationName)
	hashedScopes := hashData(strings.Join(scopes, "_"))
	return "secret_" + hashedApplication + "_" + hashedScopes + "_" + uuidComponent
}

// hashData hashes the given data using MD5.
func hashData(data string) string {
	hasher := md5.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Initialize ClientStore
func init() {
	ClientStore = store.NewClientStore()
}
