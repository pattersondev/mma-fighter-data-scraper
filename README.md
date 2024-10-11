# MMA Fighter Data Scraper

This project is a web scraper built using Go and the Colly library. It scrapes MMA fighter statistics from ESPN's website and stores the data in a JSON file. The scraper collects various statistics, including striking, clinch, and ground stats for each fighter.

## Features

- Scrapes fighter statistics from ESPN's MMA section.
- Collects detailed stats such as striking, clinch, and ground performance.
- Stores the collected data in a structured JSON format.
- Utilizes concurrency to efficiently scrape multiple pages.

## Prerequisites

- Go 1.16 or later
- Internet connection

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/mma-fighter-data-scraper.git
   cd mma-fighter-data-scraper
   ```

2. Install the dependencies:

   ```bash
   go mod tidy
   ```

## Usage

1. Run the scraper:

   ```bash
   go run main.go
   ```

2. The scraper will visit ESPN's MMA fight center and collect data on fighters. The data will be saved to a file named `fighters.json` in the project directory.

## Code Structure

- `main.go`: The main file containing the scraper logic.
- `FighterStats`: Struct to hold fighter's personal and performance data.
- `StrikingStats`, `ClinchStats`, `GroundStats`: Structs to hold specific types of performance data.
- Helper functions to parse HTML and extract relevant data.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Colly](https://github.com/gocolly/colly) - Elegant Scraping Framework for Gophers
- [ESPN](https://www.espn.com) - Source of MMA fighter data
