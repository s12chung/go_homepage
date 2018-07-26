package models

type Factory struct {
	postsPath  string
	draftsPath string
}

var F *Factory

func Config(postsPath, draftsPath string) {
	F = &Factory{postsPath, draftsPath}
}
