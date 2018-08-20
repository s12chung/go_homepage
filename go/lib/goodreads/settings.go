package goodreads

type Settings struct {
	CachePath  string `json:"cache_path,omitempty"`
	ApiURL     string `json:"api_url,omitempty"`
	ApiKey     string `json:"api_key,omitempty"`
	UserId     int    `json:"user_id,omitempty"`
	PerPage    int    `json:"per_page,omitempty"`
	MaxPerPage int    `json:"max_per_page,omitempty"`
	RateLimit  int    `json:"rate_limit,omitempty"`
}

func (settings *Settings) invalid() bool {
	return settings.ApiKey == "" && settings.UserId == 0
}

func DefaultSettings() *Settings {
	return &Settings{
		"./cache",
		"https://www.goodreads.com",
		"",
		0,
		50,
		200,
		1000,
	}
}

func TestSettings(cachePath, apiUrl string) *Settings {
	settings := DefaultSettings()
	settings.ApiURL = apiUrl
	settings.ApiKey = "good_test"
	settings.UserId = 1
	settings.RateLimit = 1
	settings.CachePath = cachePath
	return settings
}

func InvalidateSettings(settings *Settings) {
	settings.ApiKey = ""
	settings.UserId = 0
}
