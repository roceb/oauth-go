package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

var (
	twitchOauthConfig = oauth2.Config{
		ClientID:     os.Getenv("TWITCHKEY"),
		ClientSecret: os.Getenv("TWITCHSECRET"),
		Endpoint:     twitch.Endpoint,
		RedirectURL:  "http://localhost:8000/callback",
		Scopes:       []string{"channel:manage:broadcast", "moderation:read", "channel:moderate", "chat:edit"},
	}
	googleOauthConfig = oauth2.Config{
		ClientID:     os.Getenv("GKEY"),
		ClientSecret: os.Getenv("GSECRET"),
		Endpoint:     twitch.Endpoint,
		RedirectURL:  "http://localhost:8000/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/youtube", "https://www.googleapis.com/auth/tasks", "https://www.googleapis.com/auth/calendar", ""},
	}
	randomState = "Monkey Dragon Kiwi Japan Corn bath Jupiter"
	cmd         string
	name        string
	rs          = base64.StdEncoding.EncodeToString([]byte(randomState))
)

type TokenJson struct {
	Name         string    `json:"name"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
}
type Config map[string]*TokenJson

func main() {

	cmdFun := os.Args[1:]
	cmd = cmdFun[0]
	name = cmdFun[1]
	if cmd == "token" {
		getToken(name)
		os.Exit(0)
	}
	if cmd == "getid" {
		getId(name)
		os.Exit(0)
	}
	if cmd == "add" {
		http.HandleFunc("/", handleHome)
		http.HandleFunc("/login", handleLogIn)
		http.HandleFunc("/callback", handleCallback)
		http.ListenAndServe(":8000", nil)
	}
}

func getToken(name string) {
	data, err := ioutil.ReadFile(ConfigFilePath())
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	var jsonData TokenJson

	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		fmt.Print("Error unmarshal json", err.Error())
	}
	token := jsonData.AccessToken
	fmt.Println(token)
	return
}

func getId(name string) {
	fmt.Println(os.Getenv("TWITCHKEY"))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	var html = `<html><body><a href="/login">Log In </a></body></html>`
	fmt.Fprint(w, html)
}
func handleLogIn(w http.ResponseWriter, r *http.Request) {
	url := twitchOauthConfig.AuthCodeURL(rs)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
func handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != rs {
		fmt.Println("state is not valid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	token, err := twitchOauthConfig.Exchange(oauth2.NoContext, r.FormValue("code"))
	if err != nil {
		fmt.Println("could not get token: \n", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return

	}
	var rawData TokenJson
	rawData = joinJson(rawData, name, token)
	file, _ := json.MarshalIndent(rawData, "", " ")

	_ = ioutil.WriteFile(ConfigFilePath(), file, 0644)
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}
func joinJson(t TokenJson, n string, o *oauth2.Token) TokenJson {
	t.AccessToken = o.AccessToken
	t.Expiry = o.Expiry
	t.RefreshToken = o.RefreshToken
	t.TokenType = o.TokenType
	t.Name = n
	return t
}
func ConfigFilePath() string {
	file := os.Getenv("AUTHCONF")
	if file == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(dir, "auth")
		err = os.Mkdir(dir, 0700)
		if err != nil {
			return ""
		}
		file = filepath.Join(dir, "config.json")
	}
	return file
}
func (c Config) Store() error { return c.Save(ConfigFilePath()) }

// Open loads the JSON data from the ConfigFilePath path.
func (c *Config) Open() error { return c.Load(ConfigFilePath()) }

func (c Config) Save(path string) error {
	return ioutil.WriteFile(path, []byte(c.String()), 0600)
}
func (c *Config) Load(path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return c.Parse(buf)
}
func (c Config) String() string {
	byt, _ := json.MarshalIndent(c, "", "  ")
	return string(byt)
}
func (c *Config) Parse(buf []byte) error { return json.Unmarshal(buf, c) }
