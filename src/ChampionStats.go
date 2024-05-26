package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func getRecommendedChampion(filePath string) ([]Champion, error) {

	jsonFile, err := os.Open(filePath)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("bestChamps.json åpnet")

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var stats Stats

	json.Unmarshal(byteValue, &stats)

	for i := 0; i < len(stats.Stats); i++ {

		fmt.Println("Champion: " + stats.Stats[i].Champion)
		fmt.Println("Role: " + stats.Stats[i].Role)
		fmt.Println("Winrate: " + stats.Stats[i].Winrate)
	}

	return stats.Stats, nil
}
