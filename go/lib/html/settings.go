package html

type Settings struct {
	TemplatePath string `json:"template_path,omitempty"`
	WebsiteTitle string `json:"website_title,omitempty"`
}

func DefaultSettings() *Settings {
	return &Settings{
		"./go/content/templates",
		"Your Website Title",
	}
}
