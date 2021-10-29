# RSS Aggregator

Service for making one RSS Feed out of many.

## Development

For local launch from IDE you will need to start PostreSQL DB in Docker.

For starting/stoping DB use make commands from repository root:
```
make run-local-db
make stop-local-db
```

<b>Please note: container is removed on stopping with all data in it.</b>

## Local launch

It is necessary to specify OAuth client ID and secret for Google authentication in `RSS_GOOGLE_AUTH_CLIENT_ID` and `RSS_GOOGLE_AUTH_CLIENT_SECRET` environment variables accordingly. All other environment variables are configured for local launch out of the box including database settings.

Example:
```
export RSS_GOOGLE_AUTH_CLIENT_ID=client_id
export RSS_GOOGLE_AUTH_CLIENT_SECRET=secret
```

### Docker compose

Run command in repository root

```
docker compose start
```

Service url: http://localhost/

### Minikube
Assuming minikube is running and ingress addon is enabled, only one command required in repository root:

```
make run-minikube
```

You will be prompted for sudo password to edit `/etc/hosts`

Service url: http://rss.aggregator.test.com/

Minikube could be run and configured as following:
```
minikube start --driver=hyperkit
minikube addons enable ingress
```