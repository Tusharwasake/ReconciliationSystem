package models

import "time"

type Record struct {
	ID          int       `db:"id"`
	Source      string    `db:"source"`
	OrderID     string    `db:"order_id"`
	Date        time.Time `db:"date"`
	TotalAmount float64   `db:"total_amount"`
	RawData     string    `db:"raw_data"`
}

type ReconciledRecord struct {
	ID                  int     `db:"id"`
	PaymentsRecordID    int     `db:"payments_record_id"`
	SettlementsRecordID int     `db:"settlements_record_id"`
	AmountDifference    float64 `db:"amount_difference"`
}
