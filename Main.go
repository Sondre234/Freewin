package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Config struct {
	APIKey string `json:"apiKey"`
}

type Summoner struct {
	Puuid    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

func readConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	config, err := readConfig("config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		summonerName := "shao la int" // Replace with the actual summoner name
		region := "europe"            // Replace with the appropriate region code
		encodedSummonerName := url.QueryEscape(summonerName)
		apiURL := fmt.Sprintf("https://%s.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/euw?api_key=%s",
			region, encodedSummonerName, config.APIKey)

		resp, err := http.Get(apiURL)
		if err != nil {
			log.Fatalf("Error fetching summoner data: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			bodyString := string(bodyBytes)
			log.Fatalf("Error: %s", bodyString)
		}

		var summoner Summoner
		if err := json.NewDecoder(resp.Body).Decode(&summoner); err != nil {
			log.Fatalf("Error decoding response: %v", err)
		}

		tmpl := template.Must(template.ParseFiles("index.html"))
		if err := tmpl.Execute(w, summoner); err != nil {
			log.Fatal("Error executing template: ", err)
		}
	})

	log.Println("Starting server on :8000")
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", nil))
}
