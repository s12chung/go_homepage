package models

import "github.com/sirupsen/logrus"

type Factory struct {
	postsPath  string
	draftsPath string
	githubUrl  string
	log        logrus.FieldLogger
}

var factory *Factory

func Config(postsPath, draftsPath, githubUrl string, log logrus.FieldLogger) {
	factory = NewFactory(postsPath, draftsPath, githubUrl, log)
}

func NewFactory(postsPath, draftsPath, githubUrl string, log logrus.FieldLogger) *Factory {
	return &Factory{
		postsPath,
		draftsPath,
		githubUrl,
		log,
	}
}
