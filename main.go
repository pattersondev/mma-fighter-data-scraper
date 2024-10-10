package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"golang.org/x/net/html"
)

type FighterStats struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	HeightAndWeight string `json:"height_and_weight"`
	Birthdate       string `json:"birthdate"`
	Team            string `json:"team"`
	Nickname        string `json:"nickname"`
	Stance          string `json:"stance"`
}

func main() {

	// Slice to store all fighter stats we scrape
	var fighters []FighterStats

	c := colly.NewCollector(
		colly.AllowedDomains("espn.com", "www.espn.com"),
	)

	// Define the regular expression to filter only fighter stats pages
	statsPageRegex := regexp.MustCompile(`/mma/fighter/stats/_/id/\d+/.*`)

	// Visit the fighter stats page
	c.OnResponse(func(r *colly.Response) {
		// Parse the HTML response using golang.org/x/net/html
		if statsPageRegex.MatchString(r.Request.URL.Path) {
			var stats FighterStats // Create a new FighterStats object
			doc, err := html.Parse(bytes.NewReader(r.Body))
			if err != nil {
				log.Fatalf("Error parsing HTML: %v", err)
			}
			parseFighterStats(doc, &stats)

			// Append the new stats object to the fighters slice
			fighters = append(fighters, stats)

			// Output JSON formatted data once parsing is done
			jsonData, err := json.MarshalIndent(stats, "", "  ")
			if err != nil {
				log.Fatalf("Error marshaling JSON: %v", err)
			}
			fmt.Println(string(jsonData))
		}
	})

	// Start the scraping process
	c.Visit("https://www.espn.com/mma/fighter/stats/_/id/3088812/kamaru-usman")

	// Start the collector and wait for the task to complete
	c.Wait()
}

// Helper function to recursively parse HTML nodes and fill the FighterStats struct
func parseFighterStats(n *html.Node, stats *FighterStats) {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			// Find the PlayerHeader__Name class to extract the fighter's name
			fmt.Println(attr.Val)
			if strings.Contains(attr.Val, "PlayerHeader__Main") {
				fmt.Println("CALOL")
				extractNameFromHeader(n, stats)
			}
		}
	}

	// Find the PlayerHeader__Bio_List to extract bio details
	if n.Type == html.ElementNode && n.Data == "ul" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, "PlayerHeader__Bio_List") {
				extractBioDetails(n, stats)
			}
		}
	}

	// Recursively process child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		parseFighterStats(c, stats)
	}
}

// Helper function to extract text from a node
func extractTextFromNode(n *html.Node) string {
	if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
		return strings.TrimSpace(n.FirstChild.Data)
	}
	return ""
}

// Helper function to extract name from the PlayerHeader__Name class
func extractNameFromHeader(n *html.Node, stats *FighterStats) {
	fmt.Println("CALLED")
	if n.Type == html.ElementNode && n.Data == "span" {
		fmt.Println(n)
		if stats.FirstName == "" {
			stats.FirstName = extractTextFromNode(n)
		} else if stats.LastName == "" {
			stats.LastName = extractTextFromNode(n)
		}
	}

	// Recursively process child nodes to find spans
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractNameFromHeader(c, stats)
	}
}

// Helper function to extract bio details like height, weight, birthdate, team, etc.
func extractBioDetails(n *html.Node, stats *FighterStats) {
	if n.Type == html.ElementNode && n.Data == "li" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "div" {
				switch c.FirstChild.Data {
				case "HT/WT":
					// Extract height and weight
					stats.HeightAndWeight = extractHeightWeight(c.NextSibling)
				case "Birthdate":
					// Extract birthdate
					stats.Birthdate = extractTextFromNestedDiv(c.NextSibling)
				case "Team":
					// Extract team
					stats.Team = extractTextFromNestedDiv(c.NextSibling)
				case "Nickname":
					// Extract nickname
					stats.Nickname = extractTextFromNestedDiv(c.NextSibling)
				case "Stance":
					// Extract stance
					stats.Stance = extractTextFromNestedDiv(c.NextSibling)
				}
			}
		}
	}

	// Recursively process child nodes for bio details
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractBioDetails(c, stats)
	}
}

// Extract height and weight as a single string
func extractHeightWeight(n *html.Node) string {
	if n != nil && n.FirstChild != nil && n.FirstChild.Type == html.ElementNode && n.FirstChild.Data == "div" {
		return strings.TrimSpace(n.FirstChild.FirstChild.Data)
	}
	return ""
}

// Extract text from nested div
func extractTextFromNestedDiv(n *html.Node) string {
	if n != nil && n.FirstChild != nil && n.FirstChild.Type == html.ElementNode && n.FirstChild.Data == "div" {
		return strings.TrimSpace(n.FirstChild.FirstChild.Data)
	}
	return ""
}
