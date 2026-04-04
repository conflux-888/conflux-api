package sync

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	gdeltLastUpdateURL = "http://data.gdeltproject.org/gdeltv2/lastupdate.txt"
	gdeltTotalFields   = 61
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// FetchLatestEvents downloads the latest GDELT export CSV and returns conflict events
func (c *Client) FetchLatestEvents(ctx context.Context) ([]GDELTEvent, string, error) {
	// 1. Get latest update URL
	csvURL, err := c.getLatestExportURL(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get latest export URL: %w", err)
	}

	log.Info().Str("url", csvURL).Msg("[sync.Client] downloading GDELT export")

	// 2. Download ZIP
	zipData, err := c.download(ctx, csvURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download export: %w", err)
	}

	// 3. Unzip and parse
	events, err := parseExportCSV(zipData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse export CSV: %w", err)
	}

	// 4. Filter conflict events (QuadClass == 4)
	var conflicts []GDELTEvent
	var maxDateAdded string
	for _, e := range events {
		if e.QuadClass != "4" {
			continue
		}
		if e.EventRootCode < "18" {
			continue
		}
		conflicts = append(conflicts, e)
		if e.DateAdded > maxDateAdded {
			maxDateAdded = e.DateAdded
		}
	}

	log.Info().Int("total", len(events)).Int("conflicts", len(conflicts)).Msg("[sync.Client] parsed GDELT events")

	return conflicts, maxDateAdded, nil
}

func (c *Client) getLatestExportURL(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gdeltLastUpdateURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse lastupdate.txt — each line: <size> <md5> <url>
	// First line is the .export.CSV.zip
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 && strings.Contains(fields[2], ".export.CSV") {
			return fields[2], nil
		}
	}

	return "", fmt.Errorf("export CSV URL not found in lastupdate.txt")
}

func (c *Client) download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func parseExportCSV(zipData []byte) ([]GDELTEvent, error) {
	r, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}

	if len(r.File) == 0 {
		return nil, fmt.Errorf("empty zip file")
	}

	f, err := r.File[0].Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV in zip: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	events := make([]GDELTEvent, 0, len(lines))

	for _, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) < gdeltTotalFields {
			continue
		}

		goldstein, _ := strconv.ParseFloat(fields[30], 64)
		numMentions, _ := strconv.Atoi(fields[31])
		numSources, _ := strconv.Atoi(fields[32])
		numArticles, _ := strconv.Atoi(fields[33])
		avgTone, _ := strconv.ParseFloat(fields[34], 64)
		lat, _ := strconv.ParseFloat(fields[56], 64)
		lng, _ := strconv.ParseFloat(fields[57], 64)

		e := GDELTEvent{
			GlobalEventID:        fields[0],
			Day:                  fields[1],
			Actor1Name:           fields[6],
			Actor2Name:           fields[16],
			IsRootEvent:          fields[25],
			EventCode:            fields[26],
			EventBaseCode:        fields[27],
			EventRootCode:        fields[28],
			QuadClass:            fields[29],
			GoldsteinScale:       goldstein, // fields[30]
			NumMentions:          numMentions,
			NumSources:           numSources,
			NumArticles:          numArticles,
			AvgTone:              avgTone,
			ActionGeoType:        fields[51],
			ActionGeoFullName:    fields[52],
			ActionGeoCountryCode: fields[53],
			ActionGeoLat:         lat,
			ActionGeoLong:        lng,
			DateAdded:            fields[59],
			SourceURL:            fields[60],
		}
		events = append(events, e)
	}

	return events, nil
}
