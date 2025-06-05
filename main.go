package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	defaultPort       = "8000"
	secretKeyHeader   = "X-Telegram-Bot-Api-Secret-Token"
	charsForsecretKey = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"
)

type config struct {
	port    string
	baseURL string
	botURL  string
}

func loadConfig() (config, error) {
	var c config
	c.port = os.Getenv("PORT")
	if c.port == "" {
		c.port = defaultPort
	}
	t := os.Getenv("BOT_TOKEN")
	if t == "" {
		return config{}, errors.New("missing BOT_TOKEN environment variable")
	}
	c.baseURL = fmt.Sprintf("https://api.telegram.org/bot%s/", t)
	c.botURL = os.Getenv("BOT_URL")
	if c.botURL == "" {
		return config{}, errors.New("missing BOT_URL environment variable")
	}
	return c, nil
}

type chat struct {
	ID int64 `json:"id"`
}

type message struct {
	ID    int64     `json:"message_id"`
	Chat  chat      `json:"chat"`
	Story *struct{} `json:"story"`
}

type update struct {
	Message message `json:"message"`
}

func (u *update) isStory() bool {
	return u.Message.Story != nil
}

type bot struct {
	config    config
	secretKey string
}

func (b *bot) post(n string, body io.Reader) error {
	u := fmt.Sprintf("%s%s", b.config.baseURL, n)
	r, err := http.Post(u, "application/json", body)
	if err != nil {
		return fmt.Errorf("error sending http request to %s: %w", n, err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		c, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("error reading http response body from %s: %w", n, err)
		}
		return fmt.Errorf("got unexpected status code from %s: %d, body: %s", n, r.StatusCode, string(c))
	}
	return nil
}

type deleteMessagePayload struct {
	ChatID    int64 `json:"chat_id"`
	MessageID int64 `json:"message_id"`
}

func (b *bot) deleteStory(m message) error {
	d := new(bytes.Buffer)
	p := deleteMessagePayload{m.Chat.ID, m.ID}
	if err := json.NewEncoder(d).Encode(p); err != nil {
		return fmt.Errorf("error encoding deleteMessage payload: %w", err)
	}
	return b.post("deleteMessage", d)
}

type setWebhookPayload struct {
	Url            string   `json:"url"`
	MaxConnections uint8    `json:"max_connections"`
	AllowedUpdates []string `json:"allowed_updates"`
	SecretToken    string   `json:"secret_token"`
}

func (b *bot) setWebhook() error {
	d := new(bytes.Buffer)
	p := setWebhookPayload{
		Url:            b.config.botURL,
		MaxConnections: 100,
		AllowedUpdates: []string{"message", "edited_message", "channel_post", "edited_channel_post", "business_message", "edited_business_message"},
		SecretToken:    b.secretKey,
	}
	if err := json.NewEncoder(d).Encode(p); err != nil {
		return fmt.Errorf("error encoding setWebhook payload: %w", err)
	}
	return b.post("setWebhook", d)
}

func (b *bot) deleteWebhook() error {
	return b.post("deleteWebhook", bytes.NewReader([]byte{}))
}

func (b *bot) handler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(secretKeyHeader) != b.secretKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var u update
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		slog.Error("failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if u.isStory() {
		if err := b.deleteStory(u.Message); err != nil {
			slog.Error("failed to delete story", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func newBot(c config) bot {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := make([]byte, r.Intn(128)+128)
	for i := range s {
		s[i] = charsForsecretKey[r.Intn(len(charsForsecretKey))]
	}
	b := bot{config: c, secretKey: string(s)}
	return b
}

func main() {
	c, err := loadConfig()
	if err != nil {
		slog.Error("error loading config", "error", err)
		os.Exit(1)
	}
	b := newBot(c)
	if err := b.setWebhook(); err != nil {
		slog.Error("error setting webhook", "error", err)
		os.Exit(1)
	}
	defer b.deleteWebhook()
	addr := fmt.Sprintf("0.0.0.0:%s", b.config.port)
	http.HandleFunc("/", b.handler)
	slog.Info(fmt.Sprintf("Serving at %s", addr))
	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("error running the webhook", "error", err)
		os.Exit(1)
	}
}
