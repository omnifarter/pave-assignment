# Pave Assignment

This is a RESTful API Starter with a single Hello World API endpoint.

## Prerequisites

**Install Encore:**

- **macOS:** `brew install encoredev/tap/encore`
- **Linux:** `curl -L https://encore.dev/install.sh | bash`
- **Windows:** `iwr https://encore.dev/install.ps1 | iex`

**Setup local DB**
Encore uses Docker for local DB setup. Ensure your local Docker instance is running before starting encore. [Follow the instructions here to learn more.](https://encore.dev/docs/platform/infrastructure/infra)

**Install Temporal CLI**

[Follow the instructions here for getting started.](https://learn.temporal.io/getting_started/go/dev_environment/?os=linux#set-up-a-local-temporal-service-for-development-with-temporal-cli)

## Run Encore locally

Run this command from your application's root folder:

```bash
encore run
```

While `encore run` is running, open [http://localhost:9400/](http://localhost:9400/) to access Encore's [local developer dashboard](https://encore.dev/docs/go/observability/dev-dash).

Here you can see traces for all requests that you made, see your architecture diagram (just a single service for this simple example), and view API documentation in the Service Catalog.

## Run Temporal locally

In a separate terminal instance, run this command from the application root folder:

```bash
temporal server start-dev --db-filename temporal-local.db
```

You can then access the Temporal UI at http://localhost:8233

## Testing

```bash
encore test ./...
```
