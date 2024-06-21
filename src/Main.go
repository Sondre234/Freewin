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

func giveChamp(summonerInfo string) {

	getName := summonerInfo
	fmt.Println(getName)
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

func getAPIKEY() string {

	jsonFile, err := os.Open("C:\\Users\\sondr\\Stuff1\\resources\\config.json")
	if err != nil {
		log.Fatal("Error: ", err)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	// Unmarshal the JSON data into a Config struct
	var config Config
	if err := json.Unmarshal(byteValue, &config); err != nil {
		log.Fatalf("Failed to parse JSON file: %v", err)
	}

	apikey := config.APIKey
	return apikey
}

func main() {

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

	log.Println("Starting server on http://localhost:8000/")
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", nil))

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

		// Store PUUID in a variable
		puuid := summoner.Puuid

		// Fetch champion mastery data
		masteryURL := fmt.Sprintf("https://euw1.api.riotgames.com/lol/champion-mastery/v4/champion-masteries/by-puuid/%s?api_key=%s", puuid, config.APIKey)
		masteryResp, err := http.Get(masteryURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching champion mastery data: %v", err), http.StatusInternalServerError)
			return
		}
		defer masteryResp.Body.Close()

		if masteryResp.StatusCode != http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(masteryResp.Body)
			bodyString := string(bodyBytes)
			http.Error(w, bodyString, masteryResp.StatusCode)
			return
		}

		var champs []Champs
		if err := json.NewDecoder(masteryResp.Body).Decode(&champs); err != nil {
			http.Error(w, fmt.Sprintf("Error decoding champion mastery response: %v", err), http.StatusInternalServerError)
			return
		}

		// Fetch champion names
		champNames := make(map[int]string)
		file, err := os.Open("resources/champids.txt")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error opening champids.txt: %v", err), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.Split(line, "\t")
			if len(parts) == 2 {
				champID, err := strconv.Atoi(parts[0])
				if err != nil {
					continue
				}
				champNames[champID] = parts[1]
			}
		}
		if err := scanner.Err(); err != nil {
			http.Error(w, fmt.Sprintf("Error reading champids.txt: %v", err), http.StatusInternalServerError)
			return
		}

		// Prepare data with champion names
		type ChampInfo struct {
			Name           string
			ChampionLevel  int
			ChampionPoints int
			ChestGranted   bool
		}

		var champInfos []ChampInfo
		for _, champ := range champs {
			name := champNames[champ.ChampionId]
			champInfos = append(champInfos, ChampInfo{
				Name:           name,
				ChampionLevel:  champ.ChampionLevel,
				ChampionPoints: champ.ChampionPoints,
				ChestGranted:   champ.ChestGranted,
			})
		}

		// Create a response template with champion mastery information
		tmpl := template.Must(template.New("champion-mastery-info").Parse(`
            <h2>Champion Mastery Information</h2>
            {{range .}}
                <div class="champion">
                    <p><strong>Champion:</strong> {{.Name}}</p>
                    <p><strong>Level:</strong> {{.ChampionLevel}}</p>
                    <p><strong>Points:</strong> {{.ChampionPoints}}</p>
                    <p><strong>Chest Granted:</strong> {{.ChestGranted}}</p>
                </div>
            {{end}}
        `))
		if err := tmpl.Execute(w, champInfos); err != nil {
			http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		}
	}
}
