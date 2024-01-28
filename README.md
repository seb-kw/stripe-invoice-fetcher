# Stripe Invoice Fetcher

This is a simple script that fetches all invoices from all payouts of your Stripe Account and saves them as PDF files grouped by payout.

## Installation

1. Clone this repository
2. Set your Stripe API key as environment variable `STRIPE_API_KEY`, e.g. `export STRIPE_API_KEY=sk_test_1234567890`
3. Run `go run main.go`

## Output

The script will create a folder with the name scheme `<Payout Arrival Date formatted as YYYY-MM-DD>-<Payout ID>` for each payout. The invoices will be saved as PDF files in this folder as `<Invoice ID>.pdf`