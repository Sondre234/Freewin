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

func searchHandler(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		summonerName := r.FormValue("summonerName")
		tagLine := r.FormValue("tagLine")

		if summonerName == "" || tagLine == "" {
			http.Error(w, "Missing summonerName or tagLine", http.StatusBadRequest)
			return
		}

		encodedSummonerName := url.QueryEscape(summonerName)
		encodedTagLine := url.QueryEscape(tagLine)
		apiURL := fmt.Sprintf("https://europe.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s?api_key=%s", encodedSummonerName, encodedTagLine, config.APIKey)

		resp, err := http.Get(apiURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching summoner data: %v", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			bodyString := string(bodyBytes)
			http.Error(w, bodyString, resp.StatusCode)
			return
		}

		var summoner Summoner
		if err := json.NewDecoder(resp.Body).Decode(&summoner); err != nil {
			http.Error(w, fmt.Sprintf("Error decoding response: %v", err), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.New("summoner-info").Parse(`
			<p class="card-text"><strong>Puuid:</strong> {{.Puuid}}</p>
			<p class="card-text"><strong>Game Name:</strong> {{.GameName}}</p>
			<p class="card-text"><strong>Tag Line:</strong> {{.TagLine}}</p>
		`))
		if err := tmpl.Execute(w, summoner); err != nil {
			http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		}
	}
}

func main() {
	config, err := readConfig("config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	http.HandleFunc("/search", searchHandler(config))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("index.html"))
		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		}
	})

	log.Println("Starting server on :8000")
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", nil))
}
