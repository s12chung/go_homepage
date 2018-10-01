package models

type Settings struct {
	PostsPath  string `json:"posts_path,omitempty"`
	DraftsPath string `json:"drafts_path,omitempty"`
	GithubURL  string `json:"github_url,omitempty"`
}

func DefaultSettings() *Settings {
	return &Settings{
		"./content/posts",
		"./content/drafts",
		"",
	}
}
