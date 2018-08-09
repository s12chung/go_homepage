package goodreads

type Settings struct {
	CachePath  string `json:"cache_path,omitempty"`
	ApiKey     string `json:"api_key,omitempty"`
	UserId     int    `json:"user_id,omitempty"`
	PerPage    int    `json:"per_page,omitempty"`
	MaxPerPage int    `json:"max_per_page,omitempty"`
	RateLimit  int    `json:"rate_limit,omitempty"`
}

func DefaultSettings() *Settings {
	return &Settings{
		"./cache",
		"",
		0,
		50,
		200,
		1000,
	}
}
