// Service bill implements a simple bill world REST API.
package billing

import (
	"context"
	"time"

	"encore.app/billing/models"
	"encore.app/billing/workflows"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"go.temporal.io/sdk/client"
)

type Response struct {
	Message string
}

type CreateBillResponse struct {
	BillId string `json:"billId"`
}

//encore:api private path=/bill/:billId
func (s *Service) GetBill(ctx context.Context, billId string) (*models.Bill, error) {
	bill,err := workflows.GetBill(ctx, billId)
	if err != nil {
        return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: err.Error(),
		}
    }

	return bill, nil
}



type CreateBillRequest struct {
	CloseDate time.Time `json:"CloseDate"`
}

//encore:api private method=POST path=/bill
func (s *Service) CreateBill(ctx context.Context, createBillRequest CreateBillRequest) (*CreateBillResponse, error) {
	bill, err := workflows.CreateBill(ctx,createBillRequest.CloseDate)
	if err != nil {
		return nil, err
	}
	options := client.StartWorkflowOptions{
        ID:        bill.BillId,
        TaskQueue: billingTaskQueue,
    }
	s.Client.ExecuteWorkflow(ctx, options, workflows.ComposeBill, bill)
	return &CreateBillResponse{BillId: bill.BillId}, nil
}



type AddBillItemsRequest struct {
	BillItems []models.BillItem `json:"billItems"`
}

//encore:api private method=POST path=/bill/:billId/items
func (s *Service) AddBillItems(ctx context.Context, billId string, billItems AddBillItemsRequest) (*Response, error) {
	rlog.Info("Bill ID" + billId)
	isOpen, err := workflows.CheckOpenBill(ctx,billId)
	if err != nil {
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: err.Error(),
		}
	}
	if !isOpen {
		return nil, &errs.Error{
			Code: errs.FailedPrecondition,
			Message: "Bill is already closed.",
		}
	}

	updateHandle , err := s.Client.UpdateWorkflow(context.Background(),client.UpdateWorkflowOptions{
		WorkflowID: billId,
		UpdateName: workflows.UpdateBillItems,
		Args: []interface{}{billItems.BillItems},
		WaitForStage: client.WorkflowUpdateStageCompleted,
	})

	if err != nil {
		rlog.Error("Failed to update bill", billId)
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: err.Error(),
		}
	}

	var updateResult string
	err = updateHandle.Get(context.Background(),&updateResult)

	if err != nil {
		rlog.Error("Failed to update bill", billId)
		return nil, & errs.Error{
			Code: errs.InvalidArgument,
			Message: err.Error(),
		}
	}

	return &Response{Message: "Bill items added."}, nil
}



//encore:api private path=/bill/:billId/close
func (s *Service) CloseBill(ctx context.Context, billId string) (*Response, error) {
	err := s.Client.SignalWorkflow(ctx, billId, "", "CLOSE_BILL",nil)
	if err != nil {
        return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: err.Error(),
		}
    }
	return  &Response{Message: "Bill closed"}, nil
}

//encore:api private path=/bill/:billId/summary
func (s *Service) GetBillSummary(ctx context.Context, billId string) (*models.BillSummary, error) {
	isOpen, err := workflows.CheckOpenBill(ctx,billId)
	if err != nil {
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: err.Error(),
		}
	}
	if isOpen {
		return nil, &errs.Error{
			Code: errs.FailedPrecondition,
			Message: "Bill is still open.",
		}
	}
	billSummary, err := workflows.GetBillSummary(ctx, billId)
	if err != nil {
        return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: err.Error(),
		}
    }
	return  billSummary, nil
}