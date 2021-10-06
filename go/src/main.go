package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

//games struct refers to JSON structure retrieved from website : https://www.epicgames.com/store/en-US/free-games
type games struct {
	Data struct {
		Catalog struct {
			SearchStore struct {
				Elements []struct {
					Title                string      `json:"title"`
					ID                   string      `json:"id"`
					EffectiveDate        time.Time   `json:"effectiveDate"`
					KeyImages            []struct {
						Type string `json:"type"`
						URL  string `json:"url"`
					} `json:"keyImages"`
					Price struct {
						TotalPrice struct {
							Discount        int `json:"discount"`
						} `json:"totalPrice"`
					} `json:"price"`
				} `json:"elements"`
			} `json:"searchStore"`
		} `json:"Catalog"`
	} `json:"data"`
	Extensions struct {
	} `json:"extensions"`
}

type SlackRequestBody struct {
	Text string `json:"text"`
}

var (
	driverName               = "mysql"
	dataSourceName           = "root:root@tcp(database:3306)/golang-docker"
	url                      = "https://store-site-backend-static-ipv4.ak.epicgames.com/freeGamesPromotions?locale=en-US&country=PL&allowCountries=PL"
	numberAffectedRows int64 = 0
	games1                   = games{}
)

func main() {
	db := getDB(driverName, dataSourceName)

	ticker := time.NewTicker(time.Second * 10)
	for ; true; <-ticker.C {
		games1 = games{}

		if dataByte, err := getData(url); err != nil {
			log.Printf("Failed to get JSON: %v", err)
		} else {
			jsonErr := json.Unmarshal(dataByte, &games1)
			if jsonErr != nil {
				log.Fatal(jsonErr)
			}

			insertNewGames(db, games1)
		}

		if numberAffectedRows > 0 {
			sendMessageForFreeGames(games1)
			numberAffectedRows = 0
		}
	}
	defer db.Close()
}

//insertNewGames insert data from games1 to specified database.
//Every successive operation of insert will increase numberAffectedRows.
func insertNewGames(db *sql.DB, games1 games) {
	for _, t := range games1.Data.Catalog.SearchStore.Elements {
		if t.Price.TotalPrice.Discount > 0 {
			results, err := db.Exec("INSERT INTO free_games (`site_id`, `name`, `image`)" +
				" VALUES ( '" + string(t.ID) + "', '" + string(t.Title) + "', '" + string(t.KeyImages[0].URL) + "')" +
				" ON DUPLICATE KEY UPDATE site_id = site_id ;")
			if err != nil {
				panic(err.Error())
			}
			count, _ := results.RowsAffected()
			numberAffectedRows += count
		}
	}
}

//sendMessageForFreeGames send message for every element of games1 that is discounted
//Message contains title of the game and time of discount
func sendMessageForFreeGames(gm games) {
	for _, e := range gm.Data.Catalog.SearchStore.Elements {
		if e.Price.TotalPrice.Discount > 0 {
			log.Println(e.Title)
			log.Println(e.KeyImages[0])
			msg := "New game available for free! " + e.Title + " for free for 7 days from " + e.EffectiveDate.Format("01-02-2006")
			sendMessage(msg)
		}
	}
}

//sendMessage send given message to webhook URL set in Slack Api
func sendMessage(message string) {
	webhookUrl := "https://hooks.slack.com/services/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"
	err := sendSlackNotification(webhookUrl, message)
	if err != nil {
		log.Fatal(err)
	}
}

//getDB open the connection to database specified by its database driver name and a driver-specific data source name
//If there is an error opening the connection, it will be handled.
func getDB(driverName string, dataSourceName string) *sql.DB {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(err.Error())
	}
	return db
}


//getData issues a GET to specified given URL and read data from response body.
//If function has no error, it returns data in []byte type. Otherwise, it returns an error.
func getData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status error: %v", resp.StatusCode)
	}
	data, readError := ioutil.ReadAll(resp.Body)
	if readError != nil {
		return nil, fmt.Errorf("Read body: %v", readError)
	}
	return data, nil
}

//sendSlackNotification will post to an 'Incoming Webhook' url setup in Slack Apps.
//It accepts text and the webhook URL connected to Slack channel.
func sendSlackNotification(webhookUrl string, msg string) error {
	slackBody, _ := json.Marshal(SlackRequestBody{Text: msg})
	req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack")
	}
	return nil
}