package workflows

import (
	"fmt"
	"testing"
	"time"

	"encore.app/billing/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/testsuite"
)

func TestWorkflow_CloseBill(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock CloseBill activity
	env.OnActivity(CloseBill, mock.Anything, "TEST_BILL").Return(nil)

	env.ExecuteWorkflow(ComposeBill, &models.Bill{BillId: "TEST_BILL", CloseDate: time.Now().Add(24 * time.Hour)})

	env.SignalWorkflow("CLOSE_BILL", nil)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertActivityCalled(t,"CloseBill",mock.Anything, "TEST_BILL")
}
func TestWorkflow_CloseBill_Signal(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock CloseBill activity
	env.OnActivity(CloseBill, mock.Anything, "TEST_BILL").Return(nil)

	env.ExecuteWorkflow(ComposeBill, &models.Bill{BillId: "TEST_BILL", CloseDate: time.Now().Add(24 * time.Hour)})

	env.SignalWorkflow("CLOSE_BILL", nil)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertActivityNotCalled(t,"CloseBill")
}

func TestWorkflow_AddBillItems(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock CloseBill activity
	env.OnActivity(CloseBill, mock.Anything, "TEST_BILL").Return(nil)
	// Mock AddBillItem activity
	env.OnActivity(AddBillItem, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	var items [2]models.BillItem
	items[0] = models.BillItem{}
	items[1] = models.BillItem{}
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow("update_bill_items","",&testsuite.TestUpdateCallback{
				OnAccept: func() {},
				OnReject: func(err error) {
					require.Fail(t, "unexpected rejection")
				},
				OnComplete: func(i interface{}, err error) {
					require.NoError(t, err)
				},
			},items)
	},0)

	env.ExecuteWorkflow(ComposeBill, &models.Bill{BillId: "TEST_BILL", CloseDate: time.Now().Add(24 * time.Hour)})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	env.AssertActivityCalled(t,"CloseBill",mock.Anything, "TEST_BILL")
	env.AssertActivityNumberOfCalls(t,"AddBillItem",2)
}

func TestWorkflow_AddBillItemsError(t *testing.T) {
	const RETRY_COUNT = 3

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock CloseBill activity
	env.OnActivity(CloseBill, mock.Anything, "TEST_BILL").Return(nil)
	// Mock AddBillItem activity
	env.OnActivity(AddBillItem, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("crash"))

	var items [2]models.BillItem
	items[0] = models.BillItem{}
	items[1] = models.BillItem{}
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow("update_bill_items","",&testsuite.TestUpdateCallback{
				OnAccept: func() {},
				OnReject: func(err error) {
					require.Fail(t, "unexpected rejection")
				},
				OnComplete: func(i interface{}, err error) {
					require.NoError(t, err)
				},
			},items)
	},0)

	env.ExecuteWorkflow(ComposeBill, &models.Bill{BillId: "TEST_BILL", CloseDate: time.Now().Add(24 * time.Hour)})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	env.AssertActivityCalled(t,"CloseBill",mock.Anything, "TEST_BILL")
	
	env.AssertActivityNumberOfCalls(t,"AddBillItem", len(items) * RETRY_COUNT)
}
