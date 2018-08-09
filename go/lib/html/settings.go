package html

type Settings struct {
	WebsiteTitle  string `json:"website_title,omitempty"`
	MarkdownsPath string `json:"markdowns_path,omitempty"`
}

func DefaultSettings() *Settings {
	return &Settings{
		"Your Website Title",
		"./assets/markdowns",
	}
}
