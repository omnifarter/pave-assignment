package models

import "time"

type BillItem struct {
	Id string `json:"id"`
	Amount int `json:"amount"`
	Currency string `json:"currency"`
}

type Bill struct {
	BillId string `json:"id"`
	CloseDate time.Time `json:"closeDate"`
	Status string `json:"status"`
	BillItems []BillItem
}

type BillSummary struct {
	BillId string `json:"id"`
	ClosedAt time.Time `json:"closedAt"`
	Status string `json:"status"`
	BillItems []BillItem `json:"billItems"`
	BillItemSummary []BillItemSummary `json:"billItemSummary"`
}

type BillItemSummary struct {
	BillId string `json:"billId"`
	TotalAmount int `json:"totalAmount"`
	Currency string `json:"currency"`
}
