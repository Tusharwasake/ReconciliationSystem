# Data Files

Put your data files here before running the reconciliation.

## What files to add:

- `payment_data.csv` - your payment transactions (CSV format)
- `settlement_data.txt` - your settlement data (tab-separated)

## How to run:

1. Copy your files to this folder
2. Run `go run main.go` from the main directory
3. Check the `output/` folder for results

## File formats:

Your payment CSV needs a header row with "date/time" in it. The settlement file should be tab-separated values.
