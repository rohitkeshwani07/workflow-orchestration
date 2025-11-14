package engine

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/workflow-orchestration/backend/database"
	"github.com/workflow-orchestration/backend/models"
)

type WorkflowEngine struct {
	db       *database.DB
	executor *NodeExecutor
}

func NewWorkflowEngine(db *database.DB, anthropicAPIKey string) *WorkflowEngine {
	return &WorkflowEngine{
		db:       db,
		executor: NewNodeExecutor(db, anthropicAPIKey),
	}
}

func (we *WorkflowEngine) ExecuteWorkflow(workflowID string, initialContext map[string]interface{}) (*models.Execution, error) {
	workflow, err := we.db.GetWorkflow(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}
	if workflow == nil {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	execution := &models.Execution{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		Status:     models.ExecutionStatusRunning,
		StartedAt:  time.Now(),
		Context:    initialContext,
	}

	// Initialize Context map if nil to prevent nil map assignment panic
	if execution.Context == nil {
		execution.Context = make(map[string]interface{})
	}

	if err := we.db.CreateExecution(execution); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// Find trigger node
	var triggerNode *models.Node
	for i := range workflow.Nodes {
		if workflow.Nodes[i].Type == models.NodeTypeChatTrigger {
			triggerNode = &workflow.Nodes[i]
			break
		}
	}

	if triggerNode == nil {
		execution.Status = models.ExecutionStatusError
		errMsg := "no trigger node found in workflow"
		execution.Error = &errMsg
		now := time.Now()
		execution.FinishedAt = &now
		we.db.UpdateExecution(execution)
		return execution, fmt.Errorf(errMsg)
	}

	// Execute workflow
	if err := we.executeNode(triggerNode, workflow, execution); err != nil {
		execution.Status = models.ExecutionStatusError
		errMsg := err.Error()
		execution.Error = &errMsg
		now := time.Now()
		execution.FinishedAt = &now
	} else {
		execution.Status = models.ExecutionStatusSuccess
		now := time.Now()
		execution.FinishedAt = &now
	}

	we.db.UpdateExecution(execution)
	return execution, nil
}

func (we *WorkflowEngine) ExecuteFromTrigger(workflowID string, triggerData map[string]interface{}) (*models.Execution, error) {
	return we.ExecuteWorkflow(workflowID, map[string]interface{}{
		"trigger": triggerData,
	})
}

func (we *WorkflowEngine) executeNode(node *models.Node, workflow *models.Workflow, execution *models.Execution) error {
	// Execute the node
	result, err := we.executor.Execute(node, execution.Context)
	if err != nil {
		errMsg := err.Error()
		we.db.AddExecutionLog(execution.ID, node.ID, "error", nil, &errMsg)
		return err
	}

	// Log successful execution
	resultJSON, _ := json.Marshal(result)
	resultStr := string(resultJSON)
	we.db.AddExecutionLog(execution.ID, node.ID, "success", &resultStr, nil)

	// Update context
	execution.Context[node.ID] = result

	// Find and execute next nodes
	nextNodes := we.getNextNodes(node.ID, workflow)
	for _, nextNode := range nextNodes {
		if err := we.executeNode(nextNode, workflow, execution); err != nil {
			return err
		}
	}

	return nil
}

func (we *WorkflowEngine) getNextNodes(nodeID string, workflow *models.Workflow) []*models.Node {
	var nextNodes []*models.Node

	// Find outgoing edges
	for _, edge := range workflow.Edges {
		if edge.Source == nodeID {
			// Find target node
			for i := range workflow.Nodes {
				if workflow.Nodes[i].ID == edge.Target {
					nextNodes = append(nextNodes, &workflow.Nodes[i])
					break
				}
			}
		}
	}

	return nextNodes
}
