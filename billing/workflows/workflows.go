package workflows

import (
	"strconv"
	"time"

	"encore.app/billing/models"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const UpdateBillItems = "update_bill_items"

func ComposeBill(ctx workflow.Context, bill *models.Bill) error {
    logger := workflow.GetLogger(ctx)

    options := workflow.ActivityOptions{
        StartToCloseTimeout: time.Second * 5,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
        },
        
    }

    ctx = workflow.WithActivityOptions(ctx, options)

    // Create update handler for updating bill with additional items
    err := workflow.SetUpdateHandler(ctx,UpdateBillItems, func(ctx workflow.Context, billItems []models.BillItem) error {
		logger.Info("Received update to add bill items.")
        ctx = workflow.WithActivityOptions(ctx, options)
        for _, billItem := range billItems {
            err := workflow.ExecuteActivity(ctx,AddBillItem, bill.BillId, billItem.Amount, billItem.Currency).Get(ctx,nil)
            if err != nil {
                logger.Error("failed to process a bill item: ", strconv.Itoa(billItem.Amount) + billItem.Currency)
                // probably send some notification or alert
            }
        }
        return nil
    },
)

    if err != nil {
        return temporal.NewNonRetryableApplicationError(err.Error(),"WORKFLOW_ERROR",nil)
    }

	selector := workflow.NewSelector(ctx)
    timerCtx, cancelHandler := workflow.WithCancel(ctx)

    // create timer future to close bill at close date.
    closeBillFuture := workflow.NewTimer(timerCtx, time.Until(bill.CloseDate))

    selector.AddFuture(closeBillFuture, func (f workflow.Future) {
        workflow.ExecuteActivity(ctx, CloseBill, bill.BillId).Get(ctx,nil)
    })

    // Listen for external signals (manual bill closure)
	signalChan := workflow.GetSignalChannel(ctx, "CLOSE_BILL")
	selector.AddReceive(signalChan, func(_ workflow.ReceiveChannel, _ bool) {
		logger.Info("Received signal to close bill early.")
        cancelHandler()
	})

	selector.Select(ctx)

    logger.Info("Bill workflow completed.")
    return nil 
}