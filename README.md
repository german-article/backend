# German Article Telegram Bot

A Telegram bot that helps users determine the correct German article (der, die, das) for German nouns, along with usage examples in different grammatical cases.

## Features

- Determine the correct article for German nouns
- Provide translations in English, Russian, or German
- Show usage examples in nominative, with indefinite article, dative, and accusative cases
- Support for multiple interfaces: Telegram bot, HTTP API, and console
- Clean architecture with domain-driven design
- Comprehensive logging and tracing

## Architecture

The application follows clean architecture principles with clear separation of concerns:

```
├── internal/
│   ├── domain/           # Domain entities and business rules
│   ├── application/      # Use cases and application logic
│   ├── infrastructure/   # External dependencies (AI, logging, tracing)
│   └── adapters/         # Interface adapters (HTTP, Telegram, Console)
├── libs/                 # Shared libraries
└── cmd/                  # Application entry points
```

## Technologies Used

- Go 1.23+
- Google Cloud Functions
- Telegram Bot API (telebot.v3)
- Google Gemini API for language processing
- OpenTelemetry for tracing
- Google Cloud Logging

## Setup

### Prerequisites

- Go 1.23 or later
- A Google Cloud account
- A Telegram Bot token (obtain from [@BotFather](https://t.me/botfather))
- Google Cloud credentials configured

### Environment Variables

Set the following environment variables:

- `TELEGRAM_BOT_TOKEN`: Your Telegram bot token
- `PROJECT_ID`: Your Google Cloud project ID (default: "german-article-bot")
- `APPLICATION_NAME`: Application name for logging (default: "article-bot")
- `GCP_ENABLED`: Enable GCP services (default: "true")
- `MODE`: Application mode - "telegram", "http", or "console" (default: "telegram")

### Local Development

1. Clone this repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up Google Cloud credentials:
   ```bash
   gcloud auth application-default login
   ```

4. Run locally:
   ```bash
   # For HTTP server
   go run cmd/app/main.go
   
   # For console testing
   go run cmd/console/main.go Haus en
   ```

### Testing with cURL

Test the HTTP API locally:

```bash
# GET request
curl "http://localhost:8080/article?word=Haus" \
  -H "Accept-Language: en"

# POST request
curl -X POST "http://localhost:8080/article" \
  -H "Content-Type: application/json" \
  -H "Accept-Language: ru" \
  -d '{"word": "Katze"}'
```

### Deployment

Deploy to Google Cloud Functions:

```bash
gcloud functions deploy article-service \
  --entry-point=Invoke \
  --set-secrets="TELEGRAM_BOT_TOKEN=projects/61311145515/secrets/TELEGRAM_BOT_TOKEN:latest" \
  --runtime go123 \
  --gen2 \
  --region=europe-west3 \
  --trigger-http \
  --allow-unauthenticated
```

Set up the Telegram webhook:

```bash
curl -X POST "https://api.telegram.org/bot<TELEGRAM_BOT_TOKEN>/setWebhook" \
  -d "url=https://your-region-your-project.cloudfunctions.net/german-article-bot"
```

## Usage

### Telegram Bot

1. Start a chat with your bot on Telegram
2. Send `/start` to get a welcome message
3. Send any German noun to get information about its article and usage examples

### HTTP API

The API supports both GET and POST requests:

**GET Request:**
```
GET /article?word=Haus
Accept-Language: en
```

**POST Request:**
```
POST /article
Content-Type: application/json
Accept-Language: ru

{
  "word": "Katze"
}
```

**Response Format:**
```json
{
  "success": true,
  "data": [
    {

    }
  ]
}
```

### Console

For testing and development:

```bash
go run cmd/console/main.go Haus en
go run cmd/console/main.go Katze ru
```

## Project Structure Details

- **Domain Layer**: Contains business entities and interfaces
- **Application Layer**: Implements use cases and orchestrates business logic
- **Infrastructure Layer**: Handles external dependencies (AI service, logging, tracing)
- **Adapters Layer**: Implements interfaces for different input/output methods
- **Libraries**: Shared utilities for logging, tracing, and cleanup

## Language Support

The bot automatically detects user language preferences:
- **Telegram**: Uses user's Telegram language settings
- **HTTP API**: Extracts language from `Accept-Language` header
- **Console**: Specified as command line argument

Supported languages:
- English (en) - default
- Russian (ru)
- German (de)

## Error Handling

The application includes comprehensive error handling:
- Input validation
- AI service failures
- Network issues
- Graceful degradation

## Monitoring

- Structured logging with Google Cloud Logging
- Distributed tracing with OpenTelemetry
- Health check endpoint at `/health`

## License

MIT