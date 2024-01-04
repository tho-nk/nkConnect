// internal/server/http.go
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"

	"nkConnect/internal/app"
	"nkConnect/internal/oauth"
)

var manager *manage.Manager
var srv *server.Server

// Run initializes and runs the OAuth2 server.
func Run() {
	manager = manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.SetClientTokenCfg(&manage.Config{AccessTokenExp: 3600}) // Set the access token expiration time

	// token memory store
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// client memory store
	manager.MapClientStorage(app.ClientStore)

	srv = server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	// Register HTTP endpoints...
	http.HandleFunc("/register/application", handleRegisterApplication)
	http.HandleFunc("/register/client", handleRegisterClient)
	http.HandleFunc("/token", handleTokenRequest)
	http.HandleFunc("/inspect", handleInspectToken)
	http.HandleFunc("/validate", handleValidateToken)

	log.Fatal(http.ListenAndServe(":9096", nil))
}

// Add your endpoint handler functions here...

// handleRegisterApplication handles the "/register/application" endpoint.
func handleRegisterApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var requestData struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}

	if err := decoder.Decode(&requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if requestData.Name == "" || len(requestData.Scopes) == 0 {
		http.Error(w, "Name and Scopes are required", http.StatusBadRequest)
		return
	}

	// Register the application
	appID, err := app.GetApplicationStore().RegisterApplication(requestData.Name, requestData.Scopes)
	if err != nil {
		http.Error(w, "Failed to register application", http.StatusInternalServerError)
		return
	}

	responseData := map[string]string{"ApplicationID": appID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

// handleRegisterClient handles the "/register/client" endpoint.
func handleRegisterClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var requestData struct {
		ApplicationName string   `json:"application_name"`
		Scopes          []string `json:"scopes"`
	}

	if err := decoder.Decode(&requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Check if the requested application exists
	application, err := app.GetApplicationStore().GetApplicationByName(requestData.ApplicationName)
	if err != nil {
		http.Error(w, "Requested application does not exist", http.StatusBadRequest)
		return
	}

	// Check if the requested scopes are available for the application
	if !oauth.AreScopesAvailable(application.Scopes, requestData.Scopes) {
		http.Error(w, "Invalid or unavailable scopes requested", http.StatusBadRequest)
		return
	}

	// Call the RegisterClient function to handle client registration
	newClient, err := app.GetApplicationStore().RegisterClient(application, requestData.Scopes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to register client: %v", err), http.StatusInternalServerError)
		return
	}

	responseData := map[string]string{"CLIENT_ID": newClient.ID, "CLIENT_SECRET": newClient.Secret}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

// handleTokenRequest handles the "/token" endpoint.
func handleTokenRequest(w http.ResponseWriter, r *http.Request) {
	srv.HandleTokenRequest(w, r)
}

// handleInspectToken handles the "/inspect" endpoint.
func handleInspectToken(w http.ResponseWriter, r *http.Request) {
	// Implementation...
}

// handleValidateToken handles the "/validate" endpoint.
func handleValidateToken(w http.ResponseWriter, r *http.Request) {
	// Implementation...
}
