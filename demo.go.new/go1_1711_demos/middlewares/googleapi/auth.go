package googleapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

//
// Gsheet Auth
// https://developers.google.com/sheets/api/quickstart/go
//
// 1. Open Google Cloud Console
// 2. Create a Google Cloud Platform project
// 3. Enable Google Sheets Api
// 4. Create a OAuth Client Id (type:Desktop), and downlaod to credentials.json
//

func getAuthClient(authScope string) *http.Client {
	b, err := ioutil.ReadFile(filepath.Join(os.Getenv("PRJ_PATH"), "internal", "credentials.json"))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, authScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	return getClient(config)

}

func getClient(config *oauth2.Config) *http.Client {
	prjPath, ok := os.LookupEnv("PRJ_PATH")
	if !ok {
		log.Fatalln("env [PRJ_PATH] is not set")
	}

	tokFile := filepath.Join(prjPath, "internal", "token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	// http://localhost/?state=state-token&code={AuthCode}&scope=https://www.googleapis.com/auth/drive.metadata.readonly
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func saveToken(path string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}

	if err := json.NewEncoder(f).Encode(token); err != nil {
		log.Fatalf("Json encode error: %v", err)
	}
}
