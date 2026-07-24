package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "../prism.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var configName string
	err = db.QueryRowContext(context.Background(), `Select ConfigName from "Configurations" where "ConfigName" like ?`, "Scat-V-HalfTx-PPM-A").Scan(&configName)
	if err != nil {
		fmt.Println("Error for Scat-V-HalfTx-PPM-A:", err)
	} else {
		fmt.Println("Found for Scat-V-HalfTx-PPM-A:", configName)
	}

	err = db.QueryRowContext(context.Background(), `Select ConfigName from "Configurations" where "ConfigName" like ?`, "Scat-V-HalfTx-PPM-A ").Scan(&configName)
	if err != nil {
		fmt.Println("Error for Scat-V-HalfTx-PPM-A (with space):", err)
	} else {
		fmt.Println("Found for Scat-V-HalfTx-PPM-A (with space):", configName)
	}

	err = db.QueryRowContext(context.Background(), `Select ConfigName from "Configurations" where "ConfigName" like ?`, "scat-H-HalfTx-PPMB").Scan(&configName)
	if err != nil {
		fmt.Println("Error for scat-H-HalfTx-PPMB:", err)
	} else {
		fmt.Println("Found for scat-H-HalfTx-PPMB:", configName)
	}
}
