# Lambda Code Execution

This directory contains Lambda functions for executing code nodes in workflows.

## Overview

Code nodes are executed in AWS Lambda (or LocalStack for local development) to provide:
- **Isolation**: Each execution runs in a separate Lambda container
- **Security**: Sandboxed execution environment
- **Scalability**: Automatic scaling based on workload
- **Timeout Control**: Configurable execution timeouts

## Local Development with LocalStack

LocalStack is configured in `docker-compose.yml` to emulate AWS Lambda locally.

### Available Lambda Functions

1. **code-executor** (Node.js)
   - Runtime: Node.js 18.x
   - Executes JavaScript code
   - Auto-deployed on LocalStack startup

2. **python-executor** (Python)
   - Runtime: Python 3.9+
   - Executes Python code
   - Located in `functions/python-executor.py`

## Event Structure

### JavaScript Executor

```json
{
  "code": "return input.a + input.b;",
  "language": "javascript",
  "input": {
    "a": 5,
    "b": 10
  }
}
```

### Python Executor

```json
{
  "code": "result = input['a'] + input['b']",
  "language": "python",
  "input": {
    "a": 5,
    "b": 10
  }
}
```

## Response Structure

### Success Response

```json
{
  "statusCode": 200,
  "body": {
    "success": true,
    "result": 15,
    "executedAt": "2024-01-01T00:00:00Z"
  }
}
```

### Error Response

```json
{
  "statusCode": 500,
  "body": {
    "success": false,
    "error": "ReferenceError: x is not defined",
    "stack": "..."
  }
}
```

## Testing Lambda Functions Locally

### Using AWS CLI with LocalStack

```bash
# List functions
awslocal lambda list-functions

# Invoke code-executor
awslocal lambda invoke \
  --function-name code-executor \
  --payload '{"code":"return 1+1;","language":"javascript"}' \
  output.json

# View result
cat output.json
```

### Using the Backend API

The backend automatically routes Code nodes to Lambda:

```bash
# Execute a workflow with a Code node
curl -X POST http://localhost:3001/api/workflows/{workflow-id}/execute \
  -H "Content-Type: application/json" \
  -d '{"trigger": {"message": "test"}}'
```

## Deployment to AWS

For production deployment to real AWS Lambda:

1. Update environment variables in `.env`:
```env
USE_LOCALSTACK=false
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
```

2. Deploy functions using AWS CLI or Terraform:

```bash
# Package and deploy
cd lambda/functions
zip -r code-executor.zip index.js
aws lambda create-function \
  --function-name code-executor \
  --runtime nodejs18.x \
  --role arn:aws:iam::YOUR_ACCOUNT:role/lambda-execution-role \
  --handler index.handler \
  --zip-file fileb://code-executor.zip
```

## Security Considerations

### Sandboxing
- Lambda functions run in isolated containers
- Limited execution time (default: 300 seconds)
- Memory limits prevent resource exhaustion
- No access to host filesystem or network (unless configured)

### Input Validation
- Always validate code input before execution
- Implement rate limiting on code execution endpoints
- Consider using AWS WAF for additional protection

### Environment Variables
- Never include secrets in code
- Use AWS Secrets Manager or environment variables
- Rotate credentials regularly

## Custom Lambda Functions

To add custom Lambda functions:

1. Create a new file in `lambda/functions/`
2. Update `init-lambda.sh` to deploy it
3. Update backend executor to support the new function
4. Restart LocalStack: `docker-compose restart localstack`

## Troubleshooting

### Lambda function not found
```bash
# Check if functions are deployed
docker exec workflow-localstack awslocal lambda list-functions
```

### Execution timeout
- Increase Lambda timeout in `init-lambda.sh`
- Check CloudWatch logs (LocalStack logs in Docker)

### Permission denied
- Verify IAM role exists
- Check Lambda execution role permissions

## Logs

View Lambda execution logs:

```bash
# LocalStack logs
docker logs workflow-localstack -f

# Lambda function logs
docker exec workflow-localstack \
  awslocal logs tail /aws/lambda/code-executor --follow
```
