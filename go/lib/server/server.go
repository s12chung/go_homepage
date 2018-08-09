package server

import (
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
)

func RunFileServer(targetDir string, port int, log logrus.FieldLogger) error {
	log.Infof("Serving files from '%v' at http://localhost:%v/", targetDir, port)
	handler := http.FileServer(http.Dir(targetDir))
	return http.ListenAndServe(":"+strconv.Itoa(port), handler)
}
