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
	BillId string `json:"bill_id"`
}

//encore:api private path=/bill/:bill_id
func (s *Service) GetBill(ctx context.Context, bill_id string) (*models.Bill, error) {
	bill,err := workflows.GetBill(ctx, bill_id)
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

//encore:api private method=POST path=/bill/:bill_id/items
func (s *Service) AddBillItems(ctx context.Context, bill_id string, billItems AddBillItemsRequest) (*Response, error) {
	rlog.Info("Bill ID" + bill_id)
	isOpen, err := workflows.CheckOpenBill(ctx,bill_id)
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
		WorkflowID: bill_id,
		UpdateName: workflows.UpdateBillItems,
		Args: []interface{}{billItems.BillItems},
		WaitForStage: client.WorkflowUpdateStageCompleted,
	})

	if err != nil {
		rlog.Error("Failed to update bill", bill_id)
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: err.Error(),
		}
	}

	var updateResult string
	err = updateHandle.Get(context.Background(),&updateResult)

	if err != nil {
		rlog.Error("Failed to update bill", bill_id)
		return nil, & errs.Error{
			Code: errs.InvalidArgument,
			Message: err.Error(),
		}
	}

	return &Response{Message: "Bill items added."}, nil
}



//encore:api private path=/bill/:bill_id/close
func (s *Service) CloseBill(ctx context.Context, bill_id string) (*Response, error) {
	err := s.Client.SignalWorkflow(ctx, bill_id, "", "CLOSE_BILL",nil)
	if err != nil {
        return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: err.Error(),
		}
    }
	return  &Response{Message: "Bill closed"}, nil
}