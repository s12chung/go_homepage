package robots

import "strings"

const EverythingUserAgent = "*"

func ToFileString(userAgents []*UserAgent) string {
	parts := make([]string, len(userAgents))
	for i, userAgent := range userAgents {
		parts[i] = userAgent.ToFileString()
	}
	return strings.Join(parts, "\n\n")
}

type UserAgent struct {
	name  string
	paths []string
}

func NewUserAgent(name string, paths []string) *UserAgent {
	return &UserAgent{name, paths}
}

func (userAgent *UserAgent) ToFileString() string {
	parts := []string{"User-agent: " + userAgent.name}

	for _, path := range userAgent.paths {
		parts = append(parts, "Disallow: "+path)
	}
	return strings.Join(parts, "\n")
}
