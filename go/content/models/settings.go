package models

type Settings struct {
	PostsPath  string `json:"generated_path,omitempty"`
	DraftsPath string `json:"generated_path,omitempty"`
	GithubUrl  string `json:"github_url,omitempty"`
}

func DefaultSettings() *Settings {
	return &Settings{
		"./content/posts",
		"./content/drafts",
		"",
	}
}
