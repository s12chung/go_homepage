package router

type Tracker struct {
	AllUrls       func() ([]string, error)
	dependentUrls map[string]bool
}

func NewTracker(allUrls func() ([]string, error)) *Tracker {
	return &Tracker{allUrls, map[string]bool{}}
}

func (tracker *Tracker) AddDependentUrl(url string) {
	tracker.dependentUrls[url] = true
}

func (tracker *Tracker) IndependentUrls() ([]string, error) {
	allUrls, err := tracker.AllUrls()
	if err != nil {
		return nil, err
	}

	independentUrls := make([]string, len(allUrls)-len(tracker.dependentUrls))
	i := 0
	for _, url := range allUrls {
		_, exists := tracker.dependentUrls[url]
		if !exists {
			independentUrls[i] = url
			i += 1
		}
	}
	return independentUrls, nil
}

func (tracker *Tracker) DependentUrls() []string {
	urls := make([]string, len(tracker.dependentUrls))
	i := 0
	for url := range tracker.dependentUrls {
		urls[i] = url
		i += 1
	}
	return urls
}
