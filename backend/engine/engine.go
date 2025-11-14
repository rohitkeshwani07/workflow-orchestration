package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/workflow-orchestration/backend/database"
	"github.com/workflow-orchestration/backend/logger"
	"github.com/workflow-orchestration/backend/models"
)

type WorkflowEngine struct {
	db       *database.DB
	executor *NodeExecutor
	logger   *logger.ClickHouseLogger
}

func NewWorkflowEngine(db *database.DB, anthropicAPIKey string) *WorkflowEngine {
	// Initialize ClickHouse logger
	chLogger, err := logger.NewClickHouseLogger()
	if err != nil {
		log.Printf("Warning: Failed to initialize ClickHouse logger: %v", err)
		// Continue without ClickHouse logging if it fails
		chLogger = nil
	}

	return &WorkflowEngine{
		db:       db,
		executor: NewNodeExecutor(db, anthropicAPIKey),
		logger:   chLogger,
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

	startTime := time.Now()
	execution := &models.Execution{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		Status:     models.ExecutionStatusRunning,
		StartedAt:  startTime,
		Context:    initialContext,
	}

	// Initialize Context map if nil to prevent nil map assignment panic
	if execution.Context == nil {
		execution.Context = make(map[string]interface{})
	}

	if err := we.db.CreateExecution(execution); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// Log workflow execution start to ClickHouse
	if we.logger != nil {
		ctx := context.Background()
		we.logger.LogExecution(ctx, logger.ExecutionLog{
			ExecutionID: execution.ID,
			WorkflowID:  workflowID,
			Timestamp:   startTime,
			Level:       "INFO",
			Message:     "Workflow execution started",
			Status:      "RUNNING",
			Metadata:    map[string]interface{}{"workflow_name": workflow.Name},
		})

		we.logger.LogWorkflowExecution(ctx, logger.WorkflowExecution{
			ExecutionID: execution.ID,
			WorkflowID:  workflowID,
			StartedAt:   startTime,
			Status:      "RUNNING",
			TriggerType: "manual",
			TotalNodes:  uint16(len(workflow.Nodes)),
		})
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
	var nodeStats struct {
		successful uint16
		failed     uint16
		skipped    uint16
	}

	if err := we.executeNode(triggerNode, workflow, execution, &nodeStats); err != nil {
		execution.Status = models.ExecutionStatusError
		errMsg := err.Error()
		execution.Error = &errMsg
		now := time.Now()
		execution.FinishedAt = &now

		// Log workflow execution failure to ClickHouse
		if we.logger != nil {
			duration := uint32(time.Since(startTime).Milliseconds())
			ctx := context.Background()

			we.logger.LogExecution(ctx, logger.ExecutionLog{
				ExecutionID: execution.ID,
				WorkflowID:  workflowID,
				Timestamp:   now,
				Level:       "ERROR",
				Message:     "Workflow execution failed",
				Status:      "FAILED",
				Error:       errMsg,
				DurationMs:  duration,
			})

			we.logger.LogWorkflowExecution(ctx, logger.WorkflowExecution{
				ExecutionID:     execution.ID,
				WorkflowID:      workflowID,
				StartedAt:       startTime,
				CompletedAt:     now,
				Status:          "FAILED",
				TriggerType:     "manual",
				TotalNodes:      uint16(len(workflow.Nodes)),
				SuccessfulNodes: nodeStats.successful,
				FailedNodes:     nodeStats.failed,
				SkippedNodes:    nodeStats.skipped,
				DurationMs:      duration,
				Error:           errMsg,
			})
		}
	} else {
		execution.Status = models.ExecutionStatusSuccess
		now := time.Now()
		execution.FinishedAt = &now

		// Log workflow execution success to ClickHouse
		if we.logger != nil {
			duration := uint32(time.Since(startTime).Milliseconds())
			ctx := context.Background()

			we.logger.LogExecution(ctx, logger.ExecutionLog{
				ExecutionID: execution.ID,
				WorkflowID:  workflowID,
				Timestamp:   now,
				Level:       "INFO",
				Message:     "Workflow execution completed successfully",
				Status:      "SUCCESS",
				DurationMs:  duration,
			})

			we.logger.LogWorkflowExecution(ctx, logger.WorkflowExecution{
				ExecutionID:     execution.ID,
				WorkflowID:      workflowID,
				StartedAt:       startTime,
				CompletedAt:     now,
				Status:          "SUCCESS",
				TriggerType:     "manual",
				TotalNodes:      uint16(len(workflow.Nodes)),
				SuccessfulNodes: nodeStats.successful,
				FailedNodes:     nodeStats.failed,
				SkippedNodes:    nodeStats.skipped,
				DurationMs:      duration,
			})
		}
	}

	we.db.UpdateExecution(execution)
	return execution, nil
}

func (we *WorkflowEngine) ExecuteFromTrigger(workflowID string, triggerData map[string]interface{}) (*models.Execution, error) {
	return we.ExecuteWorkflow(workflowID, map[string]interface{}{
		"trigger": triggerData,
	})
}

func (we *WorkflowEngine) executeNode(node *models.Node, workflow *models.Workflow, execution *models.Execution, stats *struct {
	successful uint16
	failed     uint16
	skipped    uint16
}) error {
	startTime := time.Now()
	ctx := context.Background()

	// Log node execution start to ClickHouse
	if we.logger != nil {
		we.logger.LogExecution(ctx, logger.ExecutionLog{
			ExecutionID: execution.ID,
			WorkflowID:  execution.WorkflowID,
			NodeID:      node.ID,
			Timestamp:   startTime,
			Level:       "INFO",
			Message:     fmt.Sprintf("Executing node: %s", node.Type),
			Status:      "RUNNING",
			Metadata:    map[string]interface{}{"node_type": node.Type},
		})
	}

	// Execute the node
	result, err := we.executor.Execute(node, execution.Context)
	completedAt := time.Now()
	duration := uint32(completedAt.Sub(startTime).Milliseconds())

	if err != nil {
		errMsg := err.Error()
		we.db.AddExecutionLog(execution.ID, node.ID, "error", nil, &errMsg)

		// Log node execution failure to ClickHouse
		if we.logger != nil {
			we.logger.LogExecution(ctx, logger.ExecutionLog{
				ExecutionID: execution.ID,
				WorkflowID:  execution.WorkflowID,
				NodeID:      node.ID,
				Timestamp:   completedAt,
				Level:       "ERROR",
				Message:     fmt.Sprintf("Node execution failed: %s", node.Type),
				Status:      "FAILED",
				Error:       errMsg,
				DurationMs:  duration,
			})

			we.logger.LogNodeExecution(ctx, logger.NodeExecution{
				ExecutionID: execution.ID,
				WorkflowID:  execution.WorkflowID,
				NodeID:      node.ID,
				StartedAt:   startTime,
				CompletedAt: completedAt,
				Status:      "FAILED",
				NodeType:    string(node.Type),
				Input:       execution.Context,
				Error:       errMsg,
				DurationMs:  duration,
			})
		}

		if stats != nil {
			stats.failed++
		}
		return err
	}

	// Log successful execution
	resultJSON, _ := json.Marshal(result)
	resultStr := string(resultJSON)
	we.db.AddExecutionLog(execution.ID, node.ID, "success", &resultStr, nil)

	// Log node execution success to ClickHouse
	if we.logger != nil {
		var outputMap map[string]interface{}
		if resultMap, ok := result.(map[string]interface{}); ok {
			outputMap = resultMap
		} else {
			outputMap = map[string]interface{}{"result": result}
		}

		we.logger.LogExecution(ctx, logger.ExecutionLog{
			ExecutionID: execution.ID,
			WorkflowID:  execution.WorkflowID,
			NodeID:      node.ID,
			Timestamp:   completedAt,
			Level:       "INFO",
			Message:     fmt.Sprintf("Node execution completed: %s", node.Type),
			Status:      "SUCCESS",
			DurationMs:  duration,
		})

		we.logger.LogNodeExecution(ctx, logger.NodeExecution{
			ExecutionID: execution.ID,
			WorkflowID:  execution.WorkflowID,
			NodeID:      node.ID,
			StartedAt:   startTime,
			CompletedAt: completedAt,
			Status:      "SUCCESS",
			NodeType:    string(node.Type),
			Input:       execution.Context,
			Output:      outputMap,
			DurationMs:  duration,
		})
	}

	if stats != nil {
		stats.successful++
	}

	// Update context
	execution.Context[node.ID] = result

	// Find and execute next nodes
	nextNodes := we.getNextNodes(node.ID, workflow)
	for _, nextNode := range nextNodes {
		if err := we.executeNode(nextNode, workflow, execution, stats); err != nil {
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
