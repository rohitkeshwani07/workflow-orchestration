package lambda

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// LambdaExecutor handles code execution via AWS Lambda
type LambdaExecutor struct {
	client      *lambda.Lambda
	useLocalStack bool
}

// CodeExecutionRequest represents the payload sent to Lambda
type CodeExecutionRequest struct {
	Code     string                 `json:"code"`
	Language string                 `json:"language"`
	Input    map[string]interface{} `json:"input"`
	Timeout  int                    `json:"timeout,omitempty"`
}

// CodeExecutionResponse represents the response from Lambda
type CodeExecutionResponse struct {
	Success     bool        `json:"success"`
	Result      interface{} `json:"result,omitempty"`
	Output      string      `json:"output,omitempty"`
	Error       string      `json:"error,omitempty"`
	Stack       string      `json:"stack,omitempty"`
	Traceback   string      `json:"traceback,omitempty"`
	ExecutedAt  string      `json:"executedAt,omitempty"`
}

// NewLambdaExecutor creates a new Lambda executor
func NewLambdaExecutor() (*LambdaExecutor, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	useLocalStack := os.Getenv("USE_LOCALSTACK") == "true"

	config := &aws.Config{
		Region: aws.String(region),
	}

	// Configure for LocalStack if enabled
	if useLocalStack {
		endpoint := os.Getenv("AWS_ENDPOINT")
		if endpoint == "" {
			endpoint = "http://localhost:4566"
		}

		accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
		if accessKeyID == "" {
			accessKeyID = "test"
		}

		secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if secretAccessKey == "" {
			secretAccessKey = "test"
		}

		config.Endpoint = aws.String(endpoint)
		config.Credentials = credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
		config.S3ForcePathStyle = aws.Bool(true)
		config.DisableSSL = aws.Bool(true)

		log.Printf("Using LocalStack Lambda at %s", endpoint)
	} else {
		log.Printf("Using AWS Lambda in region %s", region)
	}

	// Create AWS session
	sess, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create Lambda client
	client := lambda.New(sess)

	return &LambdaExecutor{
		client:        client,
		useLocalStack: useLocalStack,
	}, nil
}

// ExecuteCode executes code in a Lambda function
func (e *LambdaExecutor) ExecuteCode(code string, language string, input map[string]interface{}) (*CodeExecutionResponse, error) {
	// Determine which Lambda function to use based on language
	functionName := e.getFunctionName(language)

	// Prepare request payload
	request := CodeExecutionRequest{
		Code:     code,
		Language: language,
		Input:    input,
		Timeout:  60,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Invoke Lambda function
	log.Printf("Invoking Lambda function: %s", functionName)
	result, err := e.client.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		Payload:      payload,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to invoke Lambda: %w", err)
	}

	// Check for function errors
	if result.FunctionError != nil {
		return &CodeExecutionResponse{
			Success: false,
			Error:   fmt.Sprintf("Lambda function error: %s", *result.FunctionError),
		}, nil
	}

	// Parse Lambda response
	var lambdaResponse struct {
		StatusCode int    `json:"statusCode"`
		Body       string `json:"body"`
	}

	if err := json.Unmarshal(result.Payload, &lambdaResponse); err != nil {
		// If response is not in expected format, try parsing directly
		var directResponse CodeExecutionResponse
		if err := json.Unmarshal(result.Payload, &directResponse); err == nil {
			return &directResponse, nil
		}
		return nil, fmt.Errorf("failed to parse Lambda response: %w", err)
	}

	// Parse the body
	var response CodeExecutionResponse
	if err := json.Unmarshal([]byte(lambdaResponse.Body), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	// Check status code
	if lambdaResponse.StatusCode >= 400 {
		response.Success = false
	}

	return &response, nil
}

// getFunctionName returns the Lambda function name for a given language
func (e *LambdaExecutor) getFunctionName(language string) string {
	switch language {
	case "javascript", "js", "nodejs", "node":
		return "code-executor"
	case "python", "py":
		return "python-executor"
	case "go", "golang":
		return "go-executor"
	default:
		// Default to code-executor for unknown languages
		return "code-executor"
	}
}

// ListFunctions lists all available Lambda functions
func (e *LambdaExecutor) ListFunctions() ([]string, error) {
	result, err := e.client.ListFunctions(&lambda.ListFunctionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}

	functions := make([]string, 0, len(result.Functions))
	for _, fn := range result.Functions {
		if fn.FunctionName != nil {
			functions = append(functions, *fn.FunctionName)
		}
	}

	return functions, nil
}

// TestConnection tests the connection to Lambda service
func (e *LambdaExecutor) TestConnection() error {
	_, err := e.client.ListFunctions(&lambda.ListFunctionsInput{
		MaxItems: aws.Int64(1),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Lambda service: %w", err)
	}
	return nil
}
