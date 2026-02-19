# AWS Deployment & Configuration Guide

This guide explains how to set up the necessary AWS infrastructure and configure your Go application to connect to it.

## 1. DynamoDB Setup

The application uses DynamoDB to store tokens for reversible redaction.

### Manual Setup (AWS Console)
1. Go to **DynamoDB** > **Tables** > **Create table**.
2. **Table name**: `pii-tokens` (or whatever you set in `DYNAMO_TABLE_NAME`).
3. **Partition key**: `token` (String).
4. Leave other settings as default and click **Create table**.
5. Once created, go to the **Additional settings** tab.
6. Under **Time to Live (TTL)**, click **Turn on**.
7. **TTL attribute**: `expires_at`.
8. Click **Turn on TTL**.

### AWS CLI Setup
```bash
aws dynamodb create-table \
    --table-name pii-tokens \
    --attribute-definitions AttributeName=token,AttributeType=S \
    --key-schema AttributeName=token,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST

aws dynamodb update-time-to-live \
    --table-name pii-tokens \
    --time-to-live-specification "Enabled=true, AttributeName=expires_at"
```

## 2. IAM Permissions

The IAM Role used by your application (e.g., Lambda Execution Role or EC2 Instance Profile) needs the following permissions for the DynamoDB table:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:PutItem",
                "dynamodb:GetItem",
                "dynamodb:UpdateItem",
                "lambda:UpdateFunctionCode"
            ],

            "Resource": "arn:aws:dynamodb:*:*:table/pii-tokens"
        }
    ]
}
```

## 3. Environment Variables (Lambda/EC2)

Instead of a `.env` file, you should "save" these in the AWS service configuration.

### In AWS Lambda:
1. Go to your Lambda function > **Configuration** > **Environment variables**.
2. Click **Edit** and add the following:
   - `DYNAMO_TABLE_NAME`: `pii-tokens`
   - `AWS_REGION`: `us-east-1` (match your table's region)
   - `API_KEY`: Your secret key (e.g., `sk_prod_...`)
   - `LOG_LEVEL`: `info`

## 4. How the App Connects

The application uses the `aws-sdk-go-v2` library. When running inside AWS (Lambda, EC2, ECS), the SDK automatically searches for credentials in the following order:
1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`).
2. IAM Role for the resource (Recommended).

You **do not** need to hardcode any keys in the code. The `config.LoadDefaultConfig(ctx)` call in `internal/store/dynamodb.go` handles this automatically using the "Default Credentials Provider Chain".

## 5. Initial Lambda Creation

The GitHub Action uses `update-function-code`, which requires the function to **already exist**. You only need to do this once.

### Option A: Via AWS Console
1. Go to **Lambda** > **Functions** > **Create function**.
2. **Function name**: `pii-redaction-api`.
3. **Runtime**: Select **Amazon Linux 2023**.
4. **Architecture**: `x86_64`.
5. Under **Permissions**, ensure the execution role has DynamoDB access (see Section 2).
6. Click **Create function**.

### Option B: Via AWS CLI
```bash
# 1. Build the initial binary
GOOS=linux GOARCH=amd64 go build -o main cmd/server/main.go
zip function.zip main

# 2. Create the function (replace <YOUR_ROLE_ARN>)
aws lambda create-function \
    --function-name pii-redaction-api \
    --runtime provided.al2023 \
    --handler main \
    --architecture x86_64 \
    --role <YOUR_ROLE_ARN> \
    --zip-file fileb://function.zip
```

## 6. GitHub Actions CI/CD

