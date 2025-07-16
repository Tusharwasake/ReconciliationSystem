package main

import (
	"Reconciliation/config"
	"Reconciliation/controllers"
	"Reconciliation/views"
	"log"
)

func main() {
	if err := config.Connect(); err != nil {
		log.Fatal(err)
	}

	if err := config.RunMigrations(); err != nil {
		log.Fatal(err)
	}

	if err := controllers.IngestAllFiles("data/payment_data.csv", "data/settlement_data.txt"); err != nil {
		log.Fatal(err)
	}

	if err := controllers.RunReconciliation(); err != nil {
		log.Fatal(err)
	}

	if err := views.GenerateCSVReport(); err != nil {
		log.Fatal(err)
	}

	log.Println("Done")
}
