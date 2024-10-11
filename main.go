package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/gocolly/colly"
	"golang.org/x/net/html"
)

type StrikingStats struct {
	Date        string `json:"date"`
	Opponent    string `json:"opponent"`
	Event       string `json:"event"`
	Result      string `json:"result"`
	SDblA       string `json:"sdbl_a"`  // Significant Distance Blows Landed/Attempted
	SDhlA       string `json:"sdhl_a"`  // Significant Head Blows Landed/Attempted
	SDllA       string `json:"sdll_a"`  // Significant Leg Blows Landed/Attempted
	TSL         string `json:"tsl"`     // Total Strikes Landed
	TSA         string `json:"tsa"`     // Total Strikes Attempted
	SSL         string `json:"ssl"`     // Significant Strikes Landed
	SSA         string `json:"ssa"`     // Significant Strikes Attempted
	TSL_TSA     string `json:"tsl_tsa"` // Total Strikes Landed/Attempted
	KD          string `json:"kd"`      // Knockdowns
	PercentBody string `json:"percent_body"`
	PercentHead string `json:"percent_head"`
	PercentLeg  string `json:"percent_leg"`
}

type ClinchStats struct {
	Date     string `json:"date"`
	Opponent string `json:"opponent"`
	Event    string `json:"event"`
	Result   string `json:"result"`
	SCBL     string `json:"scbl"`   // Significant Distance Blows Landed/Attempted
	SCBA     string `json:"scba"`   // Significant Head Blows Landed/Attempted
	SCHL     string `json:"schl"`   // Significant Leg Blows Landed/Attempted
	SCHA     string `json:"scha"`   // Significant Strikes Landed
	SCLL     string `json:"scll"`   // Significant Strikes Attempted
	SCLA     string `json:"scla"`   // Significant Strikes Attempted
	RV       string `json:"rv"`     // Reversal Volumes
	SR       string `json:"sr"`     // Reversal Volumes
	TDL      string `json:"tdl"`    // takedowns landed
	TDA      string `json:"td"`     // takedowns attempted
	TDS      string `json:"tds"`    // Takedown slams
	TK_ACC   string `json:"tk_acc"` // Takedown Accuracy
}

type FighterStats struct {
	FirstName       string          `json:"first_name"`
	LastName        string          `json:"last_name"`
	HeightAndWeight string          `json:"height_and_weight"`
	Birthdate       string          `json:"birthdate"`
	Team            string          `json:"team"`
	Nickname        string          `json:"nickname"`
	Stance          string          `json:"stance"`
	WinLossRecord   string          `json:"win_loss_record"`
	TKORecord       string          `json:"tko_record"`
	SubRecord       string          `json:"sub_record"`
	StrikingStats   []StrikingStats `json:"striking_stats"` // Array of striking stats
	ClinchStats     []ClinchStats   `json:"clinch_stats"`   // Array of clinch stats
}

func shouldVisitURL(url string) bool {
	return (strings.Contains(url, "espn.com/mma/fightcenter") ||
		strings.Contains(url, "espn.com/mma/fighter/")) &&
		!strings.Contains(url, "news") && !strings.Contains(url, "history") && !strings.Contains(url, "bio") && !strings.Contains(url, "watch") && !strings.Contains(url, "schedule")
}

func main() {

	// Slice to store all fighter stats we scrape
	var fighters []FighterStats

	c := colly.NewCollector(
		colly.AllowedDomains("espn.com", "www.espn.com"),
	)

	// Visit the fighter stats page
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href")) // Convert to absolute URL
		// Use shouldVisitURL to determine if the link should be visited
		if shouldVisitURL(link) && strings.Contains(link, "stats") {
			// Visit the link if it matches
			e.Request.Visit(link)
		}
	})

	// Process the response for pages that match the regex
	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("Processing URL: %s\n", r.Request.URL.String()) // Debug: Print each processed URL
		if shouldVisitURL(r.Request.URL.String()) {
			var stats FighterStats // Create a new FighterStats object
			doc, err := html.Parse(bytes.NewReader(r.Body))
			if err != nil {
				log.Fatalf("Error parsing HTML: %v", err)
			}
			parseFighterStats(doc, &stats)

			// Check for the presence of the striking stats table
			if hasStrikingStatsTable(doc) {
				parseStrikingStats(doc, &stats)
			}

			if hasClinchStatsTable(doc) {
				parseClinchStats(doc, &stats)
			}

			// Check if both first and last names are not empty
			if stats.FirstName != "" && stats.LastName != "" {
				fighters = append(fighters, stats)
				fmt.Println("Fighter Added", stats.FirstName, stats.LastName)
			}
		}
	})

	// Start the scraping process at the fight center
	c.Visit("https://www.espn.com/mma/fighter/stats/_/id/3088812/kamaru-usman")

	// Start the collector and wait for the task to complete
	c.Wait()

	// Marshal the fighters slice into JSON
	jsonData, err := json.MarshalIndent(fighters, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	// Write the JSON data to a file
	err = ioutil.WriteFile("fighters.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON to file: %v", err)
	}

	fmt.Println("Data successfully written to fighters.json")
}

