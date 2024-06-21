package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func updateWinrateDataFromTextFileToJsonFormat() {

	// Fix file path to not rely on specific system
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

/*func main1() {

	apikey := getAPIKEY()

	puuid := "vosf4Gq_pWSOBD-7jqnqupV1ZCNvbS66k10cDcIVJjjkiI6rjl03_-OK5acnbULt3ng3xRDGvZeYNA"
	apiReq := fmt.Sprintf("https://euw1.api.riotgames.com/lol/champion-mastery/v4/champion-masteries/by-puuid/%s?api_key=%s", puuid, apikey)

	res, err := http.Get(apiReq)
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

		log.Fatalf("unexpected status code: %d %d", res.StatusCode,
			"\n400\tBad request\n401 Unauthorized\n403 Forbidden\n404 Data not found\n405 Method not allowed\n415 Unsupported media type\n429Rate limit exceeded\n500Internal server error\n502\tBad gateway\n503Service unavailable\n504Gateway timeout")
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
}*/
