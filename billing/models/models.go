package models

import "time"

type BillItem struct {
	Id string `json:"id"`
	Amount int `json:"amount"`
	Currency string `json:"currency"`
}

type Bill struct {
	BillId string `json:"id"`
	CloseDate time.Time `json:"close_date"`
	Status string `json:"status"`
	BillItems []BillItem
}
