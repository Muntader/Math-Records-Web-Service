package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func handleRequest() {
	router := mux.NewRouter()
	router.HandleFunc("/install", CreateDatabse).Methods("GET")
	router.HandleFunc("/api/v1", HomeHandler).Methods("GET")
	router.HandleFunc("/api/v1/register", Register).Methods("POST")
	router.HandleFunc("/api/v1/login", Login).Methods("POST")

	// USER AUTH
	router.Handle("/api/v1/get/records", MiddlewareAuth(http.HandlerFunc(GetAllRecords))).Methods("GET")
	router.Handle("/api/v1/create/record", MiddlewareAuth(http.HandlerFunc(CreateRecord))).Methods("POST")
	router.Handle("/api/v1/update/record/{id}", MiddlewareAuth(http.HandlerFunc(UpdateRecord))).Methods("POST")
	router.Handle("/api/v1/update/record/freeze/{id}", MiddlewareAuth(http.HandlerFunc(FreezeRecored))).Methods("PUT")
	router.Handle("/api/v1/get/record/history/{id}", MiddlewareAuth(http.HandlerFunc(GetRecordHistory))).Methods("GET")
	router.Handle("/api/v1/delete/record/{id}", MiddlewareAuth(http.HandlerFunc(DeleteRecord))).Methods("DELETE")
	router.Handle("/api/v1/update/record/rollback/{id}", MiddlewareAuth(http.HandlerFunc(RollBackRecord))).Methods("POST")

	// Basic middleware
	http.ListenAndServe(":80", router)
}

func main() {
	handleRequest()

}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "API")
}

func MiddlewareAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get Authorization Value
		authorizationHeader := r.Header.Get("Authorization")

		// Check If Authorization value match Access token in Database
		authMatch := CheckAuthenticated(authorizationHeader)
		fmt.Println(authMatch)

		if !authMatch {
			// Return Unauthorized response
			w.WriteHeader(401)
			w.Header().Set("Content-Type", "application/json")
			resErr := []byte(`{"code": 401,  message": "Unauthorized"}`)
			w.Write(resErr)
			return
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		h.ServeHTTP(w, r)
	})
}
