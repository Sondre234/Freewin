package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main1() {

	config, err := readConfig("resources/config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Serve static files from the "static" directory
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/search", searchHandler(config))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("htmx/index.html"))
		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		}
	})

	log.Println("Starting server on :8000")
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", nil))

}

type Stats struct {
	Stats []Champion `json:"stats"`
}

type Champs struct {
	ChampionId     int  `json:"championId"`
	ChampionLevel  int  `json:"championLevel"`
	ChampionPoints int  `json:"championPoints"`
	ChestGranted   bool `json:"chestGranted"`
}

type Champion struct {
	Champion string `json:"champion"`
	Role     string `json:"role"`
	Winrate  string `json:"winrate"`
}

type Config struct {
	APIKey string `json:"apiKey"`
}

type Summoner struct {
	Puuid    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

func getChampionNameByID(id int) (string, error) {
	file, err := os.Open("resources/champids.txt")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) == 2 {
			champID, err := strconv.Atoi(parts[0])
			if err != nil {
				return "", err
			}
			if champID == id {
				return parts[1], nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "Champion not found", nil
}

func main() {
	// Get PUUID and champion info
	name, err := getChampionNameByID(266)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Champion Name:", name)
	}

	res, err := http.Get("https://euw1.api.riotgames.com/lol/champion-mastery/v4/" +
		"champion-masteries/by-puuid/vosf4Gq_pWSOBD-7jqnqupV1ZCNvbS66k10cDcIVJjjkiI6rjl03_-OK5acnbULt3ng3xRDGvZeYNA?api_key=RGAPI-277d267d-c27e-4411-a87a-df4502edc571")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var champs []Champs
	json.Unmarshal(body, &champs)

	if res.StatusCode != http.StatusOK {
		log.Fatalf("unexpected status code: %d", res.StatusCode)
	}

	for i, p := range champs {
		championName, err := getChampionNameByID(p.ChampionId)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		fmt.Println("Champion", (i + 1), ":", championName)
		fmt.Println("ChampionLevel: ", p.ChampionLevel)
		fmt.Println("ChampionPoints: ", p.ChampionPoints)
		fmt.Println("ChestGranted: ", p.ChestGranted)
	}
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

		summonerInfo := r.FormValue("summonerInfo")
		parts := strings.Split(summonerInfo, "#")
		if len(parts) != 2 {
			http.Error(w, "Invalid format. Use summonerName#tagLine", http.StatusBadRequest)
			return
		}
		summonerName := parts[0]
		tagLine := parts[1]

		encodedSummonerName := url.QueryEscape(summonerName)
		encodedTagLine := url.QueryEscape(tagLine)
		apiURL := fmt.Sprintf("https://europe.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s?api_key=%s",
			encodedSummonerName, encodedTagLine, config.APIKey)

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

func updateWinrateDataFromTextFileToJsonFormat() {

	winratestxtpath := "C:\\Users\\sondr\\Stuff1\\resources\\winrates.txt"
	bestChampsjsonPath := "C:\\Users\\sondr\\Stuff1\\resources\\bestChamps.json"

	file, err := os.Open(winratestxtpath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var champions []Champion
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ": ")
		if len(parts) != 2 {
			fmt.Println("Invalid line format:", line)
			continue
		}

		roleChamp := parts[0]
		winrate := strings.TrimSuffix(parts[1], "%")

		roleChampParts := strings.SplitN(roleChamp, " ", 2)
		if len(roleChampParts) != 2 {
			fmt.Println("Invalid role and champion format:", roleChamp)
			continue
		}

		role := strings.ToLower(roleChampParts[0])
		champion := strings.ToLower(roleChampParts[1])

		champions = append(champions, Champion{
			Champion: champion,
			Role:     role,
			Winrate:  winrate,
		})
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	stats := Stats{Stats: champions}
	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	err = os.WriteFile(bestChampsjsonPath, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON to file:", err)
		return
	}

	fmt.Println("JSON data successfully written to resources/bestChamps.json")
}
