package neverbounce

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const DEFAULT_API_URL string = "https://api.neverbounce.com/v3"

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	Expires     int    `json:"expires"`
}

type VerifyEmailResponse struct {
	Success       bool   `json:"success"`
	Result        int    `json:"result"`
	ResultDetails int    `json:"result_details"`
	Msg           string `json:"msg"`
}

var NeverBounce *NeverBounceCli

type NeverBounceCli struct {
	ApiUrl string

	ApiUsername string
	ApiPassword string
	AccessToken string
}

// Sets the API url on the client
func (n *NeverBounceCli) SetApiUrl(url string) {
	n.ApiUrl = url
}

// This function will get the client access token but also
// store it in the struct to be used in subsequent calls
func (n *NeverBounceCli) GetAccessToken() string {
	log.Println("Requesting access token")

	request, _ := http.NewRequest(
		"POST",
		n.ApiUrl+"/access_token",
		bytes.NewReader([]byte(url.Values{
			"grant_type": {"client_credentials"},
			"scope":      {"basic user"},
		}.Encode())),
	)

	request.SetBasicAuth(n.ApiUsername, n.ApiPassword)

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		log.Panic(err)
	}

	var accessTokenResponse AccessTokenResponse
	decoder := json.NewDecoder(response.Body)

	if response.StatusCode != 200 {
		log.Panic("Client returned status: ", response.StatusCode)
		return ""
	}

	if err := decoder.Decode(&accessTokenResponse); err != nil {
		log.Panic(err)
	} else {
		n.AccessToken = accessTokenResponse.AccessToken
	}

	return accessTokenResponse.AccessToken
}

// Takes an email and verifies it
func (n *NeverBounceCli) VerifyEmail(email string) VerifyEmailResponse {
	log.Println("Verifying email ", email)

	response, err := http.PostForm(
		n.ApiUrl+"/single",
		url.Values{
			"access_token": {n.AccessToken},
			"email":        {email},
		})

	if err != nil {
		log.Panic(err)
	} else if response.StatusCode != 200 {
		log.Panic("Email verification returned status ", response.StatusCode)
	}

	var verifyEmailResponse VerifyEmailResponse

	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&verifyEmailResponse); err != nil {
		log.Panic(err)
	}

	if !verifyEmailResponse.Success {
		if strings.Contains(verifyEmailResponse.Msg, "Authentication failed") {
			log.Println("Email verification failed: ", verifyEmailResponse.Msg)
			n.GetAccessToken()
			return n.VerifyEmail(email)
		} else {
			return verifyEmailResponse
		}
	}

	return verifyEmailResponse
}

func VerifyEmail(email string) VerifyEmailResponse {
	return NeverBounce.VerifyEmail(email)
}

func GetAccessToken() string {
	return NeverBounce.GetAccessToken()
}

func Init(neverbounce *NeverBounceCli) {
	NeverBounce = neverbounce

	// In case the API url has not been provided set the default one.
	// Used for testing
	if NeverBounce.ApiUrl == "" {
		NeverBounce.SetApiUrl(DEFAULT_API_URL)
	}
}