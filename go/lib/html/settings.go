package html

type Settings struct {
	TemplatePath string `json:"template_path,omitempty"`
	TemplateExt  string `json:"template_ext,omitempty"`
	LayoutName   string `json:"layoutName,omitempty"`
	WebsiteTitle string `json:"website_title,omitempty"`
}

func DefaultSettings() *Settings {
	return &Settings{
		"./go/content/templates",
		".tmpl",
		"layout",
		"Your Website Title",
	}
}
