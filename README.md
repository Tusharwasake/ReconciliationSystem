# Payment Reconciliation System

A simple Go-based tool for reconciling payment and settlement data files. The system processes CSV payment files and TSV settlement files, stores them in a PostgreSQL database, and generates reconciliation reports.

## Features

- **Multi-format File Support**: Handles CSV payment files and TSV settlement files
- **Flexible Header Detection**: Automatically detects headers in payment CSV files
- **Data Aggregation**: Aggregates settlement data by order ID for accurate reconciliation
- **Comprehensive Reporting**: Generates detailed reconciliation reports with status indicators
- **Database Integration**: Uses PostgreSQL with proper indexing for efficient data operations
- **Error Handling**: Robust error handling for malformed data and missing files

## System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  File Parsers   │    │  Database       │    │  Reconciliation │
│  (CSV/TSV)      │───▶│  Storage        │───▶│  Engine         │
│                 │    │  (PostgreSQL)   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
                                            ┌─────────────────┐
                                            │  CSV Report     │
                                            │  Generator      │
                                            └─────────────────┘
```

## Data Processing Flow

1. **File Ingestion**:

   - Payment CSV files are parsed with automatic header detection
   - Settlement TSV files are processed with tab-separated values
   - Data is validated and cleaned during ingestion

2. **Database Storage**:

   - Both payment and settlement records are stored in a unified `records` table
   - Raw data is preserved for audit purposes
   - Proper indexing ensures efficient queries

3. **Reconciliation Process**:

   - Matches payment and settlement records by order ID
   - Calculates amount differences
   - Stores reconciliation results in `reconciled_records` table

4. **Report Generation**:
   - Generates CSV reports with reconciliation status
   - Includes payment totals, settlement totals, and differences

## Setup

### Prerequisites

- Go 1.24.4 or later
- PostgreSQL 12 or later

### Installation

1. Clone the repository:

```bash
git clone <repository-url>
cd PortOneReconciliation
```

2. Install dependencies:

```bash
go mod tidy
```

3. Create PostgreSQL database:

```bash
createdb portdb
```

4. Configure environment variables (optional):

```bash
# Create .env file or set environment variables
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=portdb
DB_SSLMODE=disable
```

## Usage

### Basic Usage

1. **Add your data files** to the `data/` folder:

   - Copy your payment CSV to: `data/payment_data.csv`
   - Copy your settlement file to: `data/settlement_data.txt`

2. **Run the reconciliation**:

```bash
go run main.go
```

The system will:

1. Connect to the PostgreSQL database
2. Run database migrations (create tables and indexes)
3. Clear existing data and ingest new files
4. Process payment and settlement data
5. Perform reconciliation matching
6. Generate a CSV report in the `output/` directory

### File Processing Details

#### Payment File Processing

- Automatically detects CSV headers by scanning the first 20 lines
- Looks for the line containing "date/time" as the header row
- Parses various payment fields including totals, fees, and metadata
- Handles different date formats gracefully
- Skips invalid or incomplete records

#### Settlement File Processing

- Processes TSV (tab-separated values) files
- Aggregates settlement amounts by order ID
- Preserves original settlement data for audit purposes
- Handles multiple settlement entries per order

## File Formats

### Payment Data (CSV)

Expected format with automatic header detection:

```csv
date/time,settlement id,type,order id,sku,description,quantity,marketplace,account type,fulfillment,tax collection model,product sales,product sales tax,shipping credits,shipping credits tax,gift wrap credits,giftwrap credits tax,Regulatory Fee,Tax On Regulatory Fee,promotional rebates,promotional rebates tax,marketplace withheld tax,selling fees,fba fees,other transaction fees,other,total
Jan 15, 2024 10:30:00 AM PST,12345,Order,ORD001,SKU123,Product Description,1,Amazon,Seller,MFN,MarketplaceFacilitator,100.00,8.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,15.00,0.00,0.00,0.00,93.00
```

### Settlement Data (TSV)

Expected format with tab-separated values:

```
settlement id	settlement start date	settlement end date	deposit date	total amount	currency	transaction type	order id	merchant order id	adjustment id	shipment id	marketplace name	amount type	amount description	amount	fulfillment id	posted date	posted date time	order item code	merchant order item id	merchant adjustment item id	sku	quantity purchased
12345	2024-01-01	2024-01-15	2024-01-16	93.00	USD	Order	ORD001	MORD001	ADJ001	SHIP001	Amazon	ItemPrice	Principal	100.00	FUL001	2024-01-15	2024-01-15 10:30:00	ITEM001	MITEM001	MADJ001	SKU123	1
```

## Output

### Reconciliation Report

The system generates `output/reconciliation_report.csv` with the following columns:

| Column            | Description                                |
| ----------------- | ------------------------------------------ |
| order_id          | Unique order identifier                    |
| status            | "reconciled" or "unreconciled"             |
| payments_total    | Total amount from payment data             |
| settlements_total | Total amount from settlement data          |
| difference        | Amount difference (payments - settlements) |

Example output:

```csv
order_id,status,payments_total,settlements_total,difference
ORD001,reconciled,100.00,100.00,0.00
ORD002,unreconciled,150.00,145.00,5.00
```

## Database Schema

### Tables

#### `records` Table

Stores both payment and settlement data:

```sql
CREATE TABLE records (
    id SERIAL PRIMARY KEY,
    source VARCHAR(20) NOT NULL CHECK (source IN ('payments', 'settlements')),
    order_id VARCHAR(255) NOT NULL,
    date TIMESTAMP NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    raw_data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### `reconciled_records` Table

Stores reconciliation results:

```sql
CREATE TABLE reconciled_records (
    id SERIAL PRIMARY KEY,
    payments_record_id INTEGER REFERENCES records(id),
    settlements_record_id INTEGER REFERENCES records(id),
    amount_difference DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Indexes

- `idx_records_source`: Optimizes queries by source type
- `idx_records_order_id`: Optimizes order ID lookups
- `idx_reconciled_payments`: Optimizes payment record joins
- `idx_reconciled_settlements`: Optimizes settlement record joins

## Configuration

### Environment Variables

The application supports the following environment variables:

| Variable    | Default   | Description                      |
| ----------- | --------- | -------------------------------- |
| DB_HOST     | localhost | Database host                    |
| DB_PORT     | 5432      | Database port                    |
| DB_USER     | postgres  | Database user                    |
| DB_PASSWORD | 123456    | Database password                |
| DB_NAME     | portdb    | Database name                    |
| DB_SSLMODE  | disable   | SSL mode for database connection |

### Database Connection

The system uses `sqlx` for database operations with automatic connection management and proper error handling.

## Error Handling

The system includes comprehensive error handling for:

- Missing or corrupted files
- Invalid data formats
- Database connection issues
- Data validation errors
- File permission problems

Errors are logged with descriptive messages to help with troubleshooting.

## Testing

### Test Data Files

The repository includes sample test files:

- `test_payment_data.csv`: Sample payment data
- `test_settlement_data.txt`: Sample settlement data

## Dependencies

The project uses the following Go modules:

- **[github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx)** - Enhanced database operations with struct mapping
- **[github.com/lib/pq](https://github.com/lib/pq)** - PostgreSQL driver for Go
- **[github.com/joho/godotenv](https://github.com/joho/godotenv)** - Environment variable management from .env files

## Project Structure

```
PortOneReconciliation/
├── main.go                     # Application entry point
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── schema.sql                  # Database schema and indexes
├── .env                        # Environment configuration (optional)
├── data/                       # Data directory (user files)
│   ├── README.md              # Data directory instructions
│   ├── payment_data.csv       # User payment data (place here)
│   └── settlement_data.txt    # User settlement data (place here)
├── config/
│   ├── db.go                   # Database connection configuration
│   └── migration.go            # Database migration runner
├── controllers/
│   ├── ingest_controller.go    # File ingestion orchestration
│   └── reconcile_controller.go # Reconciliation logic
├── ingest/
│   ├── payment.go              # Payment data structures and parsing
│   └── settlements.go          # Settlement data structures and parsing
├── models/
│   └── record.go               # Database record models
├── utils/
│   └── parser.go               # File parsing utilities
├── views/
│   └── report_view.go          # CSV report generation
└── output/
    └── reconciliation_report.csv # Generated reconciliation report
```

## Development

### Code Organization

- **`main.go`**: Orchestrates the entire reconciliation process
- **`config/`**: Database configuration and migration management
- **`controllers/`**: Business logic for ingestion and reconciliation
- **`ingest/`**: Data parsing and transformation logic
- **`utils/`**: Utility functions for file processing
- **`views/`**: Report generation and output formatting

### Key Components

#### File Parsers (`utils/parser.go`)

- **`ParseAndStorePayments()`**: Handles CSV payment file processing
- **`ParseAndStoreSettlements()`**: Handles TSV settlement file processing
- Smart header detection for CSV files
- Robust error handling for malformed data

#### Data Models (`ingest/`)

- **`Payment`**: Comprehensive payment data structure
- **`Settlement`**: Detailed settlement data structure
- JSON serialization for raw data preservation
- Type-safe field mapping from CSV/TSV

#### Database Layer (`config/`)

- Connection management with environment variable support
- Automatic schema migration on startup
- Proper connection pooling and error handling

#### Reconciliation Engine (`controllers/reconcile_controller.go`)

- SQL-based matching of payment and settlement records
- Amount difference calculation
- Efficient database operations with proper indexing

## Troubleshooting

### Common Issues

1. **Database Connection Errors**

   - Verify PostgreSQL is running
   - Check database credentials in environment variables
   - Ensure database exists: `createdb portdb`

2. **File Not Found Errors**

   - Ensure data files are in the root directory
   - Check file names match exactly: `test_payment_data.csv` and `test_settlement_data.txt`

3. **Permission Errors**

   - Verify read permissions on input files
   - Ensure write permissions for `output/` directory

4. **Data Format Issues**
   - Check CSV headers contain "date/time" field
   - Verify TSV files use tab separators
   - Validate data completeness and format

### Debug Mode

Add logging to troubleshoot issues:

```go
import "log"

// Add before file processing
log.Printf("Processing file: %s", filePath)

// Add after database operations
log.Printf("Records processed: %d", recordsProcessed)
```

### Performance Considerations

- For large files (>10MB), consider adding progress indicators
- Monitor memory usage during processing
- Database indexes are optimized for typical reconciliation queries
- Consider connection pooling for high-volume processing

## Areas of Improvement

## Areas of Improvement

For processing large data files efficiently, here are key optimizations needed:

### 1. **Batch Database Operations**

Current code inserts records one by one, which is slow for large files.

```go
// Current - slow for large files
config.DB.Exec(`INSERT INTO records (source, order_id, date, total_amount, raw_data)
    VALUES ($1, $2, $3, $4, $5)`, ...)
```

**Better approach**: Batch inserts

```go
// Process in batches of 1000 records
tx, _ := db.Begin()
stmt, _ := tx.Prepare("INSERT INTO records ...")
for i, payment := range payments {
    stmt.Exec(payment.Data...)
    if i%1000 == 0 {
        tx.Commit()
        tx, _ = db.Begin()
    }
}
tx.Commit()
```

### 2. **Memory Management**

Current code loads all settlement data into memory at once.

```go
// Current - uses too much memory
var settlements []*ingest.Settlement
for scanner.Scan() {
    settlements = append(settlements, settlement)
}
```

**Better approach**: Process in chunks

```go
// Process 1000 records at a time
const ChunkSize = 1000
chunk := make([]*ingest.Settlement, 0, ChunkSize)
for scanner.Scan() {
    chunk = append(chunk, settlement)
    if len(chunk) >= ChunkSize {
        processChunk(chunk)
        chunk = chunk[:0]
    }
}
```

### 3. **File Processing**

Current code processes one record at a time sequentially.

**Better approach**: Use worker pools

```go
// Process multiple records concurrently
jobs := make(chan []string, 100)
results := make(chan *Payment, 100)

// Start 4 workers
for i := 0; i < 4; i++ {
    go worker(jobs, results)
}
```

### 4. **Progress Tracking**

Add progress indicators for large files:

```go
type ProgressTracker struct {
    total, processed int64
    startTime time.Time
}

func (p *ProgressTracker) Update(processed int64) {
    percentage := float64(processed) / float64(p.total) * 100
    fmt.Printf("Progress: %.1f%% (%d/%d)\n", percentage, processed, p.total)
}
```

### 5. **Configuration for Large Files**

Add environment variables for tuning:

```bash
# In .env file
BATCH_SIZE=1000
WORKER_COUNT=4
MEMORY_LIMIT=500MB
```

These improvements would handle files with millions of records efficiently without running out of memory.

## Future Enhancements

Building on the improvements above, potential future enhancements include:

- **Batch Processing**: Add support for processing multiple files in batches
- **API Interface**: REST API for programmatic access
- **Real-time Processing**: WebSocket-based live updates
- **Export Formats**: Support for multiple output formats (JSON, XML)
- **Audit Trail**: Comprehensive logging and audit capabilities
- **Scheduling**: Automated processing with cron-like scheduling
- **Monitoring**: Health checks and metrics collection

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and add tests
4. Commit your changes: `git commit -m 'Add feature'`
5. Push to the branch: `git push origin feature-name`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:

1. Check the troubleshooting section above
2. Review the code documentation
3. Create an issue in the repository
4. Provide sample data and error logs for faster resolution
