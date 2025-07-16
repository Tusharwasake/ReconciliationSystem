package ingest

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Payment struct {
	ID                     int       `json:"id" db:"id"`
	OrderID                string    `json:"order_id" db:"order_id"`
	Date                   time.Time `json:"date" db:"date"`
	SettlementID           string    `json:"settlement_id" db:"settlement_id"`
	Type                   string    `json:"type" db:"type"`
	SKU                    string    `json:"sku" db:"sku"`
	Description            string    `json:"description" db:"description"`
	Quantity               int       `json:"quantity" db:"quantity"`
	Marketplace            string    `json:"marketplace" db:"marketplace"`
	AccountType            string    `json:"account_type" db:"account_type"`
	Fulfillment            string    `json:"fulfillment" db:"fulfillment"`
	TaxCollectionModel     string    `json:"tax_collection_model" db:"tax_collection_model"`
	ProductSales           float64   `json:"product_sales" db:"product_sales"`
	ProductSalesTax        float64   `json:"product_sales_tax" db:"product_sales_tax"`
	ShippingCredits        float64   `json:"shipping_credits" db:"shipping_credits"`
	ShippingCreditsTax     float64   `json:"shipping_credits_tax" db:"shipping_credits_tax"`
	GiftWrapCredits        float64   `json:"gift_wrap_credits" db:"gift_wrap_credits"`
	GiftwrapCreditsTax     float64   `json:"giftwrap_credits_tax" db:"giftwrap_credits_tax"`
	RegulatoryFee          float64   `json:"regulatory_fee" db:"regulatory_fee"`
	TaxOnRegulatoryFee     float64   `json:"tax_on_regulatory_fee" db:"tax_on_regulatory_fee"`
	PromotionalRebates     float64   `json:"promotional_rebates" db:"promotional_rebates"`
	PromotionalRebatesTax  float64   `json:"promotional_rebates_tax" db:"promotional_rebates_tax"`
	MarketplaceWithheldTax float64   `json:"marketplace_withheld_tax" db:"marketplace_withheld_tax"`
	SellingFees            float64   `json:"selling_fees" db:"selling_fees"`
	FBAFees                float64   `json:"fba_fees" db:"fba_fees"`
	OtherTransactionFees   float64   `json:"other_transaction_fees" db:"other_transaction_fees"`
	Other                  float64   `json:"other" db:"other"`
	Total                  float64   `json:"total" db:"total"`
	RawData                string    `json:"raw_data" db:"raw_data"`
}

// PaymentFromCSVRow creates a Payment from CSV row data
func PaymentFromCSVRow(headers []string, row []string) (*Payment, error) {
	if len(row) < len(headers) {
		return nil, fmt.Errorf("row has fewer fields than headers")
	}

	payment := &Payment{}

	// Create a map for easier field access
	data := make(map[string]string)

	for i, header := range headers {
		if i < len(row) {
			data[strings.TrimSpace(header)] = strings.TrimSpace(row[i])
		}
	}

	// Parse required fields
	payment.OrderID = data["order id"]
	payment.SettlementID = data["settlement id"]
	payment.Type = data["type"]
	payment.SKU = data["sku"]
	payment.Description = data["description"]
	payment.Marketplace = data["marketplace"]
	payment.AccountType = data["account type"]
	payment.Fulfillment = data["fulfillment"]
	payment.TaxCollectionModel = data["tax collection model"]

	// Parse numeric fields
	var err error

	if val := data["quantity"]; val != "" {
		payment.Quantity, _ = strconv.Atoi(val)
	}

	payment.ProductSales = parseFloat(data["product sales"])
	payment.ProductSalesTax = parseFloat(data["product sales tax"])
	payment.ShippingCredits = parseFloat(data["shipping credits"])
	payment.ShippingCreditsTax = parseFloat(data["shipping credits tax"])
	payment.GiftWrapCredits = parseFloat(data["gift wrap credits"])
	payment.GiftwrapCreditsTax = parseFloat(data["giftwrap credits tax"])
	payment.RegulatoryFee = parseFloat(data["Regulatory Fee"])
	payment.TaxOnRegulatoryFee = parseFloat(data["Tax On Regulatory Fee"])
	payment.PromotionalRebates = parseFloat(data["promotional rebates"])
	payment.PromotionalRebatesTax = parseFloat(data["promotional rebates tax"])
	payment.MarketplaceWithheldTax = parseFloat(data["marketplace withheld tax"])
	payment.SellingFees = parseFloat(data["selling fees"])
	payment.FBAFees = parseFloat(data["fba fees"])
	payment.OtherTransactionFees = parseFloat(data["other transaction fees"])
	payment.Other = parseFloat(data["other"])
	payment.Total = parseFloat(data["total"])

	// Parse date
	if dateStr := data["date/time"]; dateStr != "" {
		payment.Date, err = time.Parse("Jan 2, 2006 3:04:05 PM MST", dateStr)
		if err != nil {
			payment.Date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				payment.Date = time.Now()
			}
		}
	}

	// Store raw data as JSON
	rawData, _ := json.Marshal(data)
	payment.RawData = string(rawData)

	return payment, nil
}

// parseFloat safely parses a string to float64
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}
