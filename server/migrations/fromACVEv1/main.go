// Script de migration depuis la v1 ACVE.
//
// Ce script suppose que la base v2 a déjà été créée (sans données) et
//   - charge les données v1
//   - nettoie en conservant uniquement les années récentes
//   - convertit au format v2 et insère
package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"registro/config"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	dbV1, err := connectV1()
	check(err)
	defer dbV1.Close()

	dbV2, err := connectV2()
	check(err)
	defer dbV2.Close()

	fmt.Println("Connected.")
}

func connectV1() (*sql.DB, error) {
	port := 5432
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		v1_Host, port, v1_User, v1_Password, v1_Name)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("connexion DB : %s", err)
	}
	return db, nil
}

func connectV2() (*sql.DB, error) {
	config, err := config.NewDB()
	if err != nil {
		return nil, err
	}
	return config.ConnectPostgres()
}
