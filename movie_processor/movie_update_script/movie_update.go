package movie_update_script

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

const (
	spreadsheetID = "18vG--tWYMSigQWStcLqKxe4BsgZFlWYOLBvALV2hxPk"
	actualSheet   = "Actual"
	resultSheet   = "Actual2"
	apiURL        = "https://api.kinopoisk.dev/v1.4/movie/search "
)

var apiKeys = []string{"1D9R88Y-GRK4755-P3RRYM4-AXYPHP0", "MTDZVBR-1S94CD0-MMN2CVQ-5GE4TW7", "DWQE76X-GZHM9VJ-M7BAW46-8ZB5FWF"}
var apiKeyIndex = 0

type SearchResult struct {
	Docs []struct {
		ID              int    `json:"id"`
		Name            string `json:"name"`
		AlternativeName string `json:"alternativeName"`
		Year            int    `json:"year"`
		IsSeries        bool   `json:"isSeries"`
		ExternalID      struct {
			KpHD string `json:"kpHD"`
			IMDB string `json:"imdb"`
		} `json:"externalId"`
		Rating struct {
			KP float64 `json:"kp"`
		} `json:"rating"`
		Votes struct {
			KP int `json:"kp"`
		} `json:"votes"`
	} `json:"docs"`
}

func getAPIKey() string {
	return apiKeys[apiKeyIndex]
}

func switchAPIKey() bool {
	if apiKeyIndex < len(apiKeys)-1 {
		apiKeyIndex++
		return true
	}
	return false
}

func main() {
	ctx := context.Background()
	b, err := google.FindDefaultCredentials(ctx, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Unable to get Google credentials: %v", err)
	}

	srv, err := sheets.NewService(ctx, option.WithCredentials(b))
	if err != nil {
		log.Fatalf("Unable to create Sheets client: %v", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, actualSheet).Do()
	if err != nil {
		log.Fatalf("Unable to read spreadsheet: %v", err)
	}

	headers := resp.Values[0]
	rows := resp.Values[1:]

	var result [][]interface{}
	result = append(result, headers)

	for _, row := range rows {
		if len(row) > 8 && row[8] != nil && row[8].(string) != "" {
			continue // skip if kp_id is filled
		}

		name := ""
		if len(row) > 0 && row[0] != nil {
			name = row[0].(string)
		}

		fmt.Printf("Processing: %s\n", name)

		query := url.Values{}
		query.Add("query", name)
		query.Add("page", "1")
		query.Add("limit", "3")

		reqURL := apiURL + "?" + query.Encode()
		req, _ := http.NewRequest("GET", reqURL, nil)
		req.Header.Set("accept", "application/json")
		req.Header.Set("X-API-KEY", getAPIKey())

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		if err != nil || resp.StatusCode == 403 {
			if !switchAPIKey() || err != nil {
				result = append(result, append(row, "Error")) // TODO
				continue
			}
			req.Header.Set("X-API-KEY", getAPIKey())
			resp, err = client.Do(req)
			if err != nil || resp.StatusCode != 200 {
				result = append(result, append(row, "Error"))
				continue
			}
		}

		var sr SearchResult
		if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
			result = append(result, append(row, "Error"))
			continue
		}

		var selected *struct {
			ID              int
			Name            string
			AlternativeName string
			Year            int
			IsSeries        bool
			ExternalID      struct{ KpHD, IMDB string }
			Rating          struct{ KP float64 }
			Votes           struct{ KP int }
		}

		originalName := ""
		if len(row) > 1 && row[1] != nil {
			originalName = row[1].(string)
		}
		year := 0
		if len(row) > 3 && row[3] != nil {
			fmt.Sscanf(row[3].(string), "%d", &year)
		}

		for _, d := range sr.Docs {
			if originalName != "" && d.AlternativeName == originalName {
				selected = &d
				break
			}
			if d.Name == name && d.Year == year {
				selected = &d
				break
			}
		}

		if selected != nil {
			newRow := make([]interface{}, len(headers))
			copy(newRow, row)

			if len(newRow) > 1 && newRow[1] == nil || newRow[1].(string) == "" {
				newRow[1] = selected.AlternativeName
			}
			if len(newRow) > 8 {
				newRow[8] = selected.ExternalID.KpHD
			}
			if len(newRow) > 9 && (newRow[9] == nil || newRow[9].(string) == "") {
				newRow[9] = selected.ExternalID.IMDB
			}
			if len(newRow) > 10 {
				newRow[10] = fmt.Sprintf("%t", selected.IsSeries)
			}
			if len(newRow) > 5 {
				newRow[5] = selected.Rating.KP
			}
			if len(newRow) > 11 {
				newRow[11] = selected.Votes.KP
			}
			if len(newRow) > 12 {
				newRow[12] = time.Now().Format("2006-01-02 15:04:05")
			}
			result = append(result, newRow)
		} else {
			result = append(result, append(row, "Not Found"))
		}

		time.Sleep(time.Second)
	}

	// Запись результата
	body := &sheets.ValueRange{
		Values: result,
	}
	_, err = srv.Spreadsheets.Values.Clear(spreadsheetID, resultSheet, nil).Do()
	if err != nil {
		log.Fatalf("Clear error: %v", err)
	}
	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, resultSheet, body).ValueInputOption("RAW").Do()
	if err != nil {
		log.Fatalf("Write error: %v", err)
	}
}
