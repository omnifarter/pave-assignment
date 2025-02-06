// Service bill implements a simple bill world REST API.
package billing

import (
	"context"
	"fmt"

	"encore.app/billing/workflows"
	"encore.dev"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)
var (
    envName = encore.Meta().Environment.Name
    billingTaskQueue = envName + "-billing"
)

//encore:service
type Service struct {
	Client client.Client
	Worker worker.Worker
}

func initService() (*Service, error) {
	c, err := client.Dial(client.Options{})
	if err != nil {
		return nil, fmt.Errorf("create temporal client: %v", err)
	}

	w := worker.New(c, billingTaskQueue, worker.Options{})
	// Workflows
	w.RegisterWorkflow(workflows.ComposeBill)
	
	// Activities
	w.RegisterActivity(workflows.CloseBill)
	w.RegisterActivity(workflows.CreateBill)
	w.RegisterActivity(workflows.AddBillItem)
	w.RegisterActivity(workflows.GetBill)
	w.RegisterActivity(workflows.CheckOpenBill)

	err = w.Start()
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("start temporal worker: %v", err)
	}
	return &Service{Client: c, Worker: w}, nil
}

func (s *Service) Shutdown(force context.Context) {
	s.Client.Close()
	s.Worker.Stop()
}

