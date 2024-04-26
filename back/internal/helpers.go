package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

var CLIENT_ID = os.Getenv("CLIENT_ID")
var CLIENT_SECRET = os.Getenv("CLIENT_SECRET")
var httpClient = &http.Client{}

type AuthRequestBody struct {
	ClientId    string `json:"client_id"`
	GrantType   string `json:"grant_type"`
	Code        string `json:"code"`
	State       string `json:"state"`
	RedirectUri string `json:"redirect_uri"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type User struct {
	Id           string `db:"id" json:"id"`
	Username     string `db:"username" json:"username"`
	Email        string `db:"email" json:"email"`
	AccessToken  string `db:"accessToken" json:"access_token"`
	RefreshToken string `db:"refreshToken" json:"refresh_token"`
	Room         string `db:"room" json:"room"`
}

type Room struct {
	Id    string   `db:"id" json:"id"`
	Users []string `db:"users" json:"users"`
	Owner string   `db:"owner" json:"owner"`
}

func ErrorResponse(err error) string {
	encoded, _ := json.Marshal(map[string]string{
		"error": err.Error(),
	})
	return string(encoded[:])
}

// Gets the authorization code from the client and makes a request to Spotify servers to receive access and refresh tokens.
// Returns Token object
func GetTokens(client *http.Client, body *AuthRequestBody) (*Tokens, error) {
	// postBody := []byte("client_id=" + CLIENT_ID + "&grant_type=authorization_code&code=" + body.Code + "&redirect_uri=" + body.RedirectUri + "&code_verifier=" + body.CodeVerifier)
	postBody := []byte("code=" + body.Code + "&redirect_uri=" + body.RedirectUri + "&grant_type=authorization_code")

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBuffer(postBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(CLIENT_ID+":"+CLIENT_SECRET)))
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	tokenResponse := new(Tokens)
	json.NewDecoder(res.Body).Decode(&tokenResponse)
	return tokenResponse, nil
}

// Makes a request to the Spotify api and returns the response. If the access token is expired, automatically
// refreshes the token and resends the original request
func MakeSpotifyRequest(app *pocketbase.PocketBase, method, url string, tokens *Tokens) (*map[string]interface{}, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+tokens.AccessToken)
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 401 || res.StatusCode == 400 {
		tokenResponse, err := RefreshToken(app, tokens)
		if err != nil {
			return nil, err
		}
		return MakeSpotifyRequest(app, method, url, tokenResponse)
	}
	decoded := make(map[string]interface{})
	json.NewDecoder(res.Body).Decode(&decoded)
	return &decoded, nil
}

// Helper function to query the room owner from the database
func GetRoomOwner(app *pocketbase.PocketBase, roomId string) (*models.Record, error) {
	roomRecord, err := app.Dao().FindRecordById("rooms", roomId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if errs := app.Dao().ExpandRecord(roomRecord, []string{"owner"}, nil); len(errs) > 0 {
		log.Printf("failed to expand: %v\n", errs)
		return nil, err
	}
	roomOwner := roomRecord.ExpandedOne("owner")
	return roomOwner, nil
}

// Helper function to query the room owner and get the access and refresh tokens for that room
func GetRoomOwnerTokens(app *pocketbase.PocketBase, roomId string) (*Tokens, error) {
	roomOwner, err := GetRoomOwner(app, roomId)
	if err != nil {
		return nil, err
	}
	acccesToken := roomOwner.GetString("accessToken")
	refreshToken := roomOwner.GetString("refreshToken")
	tokens := &Tokens{
		AccessToken:  acccesToken,
		RefreshToken: refreshToken,
	}
	return tokens, nil
}

// Helper function for refresh token flow
func RefreshToken(app *pocketbase.PocketBase, currentTokens *Tokens) (*Tokens, error) {
	// Match the access token to a user
	client := &http.Client{}
	userRecord, err := app.Dao().FindFirstRecordByData("users", "accessToken", currentTokens.AccessToken)
	if err != nil {
		return nil, err
	}
	refreshTokenBody := []byte("grant_type=refresh_token&refresh_token=" + currentTokens.RefreshToken + "&client_id=" + CLIENT_ID)
	refreshTokenReq, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBuffer(refreshTokenBody))
	refreshTokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	refreshTokenReq.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(CLIENT_ID+":"+CLIENT_SECRET)))
	if err != nil {
		return nil, err
	}
	res, err := client.Do(refreshTokenReq)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()
	fmt.Printf("refresh token status: %v\n", res.StatusCode)
	decoded := make(map[string]interface{})
	json.NewDecoder(res.Body).Decode(&decoded)

	newAccessToken := decoded["access_token"].(string)
	userRecord.Set("accessToken", newAccessToken)
	if err := app.Dao().SaveRecord(userRecord); err != nil {
		return nil, err
	}
	return &Tokens{
		AccessToken:  newAccessToken,
		RefreshToken: currentTokens.RefreshToken,
	}, nil
}
