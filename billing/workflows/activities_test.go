package workflows

import (
	"context"
	"testing"
	"time"

	"encore.app/billing/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/testsuite"
)

func TestActivity_CreateBill(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(CreateBill)

	var bill models.Bill
	val, err := env.ExecuteActivity(CreateBill, time.Now().Add(24 * time.Hour))
	require.NoError(t, val.Get(&bill))
	require.NoError(t, err)
}

func TestActivity_CreateBill_InvalidDate(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(CreateBill)

	val, err := env.ExecuteActivity(CreateBill, time.Now().Add(-24 * time.Hour))
	require.Error(t, err)
	require.Empty(t, val)
}

func TestActivity_AddBillItem(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(AddBillItem)
	bill, _ := CreateBill(context.Background(), time.Now().Add(24 * time.Hour))
	_, err := env.ExecuteActivity(AddBillItem,bill.BillId, 100, "USD")
	require.NoError(t, err)
}

func TestActivity_AddBillItem_InvalidAmount(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(AddBillItem)
	bill, _ := CreateBill(context.Background(), time.Now().Add(24 * time.Hour))
	_, err := env.ExecuteActivity(AddBillItem, bill.BillId, -100, "USD")
	require.Error(t, err)
}

func TestActivity_AddBillItem_InvalidCurrency(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(AddBillItem)
	bill, _ := CreateBill(context.Background(), time.Now().Add(24 * time.Hour))
	_, err := env.ExecuteActivity(AddBillItem, bill.BillId, 100, "ABC")
	require.Error(t, err)
}

func TestActivity_GetBill(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(GetBill)
	bill, _ := CreateBill(context.Background(), time.Now().Add(24 * time.Hour))
	_, err := env.ExecuteActivity(GetBill, bill.BillId)
	require.NoError(t, err)
}
func TestActivity_GetBill_Invalid(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(GetBill)
	_, err := env.ExecuteActivity(GetBill, uuid.New().String())
	require.Error(t, err)
}

func TestActivity_CloseBill(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(CloseBill)
	bill, _ := CreateBill(context.Background(), time.Now().Add(24 * time.Hour))
	require.Equal(t, bill.Status, "open")
	_, err := env.ExecuteActivity(CloseBill, bill.BillId)
	require.NoError(t, err)
	bill, _ = GetBill(context.Background(),bill.BillId)
	require.Equal(t, bill.Status, "closed")
}

func TestActivity_CloseBill_Invalid(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(CloseBill)
	_, err := env.ExecuteActivity(CloseBill, uuid.New().String())
	require.Error(t, err)
}
func TestActivity_GetBillSummary(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(GetBillSummary)
	bill, _ := CreateBill(context.Background(), time.Now().Add(24 * time.Hour))
	 _  = CloseBill(context.Background(), bill.BillId)
	_, err := env.ExecuteActivity(GetBillSummary, bill.BillId)
	require.NoError(t, err)
}

func TestActivity_GetBillSummary_InvalidOpenBill(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(GetBillSummary)
	bill, _ := CreateBill(context.Background(), time.Now().Add(24 * time.Hour))
	_, err := env.ExecuteActivity(GetBillSummary, bill.BillId)
	require.Error(t, err)
}