// Helper function to recursively parse HTML nodes and fill the FighterStats struct
func parseFighterStats(n *html.Node, stats *FighterStats) {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			// Find the PlayerHeader__Name class to extract the fighter's name
			if strings.Contains(attr.Val, "PlayerHeader__Main") {
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

	// Find the PlayerHeader__Right class to extract the win-loss, (T)KO, and SUB records
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, "PlayerHeader__Right") {
				extractWinLossRecord(n, stats)
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
	if n != nil && n.FirstChild != nil {
		return strings.TrimSpace(n.FirstChild.Data)
	}
	return ""
}

// Helper function to extract name from the PlayerHeader__Name class
func extractNameFromHeader(n *html.Node, stats *FighterStats) {
	if n.Type == html.ElementNode && n.Data == "span" {
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

// Updated helper function to extract win-loss, (T)KO, and SUB records
func extractWinLossRecord(n *html.Node, stats *FighterStats) {
	if n.Type == html.ElementNode && n.Data == "div" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "div" {
				for _, attr := range c.Attr {
					if attr.Key == "aria-label" {
						switch attr.Val {
						case "Wins-Losses-Draws":
							stats.WinLossRecord = extractTextFromNode(c.NextSibling)
						case "Technical Knockout-Technical Knockout Losses":
							stats.TKORecord = extractTextFromNode(c.NextSibling)
						case "Submissions-Submission Losses":
							stats.SubRecord = extractTextFromNode(c.NextSibling)
						}
					}
				}
			}
		}
	}

	// Recursively process child nodes for win-loss, (T)KO, and SUB records
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractWinLossRecord(c, stats)
	}
}

func parseStrikingStats(n *html.Node, fighter *FighterStats) {
	if n.Type == html.ElementNode && n.Data == "tbody" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "tr" {
				var stats StrikingStats
				extractStrikingStatsFromRow(c, &stats)
				fighter.StrikingStats = append(fighter.StrikingStats, stats)
			}
		}
	}

	// Recursively process child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		parseStrikingStats(c, fighter)
	}
}

func parseClinchStats(n *html.Node, fighter *FighterStats) {
	// Check if the current node is a table header containing "Clinch"
	// isClinchTable := false
	// if n.Type == html.ElementNode && n.Data == "div" && n.FirstChild != nil && strings.Contains(n.FirstChild.Data, "Clinch") {
	// 	isClinchTable = true
	// }

	// If we are in the clinch table, look for tbody
	if n.Type == html.ElementNode && n.Data == "tbody" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "tr" {
				var stats ClinchStats
				extractClinchStatsFromRow(c, &stats)
				fighter.ClinchStats = append(fighter.ClinchStats, stats)
			}
		}
	}

	// Recursively process child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		parseClinchStats(c, fighter)
	}
}

func extractClinchStatsFromRow(n *html.Node, stats *ClinchStats) {
	tdIndex := 0

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "td" {
			text := extractTextFromNode(c)
			switch tdIndex {
			case 0:
				stats.Date = text
			case 1:
				stats.Opponent = extractTextFromNode(c.FirstChild)
			case 2:
				stats.Event = extractTextFromNode(c.FirstChild)
			case 3:
				stats.Result = extractTextFromNode(c.FirstChild)
			case 4:
				stats.SCBL = text
			case 5:
				stats.SCBA = text
			case 6:
				stats.SCHA = text
			case 7:
				stats.SCLL = text
			case 8:
				stats.SCLA = text
			case 9:
				stats.RV = text
			case 10:
				stats.SR = text
			case 11:
				stats.TDL = text
			case 12:
				stats.TDA = text
			case 13:
				stats.TDS = text
			case 14:
				stats.TK_ACC = text
			}
			tdIndex++
		}
	}

}

// Extract table stats from row.
func extractStrikingStatsFromRow(n *html.Node, stats *StrikingStats) {
	tdIndex := 0
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "td" {
			text := extractTextFromNode(c)
			switch tdIndex {
			case 0:
				stats.Date = text
			case 1:
				stats.Opponent = extractTextFromNode(c.FirstChild)
			case 2:
				stats.Event = extractTextFromNode(c.FirstChild)
			case 3:
				stats.Result = extractTextFromNode(c.FirstChild)
			case 4:
				stats.SDblA = text
			case 5:
				stats.SDhlA = text
			case 6:
				stats.SDllA = text
			case 7:
				stats.TSL = text
			case 8:
				stats.TSA = text
			case 9:
				stats.SSL = text
			case 10:
				stats.SSA = text
			case 11:
				stats.TSL_TSA = text
			case 12:
				stats.KD = text
			case 13:
				stats.PercentBody = text
			case 14:
				stats.PercentHead = text
			case 15:
				stats.PercentLeg = text
			}
			tdIndex++
		}
	}
}

// Helper function to check if the striking stats table is present
func hasStrikingStatsTable(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && attr.Val == "Table__Title" {
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode && n.FirstChild.Data == "striking" {
					return true
				}
			}
		}
	}

	// if n.Type == html.ElementNode && n.Data == "tbody" {
	// 	for c := n.FirstChild; c != nil; c = c.NextSibling {
	// 		if c.Type == html.ElementNode && c.Data == "tr" {
	// 			return true
	// 		}
	// 	}
	// }
	// Check for the presence of the specific <div> element
	// Recursively check child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasStrikingStatsTable(c) {
			return true
		}
	}
	return false
}

func hasClinchStatsTable(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && attr.Val == "Table__Title" {
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode && n.FirstChild.Data == "Clinch" {
					return true
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasStrikingStatsTable(c) {
			return true
		}
	}
	return false
}
