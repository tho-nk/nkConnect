// internal/oauth/oauth.go
package oauth

// Helper function to check if the requested scopes are available for the application.
func AreScopesAvailable(applicationScopes, requestedScopes []string) bool {
	for _, scope := range requestedScopes {
		if !contains(applicationScopes, scope) {
			return false
		}
	}
	return true
}

// Helper function to check if a slice contains a specific element.
func contains(slice []string, element string) bool {
	for _, e := range slice {
		if e == element {
			return true
		}
	}
	return false
}
