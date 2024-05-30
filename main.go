package main

import (
	"encoding/json"
	"flag"

	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
)

const DBPATH string = "./database.json"

func main() {
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		deleteDB(DBPATH)
	}

	// Use the http.NewServeMux() function to create an empty servemux.
	const root = "."
	const port = "8080"
	apiCfg := apiConfig{
		fileserverHits: 0,
	}
	router := mux.NewRouter()

	defaultHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(root))))
	router.Handle("/app/*", middlewareLog(defaultHandler))
	router.HandleFunc("/admin/metrics", apiCfg.handlerCount).Methods("GET")
	router.HandleFunc("/api/healthz", handlerStatus).Methods("GET")
	router.HandleFunc("/api/chirps", handlerAddChirp).Methods("POST")
	router.HandleFunc("/api/chirps", handlerGetChirps).Methods("GET")
	router.HandleFunc("/api/chirps/{chirpID}", handlerAddChirpId).Methods("GET")
	router.HandleFunc("/api/users", handlerAddUser).Methods("POST")
	router.HandleFunc("/api/reset", apiCfg.handlerResetCount)
	corsMux := middlewareLog(middlewareCors(router))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
	srv.ListenAndServe()
}

func handlerAddChirp(w http.ResponseWriter, req *http.Request) {
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	chirp := POST{}
	json.NewDecoder(req.Body).Decode(&chirp)

	if len(chirp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	newChirp, err := db.createChirp(censor(chirp.Body))
	if err != nil {
		panic(err)
	}
	respondWithJSON(w, 201, newChirp)
}

func handlerGetChirps(w http.ResponseWriter, req *http.Request) {
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	respondWithJSON(w, 200, db.getChirps())
}

func handlerAddChirpId(w http.ResponseWriter, req *http.Request) {
	chirpID, ok := mux.Vars(req)["chirpID"]
	if !ok {
		respondWithError(w, 400, "id is missing in parameters")
		return
	}
	id, err := strconv.Atoi(chirpID)
	if err != nil {
		respondWithError(w, 400, err.Error())

		return
	}
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	chirpMap, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	chirp, found := chirpMap.Chrips[id]
	if !found {
		respondWithError(w, 404, "Chirp not found")
		return
	}
	respondWithJSON(w, 200, chirp)
}

func handlerAddUser(w http.ResponseWriter, req *http.Request) {
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}

	chirp := POST{}
	json.NewDecoder(req.Body).Decode(&chirp)
	newChirp, err := db.createUser(chirp.Email)
	if err != nil {
		panic(err)
	}
	respondWithJSON(w, 201, newChirp)
}

func censor(s string) string {
	re := regexp.MustCompile(`(?i)kerfuffle|sharbert|fornax`)
	return re.ReplaceAllString(s, "****")
}

// To decode into
type POST struct {
	Body  string `json:"body"`
	Email string `json:"email"`
}
