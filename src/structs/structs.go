package structs

type Source struct {
	Channel  string `json:"channel"`
	BotToken string `json:"bot_token"`

	MaxRetries int  `json:"max_retries"`
	Debug      bool `json:"debug"`
}

type Params struct {
	Blocks     string `json:"blocks"`
	BlocksFile string `json:"blocks_file"`
	Timestamp  string `json:"timestamp,omitempty"`
	ThreadTS   string `json:"thread_ts,omitempty"`
	Text       string `json:"text,omitempty"`
	Channel    string `json:"channel,omitempty"`
}

type PutInput struct {
	Source Source `json:"source"`
	Params Params `json:"params"`
}

type Version struct {
	Timestamp string `json:"timestamp"`
}

type GetInput struct {
	Source  Source  `json:"source"`
	Params  Params  `json:"params"`
	Version Version `json:"version"`
}

type Configuration struct {
	Blocks     string
	Timestamp  string
	ThreadTS   string
	Channel    string
	BotToken   string
	MaxRetries int
	Text       string
	Debug      bool
}
