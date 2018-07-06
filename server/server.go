package server

import (
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

func Run(targetDir string, port int) error {
	log.Infof("Serving '%v' at http://localhost:%v/", targetDir, port)
	handler := http.FileServer(http.Dir(targetDir))
	return http.ListenAndServe(":"+strconv.Itoa(port), handler)
}
