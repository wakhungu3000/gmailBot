package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const telegramAPI = "https://api.telegram.org/bot"

type update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

func telegramRequest(token, method string, values url.Values) ([]byte, error) {
	resp, err := http.PostForm(telegramAPI+token+"/"+method, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Telegram API returned %s: %s", resp.Status, body)
	}
	return body, nil
}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	offset := 0
	for {
		body, err := telegramRequest(token, "getUpdates", url.Values{
			"offset":          {strconv.Itoa(offset)},
			"timeout":         {"30"},
			"allowed_updates": {`["message"]`},
		})
		if err != nil {
			log.Println("polling error:", err)
			time.Sleep(time.Second)
			continue
		}

		var result struct {
			OK      bool     `json:"ok"`
			Updates []update `json:"result"`
		}
		if err := json.Unmarshal(body, &result); err != nil || !result.OK {
			log.Println("invalid Telegram response:", err)
			time.Sleep(time.Second)
			continue
		}

		for _, item := range result.Updates {
			offset = item.UpdateID + 1
			if item.Message.Text == "" {
				continue
			}
			_, err := telegramRequest(token, "sendMessage", url.Values{
				"chat_id": {strconv.FormatInt(item.Message.Chat.ID, 10)},
				"text":    {"You said: " + item.Message.Text},
			})
			if err != nil {
				log.Println("send error:", err)
			}
		}
	}
}
