package utils

import (
	"Reconciliation/config"
	"Reconciliation/ingest"
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func ParseAndStorePayments(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1


	// It is reading the first 20 lines of the CSV file to find the actual header row, which is the line that contains "date/time"
	var headers []string
	for i := 0; i < 20; i++ {
		line, err := reader.Read()
		if err != nil {
			return err
		}
		
		if len(line) > 0 && strings.Contains(line[0], "date/time") {
			headers = line
			break
		}
	}

	if len(headers) == 0 {
		return fmt.Errorf("headers not found")
	}

	recordsProcessed := 0
	
	for {
		line, err := reader.Read()
		if err != nil {
			break
		}
		
		if len(line) == 0 {
			continue
		}

		payment, err := ingest.PaymentFromCSVRow(headers, line)
		if err != nil || payment.OrderID == "" || payment.Total == 0 {
			continue
		}

		config.DB.Exec(`INSERT INTO records (source, order_id, date, total_amount, raw_data)
			VALUES ($1, $2, $3, $4, $5)`, 
			"payments", payment.OrderID, payment.Date, payment.Total, payment.RawData)
		recordsProcessed++
	}
	
	fmt.Printf("Processed %d payment records\n", recordsProcessed)
	return nil
}

func ParseAndStoreSettlements(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	if !scanner.Scan() {
		return fmt.Errorf("empty file")
	}
	
	headers := strings.Split(scanner.Text(), "\t")
	var settlements []*ingest.Settlement
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		
		fields := strings.Split(line, "\t")
		if len(fields) < len(headers) {
			continue
		}

		settlement, err := ingest.SettlementFromTSVRow(headers, fields)
		if err != nil || settlement.OrderID == "" {
			continue
		}

		settlements = append(settlements, settlement)
	}

	orderTotals := ingest.AggregateSettlementsByOrderID(settlements)

	for orderID, total := range orderTotals {
		var firstSettlement *ingest.Settlement
		for _, s := range settlements {
			if s.OrderID == orderID {
				firstSettlement = s
				break
			}
		}

		if firstSettlement != nil {
			config.DB.Exec(`INSERT INTO records (source, order_id, date, total_amount, raw_data)
				VALUES ($1, $2, $3, $4, $5)`, 
				"settlements", orderID, firstSettlement.PostedDateTime, total, firstSettlement.RawData)
		}
	}

	fmt.Printf("Processed %d settlement records for %d orders\n", len(settlements), len(orderTotals))
	return nil
}
