package controllers

import (
	"Reconciliation/config"
	"Reconciliation/utils"
)

func ClearExistingData() error {
	config.DB.Exec("DELETE FROM reconciled_records")
	config.DB.Exec("DELETE FROM records")
	config.DB.Exec("ALTER SEQUENCE records_id_seq RESTART WITH 1")
	config.DB.Exec("ALTER SEQUENCE reconciled_records_id_seq RESTART WITH 1")
	return nil
}

func IngestAllFiles(paymentPath, settlementPath string) error {
	if err := ClearExistingData(); err != nil {
		return err
	}
	
	if err := utils.ParseAndStorePayments(paymentPath); err != nil {
		return err
	}
	
	return utils.ParseAndStoreSettlements(settlementPath)
}
