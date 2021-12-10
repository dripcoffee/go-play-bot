package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-lark/lark"
	"github.com/go-lark/lark-gin"
)

const (
	playgroundCompileURL = "https://play.golang.org/compile"
	playgroundFmtURL     = "https://play.golang.org/fmt"
)

type codeSession struct {
	chatID int
	code   string
	result string
}

type compileEvent struct {
	Delay   int    `json:"Delay"`
	Message string `json:"Message"`
	Kind    string `json:"Kind"`
}

type compileBody struct {
	Errors string         `json:"Errors"`
	Events []compileEvent `json:"Events"`
}

type fmtBody struct {
	Body  string `json:"Body"`
	Error string `json:"Error"`
}

var client *http.Client

func main() {
	var (
		appID     = os.Getenv("LARK_APP_ID")
		appSecret = os.Getenv("LARK_APP_SECRET")
	)
	client = http.DefaultClient

	r := gin.Default()
	middleware := larkgin.NewLarkMiddleware()
	r.Use(middleware.LarkChallengeHandler())
	r.Use(middleware.LarkMessageHandler())

	bot := lark.NewChatBot(appID, appSecret)
	bot.GetAccessTokenInternal(true)
	bot.StartHeartbeat()

	r.POST("/go", func(c *gin.Context) {
		message, ok := middleware.GetMessage(c)
		if !ok {
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
		})

		text := message.Event.RealText
		cmd, code := parseToken(text)
		var result string
		if cmd == "run" {
			resp, err := runWithPlayground(code)
			if err != nil {
				bot.PostTextMention("Run failed", message.Event.OpenID, lark.WithChatID(message.Event.OpenChatID))
				return
			}
			if len(resp.Errors) != 0 {
				bot.PostTextMention(resp.Errors, message.Event.OpenID, lark.WithChatID(message.Event.OpenChatID))
				return
			}

			for _, event := range resp.Events {
				result += fmt.Sprintf("Delay: %d [%s] %s\n", event.Delay, event.Kind, event.Message)
			}
			if result == "" {
				result = "Done without errors or outputs."
			}
		} else if cmd == "fmt" {
			resp, err := fmtWithPlayground(code)
			if err != nil {
				bot.PostTextMention("Fmt failed", message.Event.OpenID, lark.WithChatID(message.Event.OpenChatID))
				return
			}
			if len(resp.Error) != 0 {
				bot.PostTextMention(resp.Error, message.Event.OpenID, lark.WithChatID(message.Event.OpenChatID))
				return
			}
			result = resp.Body
		} else {
			result = "run\nfmt\nhelp"
		}
		bot.PostTextMention(result, message.Event.OpenID, lark.WithChatID(message.Event.OpenChatID))
	})

	r.Run(":9967")
}

func parseToken(text string) (string, string) {
	tokens := strings.SplitN(text, "\n", 2)
	if len(tokens) == 0 {
		return "", ""
	}
	token := strings.Trim(tokens[0], " ")
	code := ""
	if len(tokens) > 1 {
		code = tokens[1]
	}
	return token, code
}

func runWithPlayground(code string) (*compileBody, error) {
	resp, err := client.PostForm(playgroundCompileURL, url.Values{
		"version": {"2"},
		"body":    {code},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result compileBody
	err = json.NewDecoder(resp.Body).Decode(&result)
	return &result, err
}

func fmtWithPlayground(code string) (*fmtBody, error) {
	resp, err := client.PostForm(playgroundFmtURL, url.Values{
		"body": {code},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result fmtBody
	err = json.NewDecoder(resp.Body).Decode(&result)
	return &result, err
}
