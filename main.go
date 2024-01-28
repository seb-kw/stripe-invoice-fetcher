package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/balancetransaction"
	"github.com/stripe/stripe-go/v76/charge"
	"github.com/stripe/stripe-go/v76/invoice"
	"github.com/stripe/stripe-go/v76/payout"
)

type invoiceDownload struct {
	InvoiceID string
	PDFLink   string
}

func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func main() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		log.Fatal("Please set your STRIPE_SECRET_KEY environment variable.")
	}

	log.Println("Retrieving all payouts...")

	params := &stripe.PayoutListParams{}
	params.Limit = stripe.Int64(300)
	result := payout.List(params)

	// Show the complete object of payout po_1OcHNiDFRwBi1BohAG2amieq
	// Search for the payout with the id po_1OcHNiDFRwBi1BohAG2amieq
	for result.Next() {
		p := result.Payout()
		// store charge ids in this
		var chargeIDs []string
		var paymentIDs []string

		log.Printf("Retrieving Payout: %v\n", p.ID)
		log.Printf("Payout date: %v\n", time.Unix(p.ArrivalDate, 0).Format("2006-01-02"))

		log.Println("Retrieving Balance Transactions...")

		blParams := &stripe.BalanceTransactionListParams{}
		blParams.Limit = stripe.Int64(300)
		blParams.Payout = stripe.String(p.ID)

		result := balancetransaction.List(blParams)

		for result.Next() {
			if result.BalanceTransaction().Type == "charge" {
				chargeIDs = append(chargeIDs, result.BalanceTransaction().Source.ID)
			}
			if result.BalanceTransaction().Type == "payment" {
				paymentIDs = append(paymentIDs, result.BalanceTransaction().Source.ID)
			}
		}

		var chargeInvoiceIDs []string

		log.Println("Retrieving Charges...")
		// for each charge id, get the charge object
		for _, chargeID := range chargeIDs {

			params := &stripe.ChargeParams{}
			result, err := charge.Get(chargeID, params)

			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			chargeInvoiceIDs = append(chargeInvoiceIDs, result.Invoice.ID)
		}

		var paymentInvoiceIDs []string

		log.Println("Retrieving Payments...")
		// for each payment id, get the payment object
		for _, paymentID := range paymentIDs {

			params := &stripe.ChargeParams{}
			result, err := charge.Get(paymentID, params)

			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			paymentInvoiceIDs = append(paymentInvoiceIDs, result.Invoice.ID)
		}

		// merge the two invoice id arrays
		invoiceIDs := append(chargeInvoiceIDs, paymentInvoiceIDs...)

		var invoicePDFLinks []invoiceDownload

		log.Println("Retrieving Invoice PDF Links...")
		// for each invoice id, get the invoice object
		for _, invoiceID := range invoiceIDs {
			params := &stripe.InvoiceParams{}
			result, err := invoice.Get(invoiceID, params)

			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			invoicePDFLinks = append(invoicePDFLinks, invoiceDownload{InvoiceID: invoiceID, PDFLink: result.InvoicePDF})
		}

		// download the pdfs
		log.Println("Downloading PDFs...")
		// create folder
		folderName := time.Unix(p.ArrivalDate, 0).Format("2006-01-02") + "-" + p.ID
		dirErr := os.Mkdir(folderName, 0755)
		if dirErr != nil {
			fmt.Printf("Error: %v\n", dirErr)
		}

		for _, pdfLink := range invoicePDFLinks {
			// download pdf
			err := DownloadFile(folderName+"/"+pdfLink.InvoiceID+".pdf", pdfLink.PDFLink)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}

		log.Println("!!!!!!!!Done with this payout!!!!!!!!")

	}
}
