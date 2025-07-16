package views

import (
	"Reconciliation/config"
	"encoding/csv"
	"os"
	"strconv"
)

func GenerateCSVReport() error {
	// Create output directory if it doesn't exist
	os.MkdirAll("output", 0755)

	rows, err := config.DB.Query(`
		SELECT p.order_id, p.total_amount, s.total_amount, r.amount_difference
		FROM reconciled_records r
		JOIN records p ON r.payments_record_id = p.id
		JOIN records s ON r.settlements_record_id = s.id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	file, err := os.Create("output/reconciliation_report.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header as per assignment requirements
	writer.Write([]string{"order_id", "status", "payments_total", "settlements_total", "difference"})

	for rows.Next() {
		var orderID string
		var paymentsTotal, settlementsTotal, difference float64

		if err := rows.Scan(&orderID, &paymentsTotal, &settlementsTotal, &difference); err != nil {
			return err
		}

		status := "reconciled"
		if difference != 0 {
			status = "unreconciled"
		}

		writer.Write([]string{
			orderID,
			status,
			strconv.FormatFloat(paymentsTotal, 'f', 2, 64),
			strconv.FormatFloat(settlementsTotal, 'f', 2, 64),
			strconv.FormatFloat(difference, 'f', 2, 64),
		})
	}

	return nil
}
