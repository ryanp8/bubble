package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
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

func errorResponse(err error) string {
	encoded, _ := json.Marshal(map[string]string{
		"error": err.Error(),
	})
	return string(encoded[:])
}

func getTokens(client *http.Client, body *AuthRequestBody) (*Tokens, error) {
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

func makeSpotifyRequest(app *pocketbase.PocketBase, method, url string, tokens *Tokens) (*map[string]interface{}, error) {
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
		tokenResponse, err := refreshToken(app, tokens)
		if err != nil {
			return nil, err
		}
		return makeSpotifyRequest(app, method, url, tokenResponse)
	}
	decoded := make(map[string]interface{})
	json.NewDecoder(res.Body).Decode(&decoded)
	return &decoded, nil
}

func getRoomOwner(app *pocketbase.PocketBase, roomId string) (*models.Record, error) {
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

func getRoomOwnerTokens(app *pocketbase.PocketBase, roomId string) (*Tokens, error) {
	roomOwner, err := getRoomOwner(app, roomId)
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

func refreshToken(app *pocketbase.PocketBase, currentTokens *Tokens) (*Tokens, error) {
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

func main() {
	app := pocketbase.New()
	httpClient := new(http.Client)

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		// Index root
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/",
			Handler: func(c echo.Context) error {
				return c.String(200, "hello")
			},
		})

		// Login to Spotify using authorization code generated by the client
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/login",
			Handler: func(c echo.Context) error {
				body := new(AuthRequestBody)
				if err := c.Bind(body); err != nil {
					log.Println(err)
					return c.String(400, errorResponse(err))
				}

				// Get access token from spotify using PKCE
				tokenResponse, err := getTokens(httpClient, body)
				if err != nil {
					log.Println(err)
					return c.String(400, errorResponse(err))
				}

				// Use access token to get user info
				userInfo, err := makeSpotifyRequest(app, "GET", "https://api.spotify.com/v1/me", tokenResponse)
				if err != nil {
					log.Println(err)
					return c.String(400, errorResponse(err))
				}

				userId := (*userInfo)["id"].(string)
				username := (*userInfo)["display_name"].(string)
				userEmail := (*userInfo)["email"].(string)

				userRecord, err := app.Dao().FindRecordById("users", userId)

				fmt.Println(userRecord)
				if err != nil {
					userCollection, err := app.Dao().FindCollectionByNameOrId("users")
					if err != nil {
						log.Println(err)
						return c.String(400, errorResponse(err))
					}
					userRecord = models.NewRecord(userCollection)
					userRecord.Set("id", userId)
					userRecord.Set("username", username)
					userRecord.Set("email", userEmail)
				}

				userRecord.Set("accessToken", tokenResponse.AccessToken)
				userRecord.Set("refreshToken", tokenResponse.RefreshToken)

				if err := app.Dao().SaveRecord(userRecord); err != nil {
					log.Println(err)
					return c.String(400, errorResponse(err))
				}
				response := &map[string]string{
					"userId":       userId,
					"email":        userEmail,
					"display_name": username,
					"access_token": tokenResponse.AccessToken,
				}
				return c.JSON(200, response)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		// Make a new room
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/rooms/:id",
			Handler: func(c echo.Context) error {
				roomId := c.PathParam("id")
				body := new(struct {
					UserId string `json:"user_id"`
				})
				if err := c.Bind(body); err != nil {
					log.Println(err)
					return c.String(400, errorResponse(err))
				}

				_, err := app.Dao().FindRecordById("rooms", roomId)

				if err != nil {
					rooms, err := app.Dao().FindCollectionByNameOrId("rooms")
					if err != nil {
						log.Println(err)
						return c.String(400, err.Error())
					}
					roomRecord := models.NewRecord(rooms)
					roomRecord.Set("id", roomId)
					roomRecord.Set("owner", body.UserId)
					roomRecord.Set("users", []string{body.UserId})
					roomRecord.Set("songs", make([]interface{}, 0))
					if err := app.Dao().SaveRecord(roomRecord); err != nil {
						return err
					}
					return c.String(200, "created room")
				}

				return c.String(200, "room already exists")
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/rooms/:id",
			Handler: func(c echo.Context) error {

				roomId := c.PathParam("id")
				roomRecord, err := app.Dao().FindRecordById("rooms", roomId)
				if err != nil {
					log.Println(err)
					return c.String(404, errorResponse(err))
				}
				if errs := app.Dao().ExpandRecord(roomRecord, []string{"owner"}, nil); len(errs) > 0 {
					log.Printf("failed to expand: %v\n", errs)
					return c.String(400, errorResponse(err))
				}
				roomOwner := roomRecord.ExpandedOne("owner")
				return c.JSON(200, map[string]string{
					"owner": roomOwner.GetString("username"),
				})
			},
		})

		// Add song to room queue
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/rooms/:id/queue",
			Handler: func(c echo.Context) error {
				roomId := c.PathParam("id")

				requestBody := new(struct {
					SpotifyUri string `json:"spotify_uri"`
				})

				if err := c.Bind(requestBody); err != nil {
					log.Println(err.Error())
					return c.String(400, errorResponse(err))
				}

				tokens, err := getRoomOwnerTokens(app, roomId)
				if err != nil {
					return c.JSON(404, errorResponse(err))
				}

				url := "https://api.spotify.com/v1/me/player/queue?uri=" + requestBody.SpotifyUri

				decoded, err := makeSpotifyRequest(app, "POST", url, tokens)

				errorReceived, ok := (*decoded)["error"]
				if err != nil || (ok && errorReceived.(map[string]interface{})["status"].(float64) == 404) {
					return c.JSON(404, map[string]string{
						"message": "error occured",
					})
				}
				return c.JSON(200, decoded)
			},
		})

		// Search for song
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/rooms/:roomId/search",
			Handler: func(c echo.Context) error {
				roomId := c.PathParam("roomId")
				track := c.QueryParam("track")
				tokens, err := getRoomOwnerTokens(app, roomId)
				if err != nil {
					return c.JSON(404, errorResponse(err))
				}

				url := "https://api.spotify.com/v1/search?q=" + strings.Replace(track, " ", "+", -1) + "&type=track"

				decoded, err := makeSpotifyRequest(app, "GET", url, tokens)
				if err != nil || ((*decoded)["status"] != nil && (*decoded)["status"] != 200) {
					return c.JSON(404, map[string]string{
						"message": "error occured",
					})
				}
				return c.JSON(200, decoded)
			},
		})

		// Remove room
		e.Router.AddRoute(echo.Route{
			Method: http.MethodDelete,
			Path:   "/rooms/:roomId",
			Handler: func(c echo.Context) error {
				roomId := c.PathParam("roomId")
				fmt.Println("CLOSING ROOM")
				roomRecord, err := app.Dao().FindRecordById("rooms", roomId)
				if err != nil {
					return c.JSON(400, errorResponse(err))
				}

				if err := app.Dao().DeleteRecord(roomRecord); err != nil {
					return c.JSON(400, errorResponse(err))
				}
				return c.JSON(200, map[string]string{
					"message": "removed room",
				})
			},
		})

		// Get owner's top songs
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/rooms/:roomId/top",
			Handler: func(c echo.Context) error {
				roomId := c.PathParam("roomId")
				tokens, _ := getRoomOwnerTokens(app, roomId)

				url := "https://api.spotify.com/v1/me/top/tracks?time_range=short_term"
				response, err := makeSpotifyRequest(app, "GET", url, tokens)
				if err != nil {
					return c.String(400, "error")
				}

				return c.JSON(200, response)
			},
		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
