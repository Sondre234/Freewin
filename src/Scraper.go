package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
)

func RENAMETOMAIN() {

	c := colly.NewCollector()

	c.OnHTML("", func(h *colly.HTMLElement) {

	})
	//REMEMBER TO RENAME TO MAIN WHEN TESTING
	c.OnHTML("strong.champion-name", func(h *colly.HTMLElement) {
		fmt.Println(h.ChildAttr("h3 a", "title"))
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting", r.URL)
	})

	c.Visit("https://u.gg/lol/tier-list")
}

/*
func main() {

	destinationFile := "winratesScraped.txt"
	// Create a new collector
	c := colly.NewCollector()

	// Open file to write the results
	file, err := os.Create(destinationFile)
	if err != nil {
		log.Fatalf("Could not create file: %v", err)
	}
	defer file.Close()

	// Write header to file
	_, err = file.WriteString("Champion,Role,WinRate\n")
	if err != nil {
		log.Fatalf("Could not write to file: %v", err)
	}

	// Debugging information
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnHTML("table tbody tr", func(e *colly.HTMLElement) {
		champion := e.ChildText("td:nth-child(3) a")
		role := e.ChildAttr("td:nth-child(2) img", "alt")
		winrate := e.ChildText("td:nth-child(5)")

		fmt.Printf("Found row - Champion: %s, Role: %s, WinRate: %s\n", champion, role, winrate) // Debug line

		if champion != "" && role != "" && winrate != "" {
			line := fmt.Sprintf("%s,%s,%s\n", champion, role, winrate)
			_, err := file.WriteString(line)
			if err != nil {
				log.Printf("Could not write to file: %v", err)
			} else {
				fmt.Printf("Written: %s", line)
			}
		}
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	// Start scraping the page
	err = c.Visit("https://u.gg/lol/tier-list")
	if err != nil {
		log.Fatalf("Could not visit the website: %v", err)
	}

	fmt.Println("Scraping complete, check winratesScraped.txt for the results.")
}
*/
