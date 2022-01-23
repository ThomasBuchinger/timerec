package restapi

import "net/http"

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
	w.WriteHeader(http.StatusOK)
}
func readinessCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ready"))
	w.WriteHeader(http.StatusOK)
}
