package workflows

import (
	"context"
	"fmt"
	"time"

	"encore.app/billing/db"
	"encore.app/billing/models"
	"go.temporal.io/sdk/temporal"
)

func CreateBill(ctx context.Context, billCloseDate time.Time) (*models.Bill,error) {
	if billCloseDate.Before(time.Now()) {
		return nil,temporal.NewNonRetryableApplicationError("Invalid Bill Close Date", "INVALID-DATA",nil)
	}

	var bill models.Bill
	err := db.BillDb.QueryRow(ctx, `
	INSERT INTO bill
	(close_date)
	VALUES ($1)
	RETURNING id, status, close_date
	`,billCloseDate).Scan(&bill.BillId, &bill.Status, &bill.CloseDate)
	if err != nil {
		return nil,temporal.NewNonRetryableApplicationError(err.Error(),"DB-ERROR",nil)
	}
	fmt.Printf("Bill created successfully: %s\n", billCloseDate)

	return &bill, nil
}
func validateBillItem(amount int, currency string) error {
	if amount <= 0 {
		return temporal.NewNonRetryableApplicationError("Invalid amount "+ fmt.Sprint(amount), "INVALID-DATA",nil)
		}
	if currency != "USD" && currency != "GEP" {
		return temporal.NewNonRetryableApplicationError("Invalid currency: " + currency, "INVALID-DATA",nil)
	}
	return nil
}

func AddBillItem(ctx context.Context,billId string, amount int, currency string) error {
	err := validateBillItem(amount, currency)
	if err != nil {
		return err
	}
	_, err = db.BillDb.Exec(ctx, `
	INSERT INTO bill_item
	(bill_id,amount,currency)
	VALUES ($1,$2,$3)
	`, billId, amount, currency)

	if err != nil {
		return err
	}
	return nil
}

func GetBill(ctx context.Context, billId string) (*models.Bill, error) {
	var bill models.Bill
	err := db.BillDb.QueryRow(ctx, `
	SELECT id,status
	FROM bill
	WHERE bill.id = $1
	`,billId).Scan(&bill.BillId, &bill.Status)

	if err != nil {
		return nil, err
	}

	if bill.BillId == "" {
		return nil, temporal.NewNonRetryableApplicationError("Bill not found", "NOT_FOUND",nil)
	}

	rows, err := db.BillDb.Query(ctx,`
	SELECT id, currency, amount
	FROM bill_item
	where bill_item.bill_id = $1
	`,billId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var billItems []models.BillItem

	for rows.Next() {
		var item models.BillItem
		err := rows.Scan(&item.Id, &item.Currency, &item.Amount)
		if err != nil {
			return nil, err
		}
		billItems = append(billItems, item)
	}
	bill.BillItems = billItems

	return &bill, nil
}

func CloseBill(ctx context.Context, billId string) error {
	var bill models.Bill
	err := db.BillDb.QueryRow(ctx,`
	UPDATE bill
	SET 
    status = 'closed',
    closed_at = NOW()
	WHERE id = $1 AND status = 'open'
	RETURNING id, status
	`,billId).Scan(&bill.BillId, &bill.Status)
	if err != nil {
		return err
	}
	if bill.BillId == "" {
		return temporal.NewNonRetryableApplicationError("Bill not found","NOT_FOUND",nil)
	}
	return nil
}

func CheckOpenBill(ctx context.Context, billId string) (bool,error) {
	var bill models.Bill

	err := db.BillDb.QueryRow(ctx, `
	SELECT id,status
	FROM bill
	WHERE bill.id = $1
	`,billId).Scan(&bill.BillId, &bill.Status)

	if bill.BillId == "" {
		return false, temporal.NewNonRetryableApplicationError("Bill not found", "NOT_FOUND",nil)
	}

	if err != nil {
		return false, err
	}
	return bill.Status == "open", nil
}