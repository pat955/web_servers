package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

var NEWID int = 1

type DB struct {
	Path string
	mux  *sync.RWMutex
}
type DBStructure struct {
	Chrips map[int]Chirp `json:"chirps"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func createDB(path string) (*DB, error) {
	fmt.Println("Creating db")
	db := DB{Path: path, mux: &sync.RWMutex{}}

	if _, err := os.Stat(path); err == nil {
		return &db, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	db.writeDB(DBStructure{Chrips: make(map[int]Chirp)})
	return &db, nil
}

func deleteDB(path string) {
	fmt.Println("Deleting db")

	os.Remove(path)
}

func (db *DB) createChirp(body string) (Chirp, error) {
	if 140 >= len(body) && len(body) >= 1 {
		newChirp := Chirp{Id: NEWID, Body: body}

		chirpMap, err := db.loadDB()
		if err != nil {
			panic(err)
		}
		chirpMap.Chrips[NEWID] = newChirp
		db.writeDB(chirpMap)
		NEWID++
		return newChirp, nil
	} else {
		return Chirp{}, errors.New("invalid chirp")
	}
}

func (db *DB) writeDB(dbstruct DBStructure) {
	json, err := json.Marshal(dbstruct)
	if err != nil {
		panic(err)
	}
	db.mux.Lock()
	os.WriteFile(db.Path, json, os.ModePerm)
	db.mux.Unlock()
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	f, err := os.ReadFile(db.Path)
	if err != nil {
		panic(err)
	}
	db.mux.RUnlock()
	var dbStruct DBStructure
	json.Unmarshal(f, &dbStruct)

	return dbStruct, nil
}
func (db *DB) getChirps() []Chirp {
	db.mux.RLock()
	defer db.mux.RUnlock()
	f, err := os.ReadFile(db.Path)
	if err != nil {
		panic(err)
	}
	var chirpMap DBStructure
	json.Unmarshal(f, &chirpMap)
	var allChirps []Chirp
	for _, chirp := range chirpMap.Chrips {
		allChirps = append(allChirps, chirp)
	}
	return allChirps
}
