package controllers

import (
	"Reconciliation/config"
)

func RunReconciliation() error {
	config.DB.Exec("DELETE FROM reconciled_records")

	query := `
		SELECT p.id, p.order_id, p.total_amount, s.id, s.total_amount
		FROM records p
		JOIN records s ON p.order_id = s.order_id
		WHERE p.source = 'payments' AND s.source = 'settlements'`

	rows, err := config.DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var paymentId, settlementId int
		var orderId string
		var paymentTotal, settlementTotal float64

		rows.Scan(&paymentId, &orderId, &paymentTotal, &settlementId, &settlementTotal)

		diff := paymentTotal - settlementTotal
		config.DB.Exec(`
			INSERT INTO reconciled_records (payments_record_id, settlements_record_id, amount_difference)
			VALUES ($1, $2, $3)`, paymentId, settlementId, diff)
	}

	return nil
}
