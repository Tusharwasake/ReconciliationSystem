package ingest

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Settlement struct {
	ID                       int       `json:"id" db:"id"`
	SettlementID             string    `json:"settlement_id" db:"settlement_id"`
	SettlementStartDate      string    `json:"settlement_start_date" db:"settlement_start_date"`
	SettlementEndDate        string    `json:"settlement_end_date" db:"settlement_end_date"`
	DepositDate              string    `json:"deposit_date" db:"deposit_date"`
	TotalAmount              float64   `json:"total_amount" db:"total_amount"`
	Currency                 string    `json:"currency" db:"currency"`
	TransactionType          string    `json:"transaction_type" db:"transaction_type"`
	OrderID                  string    `json:"order_id" db:"order_id"`
	MerchantOrderID          string    `json:"merchant_order_id" db:"merchant_order_id"`
	AdjustmentID             string    `json:"adjustment_id" db:"adjustment_id"`
	ShipmentID               string    `json:"shipment_id" db:"shipment_id"`
	MarketplaceName          string    `json:"marketplace_name" db:"marketplace_name"`
	AmountType               string    `json:"amount_type" db:"amount_type"`
	AmountDescription        string    `json:"amount_description" db:"amount_description"`
	Amount                   float64   `json:"amount" db:"amount"`
	FulfillmentID            string    `json:"fulfillment_id" db:"fulfillment_id"`
	PostedDate               string    `json:"posted_date" db:"posted_date"`
	PostedDateTime           time.Time `json:"posted_date_time" db:"posted_date_time"`
	OrderItemCode            string    `json:"order_item_code" db:"order_item_code"`
	MerchantOrderItemID      string    `json:"merchant_order_item_id" db:"merchant_order_item_id"`
	MerchantAdjustmentItemID string    `json:"merchant_adjustment_item_id" db:"merchant_adjustment_item_id"`
	SKU                      string    `json:"sku" db:"sku"`
	QuantityPurchased        int       `json:"quantity_purchased" db:"quantity_purchased"`
	RawData                  string    `json:"raw_data" db:"raw_data"`
}

// SettlementFromTSVRow creates a Settlement from TSV row data
func SettlementFromTSVRow(headers []string, row []string) (*Settlement, error) {
	if len(row) < len(headers) {
		return nil, fmt.Errorf("row has fewer fields than headers")
	}

	settlement := &Settlement{}

	// Create a map for easier field access
	data := make(map[string]string)
	for i, header := range headers {
		if i < len(row) {
			data[strings.TrimSpace(header)] = strings.TrimSpace(row[i])
		}
	}

	// Parse string fields
	settlement.SettlementID = data["settlement-id"]
	settlement.SettlementStartDate = data["settlement-start-date"]
	settlement.SettlementEndDate = data["settlement-end-date"]
	settlement.DepositDate = data["deposit-date"]
	settlement.Currency = data["currency"]
	settlement.TransactionType = data["transaction-type"]
	settlement.OrderID = data["order-id"]
	settlement.MerchantOrderID = data["merchant-order-id"]
	settlement.AdjustmentID = data["adjustment-id"]
	settlement.ShipmentID = data["shipment-id"]
	settlement.MarketplaceName = data["marketplace-name"]
	settlement.AmountType = data["amount-type"]
	settlement.AmountDescription = data["amount-description"]
	settlement.FulfillmentID = data["fulfillment-id"]
	settlement.PostedDate = data["posted-date"]
	settlement.OrderItemCode = data["order-item-code"]
	settlement.MerchantOrderItemID = data["merchant-order-item-id"]
	settlement.MerchantAdjustmentItemID = data["merchant-adjustment-item-id"]
	settlement.SKU = data["sku"]

	// Parse numeric fields
	settlement.TotalAmount = parseSettlementFloat(data["total-amount"])
	settlement.Amount = parseSettlementFloat(data["amount"])

	if val := data["quantity-purchased"]; val != "" {
		settlement.QuantityPurchased, _ = strconv.Atoi(val)
	}

	// Parse date
	if dateStr := data["posted-date-time"]; dateStr != "" {
		var err error
		settlement.PostedDateTime, err = time.Parse("2006-01-02 15:04:05 MST", dateStr)
		if err != nil {
			settlement.PostedDateTime, err = time.Parse("2006-01-02 15:04:05", dateStr)
			if err != nil {
				settlement.PostedDateTime = time.Now()
			}
		}
	}

	// Store raw data as JSON
	rawData, _ := json.Marshal(data)
	settlement.RawData = string(rawData)

	return settlement, nil
}

// parseSettlementFloat safely parses a string to float64
func parseSettlementFloat(s string) float64 {
	if s == "" {
		return 0
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

// GetOrderTotal aggregates all amounts for a specific order ID
func AggregateSettlementsByOrderID(settlements []*Settlement) map[string]float64 {
	orderTotals := make(map[string]float64)

	for _, settlement := range settlements {
		if settlement.OrderID != "" {
			orderTotals[settlement.OrderID] += settlement.Amount
		}
	}

	return orderTotals
}
