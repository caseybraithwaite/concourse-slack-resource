package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/caseybraithwaite/concourse-slack-resource/structs"
	"github.com/slack-go/slack"
)

func validate(request *structs.PutInput) error {
	if request.Source.BotToken == "" {
		return fmt.Errorf("bot token is a required parameter")
	}

	if request.Source.Channel == "" && request.Params.Channel == "" {
		return fmt.Errorf("channel is a required parameter in either source or params")
	}

	if request.Params.Blocks == "" && request.Params.BlocksFile == "" && request.Params.Text == "" {
		return fmt.Errorf("blocks, blocks_file or text are required")
	}

	if request.Params.Timestamp != "" && request.Params.ThreadTS != "" {
		return fmt.Errorf("you can't provide both timestamp and thread timestamp")
	}

	return nil
}

func readFile(fileName string) (string, error) {
	f, err := os.Open(fmt.Sprintf("%s/%s", "/tmp/build/put", fileName))
	if err != nil {
		return "", err
	}

	bytes, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func interpolate(text string) string {
	re := regexp.MustCompile(`{{([^}]+)}}`)

	return re.ReplaceAllStringFunc(
		text,
		func(s string) string {
			varName := strings.ReplaceAll(s, "{", "")
			varName = strings.ReplaceAll(varName, "}", "")
			varName = strings.ReplaceAll(varName, "$", "")
			varName = strings.TrimSpace(varName)

			v, ok := os.LookupEnv(varName)
			if ok {
				return v
			}
			return s
		})
}

func configure(request *structs.PutInput) (structs.Configuration, error) {
	cfg := structs.Configuration{
		Channel:  request.Source.Channel,
		BotToken: request.Source.BotToken,
		Debug:    request.Source.Debug,
	}

	if request.Params.Channel != "" {
		cfg.Channel = request.Params.Channel
	}

	// Default max retries to 1 if not set
	if request.Source.MaxRetries == 0 {
		cfg.MaxRetries = 1
	} else {
		cfg.MaxRetries = request.Source.MaxRetries
	}

	stringBlocks := ""
	if request.Params.BlocksFile != "" {
		b, err := readFile(request.Params.BlocksFile)
		if err != nil {
			return cfg, fmt.Errorf("error reading blocks file '%s': %v", request.Params.BlocksFile, err)
		}
		stringBlocks = b
	} else if request.Params.Blocks != "" {
		stringBlocks = request.Params.Blocks
	}

	if stringBlocks != "" {
		// interpolate environment variables
		cfg.Blocks = interpolate(stringBlocks)
	}

	if request.Params.Timestamp != "" {
		cfg.Timestamp = request.Params.Timestamp
	}

	if request.Params.ThreadTS != "" {
		cfg.ThreadTS = request.Params.ThreadTS
	}

	if request.Params.Text != "" {
		cfg.Text = interpolate(request.Params.Text)
	}

	return cfg, nil
}

func sendMessage(api *slack.Client, channel string, maxRetries int, options []slack.MsgOption) (string, error) {
	for i := 0; i < maxRetries; i++ {
		_, timestamp, err := api.PostMessage(
			channel,
			options...,
		)

		if err == nil {
			return timestamp, nil
		} else {
			if rateLimitedError, ok := err.(*slack.RateLimitedError); ok {
				if rateLimitedError.Retryable() {
					log.Printf("hit rate limit - retrying after %d\n", rateLimitedError.RetryAfter)
					time.Sleep(rateLimitedError.RetryAfter)
				}
			} else if strings.Contains(strings.ToLower(err.Error()), "internal server error") {
				log.Println("internal server error - retrying after 3s")
				time.Sleep(time.Second * 3)
			} else {
				return "", err
			}
		}
	}

	return "", fmt.Errorf("couldn't send message - hit max retries")
}

func updateMessage(api *slack.Client, channel string, timestamp string, maxRetries int, options []slack.MsgOption) (string, error) {
	for i := 0; i < maxRetries; i++ {
		_, timestamp, _, err := api.UpdateMessage(
			channel,
			timestamp,
			options...,
		)

		if err == nil {
			return timestamp, nil
		} else {
			if rateLimitedError, ok := err.(*slack.RateLimitedError); ok {
				if rateLimitedError.Retryable() {
					log.Printf("hit rate limit - retrying after %d\n", rateLimitedError.RetryAfter)
					time.Sleep(rateLimitedError.RetryAfter)
				}
			} else if strings.Contains(strings.ToLower(err.Error()), "internal server error") {
				log.Println("internal server error - retrying after 3s")
				time.Sleep(time.Second * 3)
			} else {
				return "", err
			}
		}
	}

	return "", fmt.Errorf("couldn't update message - hit max retries")
}

func execute(cfg *structs.Configuration) (string, error) {
	// new slack client
	api := slack.New(cfg.BotToken, slack.OptionDebug(cfg.Debug))

	// parse blocks json into slack block objects
	blocks := slack.Blocks{}
	if cfg.Blocks != "" {
		err := blocks.UnmarshalJSON([]byte(string(cfg.Blocks)))
		if err != nil {
			return "", err
		}
	}

	// if timestamp is provided then update the message
	if cfg.Timestamp != "" {
		updateOptions := []slack.MsgOption{}
		if cfg.Blocks != "" {
			updateOptions = append(updateOptions, slack.MsgOptionBlocks(blocks.BlockSet...))
		}
		if cfg.Text != "" {
			updateOptions = append(updateOptions, slack.MsgOptionText(cfg.Text, false))
		}

		return updateMessage(api, cfg.Channel, cfg.Timestamp, cfg.MaxRetries, updateOptions)
	}

	// otherwise send a message
	sendOptions := []slack.MsgOption{
		slack.MsgOptionDisableLinkUnfurl(),
	}
	if cfg.Blocks != "" {
		sendOptions = append(sendOptions, slack.MsgOptionBlocks(blocks.BlockSet...))
	}
	if cfg.ThreadTS != "" {
		sendOptions = append(sendOptions, slack.MsgOptionTS(cfg.ThreadTS))
	}
	if cfg.Text != "" {
		sendOptions = append(sendOptions, slack.MsgOptionText(cfg.Text, false))
	}

	// basic message
	return sendMessage(api, cfg.Channel, cfg.MaxRetries, sendOptions)
}

func main() {
	var putRequest structs.PutInput
	err := json.NewDecoder(os.Stdin).Decode(&putRequest)
	if err != nil {
		log.Fatal("couldn't decode stdin: ", err)
	}

	// check required params have been set
	err = validate(&putRequest)
	if err != nil {
		log.Fatal("error whilst validating options: ", err)
	}

	cfg, err := configure(&putRequest)
	if err != nil {
		log.Fatal("error whilst setting config", err)
	}

	timestamp, err := execute(&cfg)
	if err != nil {
		log.Fatal("error whilst sending message to slack: ", err)
	}

	// construct response
	rsp := Response{
		Version: ResponseVersion{
			Timestamp: timestamp,
		},
	}
	err = json.NewEncoder(os.Stdout).Encode(rsp)
	if err != nil {
		log.Fatal("couldn't encode response", err)
	}
}
