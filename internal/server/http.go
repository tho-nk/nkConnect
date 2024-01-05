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
	"nkConnect/internal/utility"
)

type HttpServer struct {
	manager *manage.Manager
	srv     *server.Server
}

var httpServerInstance *HttpServer

func NewHttpServer() *HttpServer {
	httpServer := &HttpServer{}
	httpServer.manager = manage.NewDefaultManager()
	httpServer.manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	// httpServer.manager.SetClientTokenCfg(&manage.Config{AccessTokenExp: 3600}) // Set the access token expiration time
	httpServer.manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)
	// token memory store
	httpServer.manager.MustTokenStorage(store.NewMemoryTokenStore())
	// client memory store
	httpServer.manager.MapClientStorage(app.GetClientStore())

	httpServer.srv = server.NewDefaultServer(httpServer.manager)
	httpServer.srv.SetAllowGetAccessRequest(true)
	httpServer.srv.SetClientInfoHandler(server.ClientFormHandler)
	httpServer.srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})
	httpServer.srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	httpServer.initHandle()
	return httpServer
}

// GetApplicationStore returns the singleton instance of ApplicationStore.
func GetHttpServerInstance() *HttpServer {
	if httpServerInstance == nil {
		httpServerInstance = NewHttpServer()
	}
	return httpServerInstance
}

// GetUserID implements the oauth2.ClientInfo interface.
func (httpServer *HttpServer) initHandle() {
	// Register HTTP endpoints...
	http.HandleFunc("/nk-connect/register/application", httpServer.handleRegisterApplication)
	http.HandleFunc("/nk-connect/register/client", httpServer.handleRegisterClient)
	http.HandleFunc("/nk-connect/token", httpServer.handleTokenRequest)
	http.HandleFunc("/nk-connect/inspect", httpServer.handleInspectToken)
	http.HandleFunc("/nk-connect/validate", httpServer.handleValidateToken)

}

// Run initializes and runs the OAuth2 server.
func (httpServer *HttpServer) Run() {
	log.Fatal(http.ListenAndServe(":9096", nil))
}

// handleRegisterApplication handles the "/register/application" endpoint.
func (httpServer *HttpServer) handleRegisterApplication(w http.ResponseWriter, r *http.Request) {
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
func (httpServer *HttpServer) handleRegisterClient(w http.ResponseWriter, r *http.Request) {
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
	if !utility.AreScopesAvailable(application.Scopes, requestData.Scopes) {
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
func (httpServer *HttpServer) handleTokenRequest(w http.ResponseWriter, r *http.Request) {
	err := httpServer.srv.HandleTokenRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleInspectToken handles the "/inspect" endpoint.
func (httpServer *HttpServer) handleInspectToken(w http.ResponseWriter, r *http.Request) {
	// Parse the token from the request
	token, err := httpServer.srv.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"active":    true, // Indicates whether the token is active
		"client_id": token.GetClientID(),
		"exp":       token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Unix(),
		// Add other relevant token information here
	}

	// Convert the response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

// handleValidateToken handles the "/validate" endpoint.
func (httpServer *HttpServer) handleValidateToken(w http.ResponseWriter, r *http.Request) {
	// Parse the token from the request
	_, err := httpServer.srv.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	// Token is valid, return a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Token is valid"))
}
