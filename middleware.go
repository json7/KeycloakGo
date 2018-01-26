package keycloak

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var oauthStateString string
var token *oauth2.Token

//HandleLogin is the keycloak login funtion
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	//create a random string for oath2 verification
	oauthStateString = randSeq(20)
	//Uses random gnerated string to verify keyclock security
	url := Oauth2Config.AuthCodeURL(oauthStateString)
	//redirects to loginCallback
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

//HandleLoginCallback is a fuction that verifies login success and forwards to index
func HandleLoginCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	//Checks that the strings are in a consistent state
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	//Gets the code from keycloak
	code := r.FormValue("code")
	//Exchanges code for token
	token, err = Oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("Code exchange failed with '%v'\n", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	client := &http.Client{}
	url := keycloakserver + "/auth/realms/" + Realm + "/protocol/openid-connect/userinfo"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	//Sends the token to get user info
	response, err := client.Do(req)
	//Checks if token and authentication were successful
	if err != nil || response.Status != "200 OK" {
		//forwards back to login if not successful
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	} else {
		//forwards to index if login sucessful
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
	return
}

//AuthMiddleware is a middlefuntion that verifies authentication before each redirect
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//If running unit tests skip authentication (temp)
		client := &http.Client{}
		url := keycloakserver + "/auth/realms/" + Realm + "/protocol/openid-connect/userinfo"
		req, _ := http.NewRequest("GET", url, nil)
		if token == nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		//Check if token is still valid
		response, err := client.Do(req)
		if err != nil || response.Status != "200 OK" {
			//Go to login if token is no longer valid
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		} else {
			//Go to redirect if token is still valid
			next.ServeHTTP(w, r)
		}
	})
	//return function for page handling
	return handler
}

//Logout logs the user out
func Logout(w http.ResponseWriter, r *http.Request) {
	//Makes the logout page redirect to login page
	URI := server + "/login"
	//Logout using endpoint and redirect to login page
	http.Redirect(w, r, keycloakserver+"/auth/realms/"+Realm+"/protocol/openid-connect/logout?redirect_uri="+URI, http.StatusTemporaryRedirect)

}
