package tool

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// WebFetch retrieves the main text content from a given URL.
// It uses goquery to parse the HTML and extract text, removing script and style tags.
func WebFetch(urlString string) (string, error) {
	client := http.Client{
		Timeout: 20 * time.Second,
	}

	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for %s: %w", urlString, err)
	}
	// Set a realistic User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL %s: %w", urlString, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("request to %s failed with status code %d", urlString, resp.StatusCode)
	}

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML from %s: %w", urlString, err)
	}

	// Remove script and style elements
	doc.Find("script, style").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	// Get the text from the body
	bodyText := doc.Find("body").Text()

	if bodyText == "" {
		return "", fmt.Errorf("no text content found in the body of %s", urlString)
	}

	// Clean up whitespace
	// return strings.Join(strings.Fields(bodyText), " ")
	return bodyText, nil
}
