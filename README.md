# RSS Aggregator

Service for making one RSS Feed out of many.

## Local launch

For local launch you will need to start PostreSQL DB in Docker and service itself.

### Database
For starting/stoping DB use make commands from repository root:
```
make run-local-db
make stop-local-db
```

<b>Please note: container is removed on stopping with all data in it.</b>

### Service

For starting service you have to launch `cmd/service/main.go`

It is necessary to specify OAuth client ID and secret for Google authentication in `RSS_GOOGLE_AUTH_CLIENT_ID` and `RSS_GOOGLE_AUTH_CLIENT_SECRET` environment variables accordingly. All other environment variables are configured for local launch out of the box including database settings.