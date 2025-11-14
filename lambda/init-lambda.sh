#!/bin/bash

# Wait for LocalStack to be fully ready
echo "Waiting for LocalStack to be ready..."
sleep 5

# Create Lambda execution role
echo "Creating Lambda execution role..."
awslocal iam create-role \
  --role-name lambda-execution-role \
  --assume-role-policy-document '{
    "Version": "2012-10-17",
    "Statement": [{
      "Effect": "Allow",
      "Principal": {"Service": "lambda.amazonaws.com"},
      "Action": "sts:AssumeRole"
    }]
  }' 2>/dev/null || echo "Role already exists"

# Create default code executor Lambda function
echo "Creating code-executor Lambda function..."

# Create a simple Node.js Lambda handler
cat > /tmp/index.js << 'EOF'
exports.handler = async (event) => {
    const { code, language, input } = event;

    try {
        let result;

        if (language === 'javascript' || language === 'nodejs') {
            // Execute JavaScript code
            const func = new Function('input', code);
            result = await func(input);
        } else if (language === 'python') {
            // Python execution would require a Python runtime
            result = { error: 'Python execution requires Python runtime Lambda' };
        } else {
            result = { error: `Unsupported language: ${language}` };
        }

        return {
            statusCode: 200,
            body: JSON.stringify({
                success: true,
                result: result,
                executedAt: new Date().toISOString()
            })
        };
    } catch (error) {
        return {
            statusCode: 500,
            body: JSON.stringify({
                success: false,
                error: error.message,
                stack: error.stack
            })
        };
    }
};
EOF

# Create deployment package
cd /tmp
zip -q code-executor.zip index.js

# Deploy Lambda function
awslocal lambda create-function \
  --function-name code-executor \
  --runtime nodejs18.x \
  --role arn:aws:iam::000000000000:role/lambda-execution-role \
  --handler index.handler \
  --zip-file fileb://code-executor.zip \
  --timeout 300 \
  --memory-size 512 \
  --environment "Variables={NODE_ENV=production}" 2>/dev/null || \
awslocal lambda update-function-code \
  --function-name code-executor \
  --zip-file fileb://code-executor.zip

echo "Lambda functions initialized successfully!"

# List all functions
echo "Available Lambda functions:"
awslocal lambda list-functions --query 'Functions[].FunctionName' --output table
