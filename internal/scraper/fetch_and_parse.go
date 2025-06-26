package scraper

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func FetchAndParsePage(url string) (*goquery.Document, error) {
	client := &http.Client{}

	// Fetch the page
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	// Extract and decompress body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle gzip compression if present
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer reader.Close()

		body, err = io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress gzip content: %w", err)
		}
	}

	// Convert from Windows-1251 to UTF-8
	bodyStr := string(body)
	if strings.Contains(bodyStr, "windows-1251") || strings.Contains(bodyStr, "charset=windows-1251") {
		// Convert from Windows-1251 to UTF-8
		decoder := charmap.Windows1251.NewDecoder()
		utf8Body, _, err := transform.Bytes(decoder, body)
		if err != nil {
			fmt.Printf("Warning: failed to convert encoding: %v\n", err)
		} else {
			body = utf8Body
		}
	}

	// Clean any invalid UTF-8 sequences
	bodyStr = string(body)
	if !utf8.ValidString(bodyStr) {
		bodyStr = strings.ToValidUTF8(bodyStr, "")
		body = []byte(bodyStr)
	}

	// Parse HTML
	return goquery.NewDocumentFromReader(bytes.NewReader(body))
}
