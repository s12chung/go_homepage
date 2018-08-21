package markdown

type Settings struct {
	MarkdownsPath string `json:"path,omitempty"`
}

func DefaultSettings() *Settings {
	return &Settings{
		"./content/markdowns",
	}
}
