# PII Redaction & Data Masking API

A robust REST API built in Go for detecting and redacting Personally Identifiable Information (PII) from text.

## Features

- **3-Layer Detection Pipeline**: Regex, NER (prose), and Contextual Analysis.
- **Multiple Redaction Modes**: `mask`, `replace`, `hash`, and `tokenize`.
- **Reversible Tokenization**: Revert redacted values using a `/v1/detokenize` endpoint (powered by DynamoDB).
- **High Performance**: Pure Go implementation, <50ms processing for standard text.
- **Compliance Ready**: Helps meet GDPR, HIPAA, and PCI-DSS requirements.
- **AWS Ready**: Easy deployment to Lambda/EC2 with DynamoDB integration. See [AWS_DEPLOYMENT.md](./AWS_DEPLOYMENT.md) for setup.

## API Endpoints

- `POST /v1/detect`: Only detect PII and return metadata.
- `POST /v1/redact`: Detect and redact PII using the specified mode.
- `POST /v1/detokenize`: Restore original values from tokens.
- `GET /v1/health`: Health check.

## Configuration

The API is configured via environment variables. You can set them in a `.env` file or directly in your environment.

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | The port the server listens on | `8080` |
| `LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |
| `AWS_REGION` | AWS region for DynamoDB | `us-east-1` |
| `DYNAMO_TABLE_NAME` | DynamoDB table for token storage | `pii-tokens` |
| `API_KEY` | Secret key for Bearer authentication | `sk_test_123` |

### DynamoDB Setup

The tokenization feature requires a DynamoDB table with the following schema:
- **Partition Key**: `token` (String)
- **TTL Attribute**: `expires_at` (Number, Unix timestamp)

## API Documentation

### 1. Detect PII (`POST /v1/detect`)

Identify PII without modifying the input text.

**Request:**
```json
{
  "text": "John Smith's SSN is 123-45-6789",
  "locale": "en-US"
}
```

**Response:**
```json
{
  "entities_found": 2,
  "detections": [
    {
      "entity_type": "PERSON",
      "text": "John Smith",
      "start": 0,
      "end": 10,
      "confidence": 0.85,
      "detection_method": "ner"
    },
    {
      "entity_type": "SSN",
      "text": "123-45-6789",
      "start": 20,
      "end": 31,
      "confidence": 0.95,
      "detection_method": "regex"
    }
  ],
  "risk_summary": {
    "hipaa_relevant": 2,
    "gdpr_relevant": 2
  },
  "processing_time_ms": 12,
  "request_id": "req_123abc"
}
```

### 2. Redact PII (`POST /v1/redact`)

Detect and redact PII using one of the supported modes: `mask`, `replace`, `hash`, `tokenize`.

**Request:**
```json
{
  "text": "John Smith's SSN is 123-45-6789",
  "mode": "replace",
  "locale": "en-US"
}
```

**Response:**
```json
{
  "redacted_text": "[PERSON]'s SSN is [SSN]",
  "entities_found": 2,
  "detections": [
    {
      "entity_type": "SSN",
      "original_start": 20,
      "original_end": 31,
      "redacted_value": "[SSN]",
      "confidence": 0.95,
      "detection_method": "regex"
    },
    {
      "entity_type": "PERSON",
      "original_start": 0,
      "original_end": 10,
      "redacted_value": "[PERSON]",
      "confidence": 0.85,
      "detection_method": "ner"
    }
  ],
  "processing_time_ms": 15,
  "request_id": "req_456def"
}
```

### 3. Detokenize (`POST /v1/detokenize`)

Restore original values from tokens (requires `tokenize` mode used previously).

**Request:**
```json
{
  "text": "Hello tok_abc123",
  "tokens": ["tok_abc123"]
}
```

**Response:**
```json
{
  "detokenized_text": "Hello John Smith"
}
```

## Development

- **Build**: `make build`
- **Test**: `make test`
- **Run**: `make run`


.......