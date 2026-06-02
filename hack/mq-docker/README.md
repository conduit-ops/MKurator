# IBM MQ (Docker) for integration tests

Runs a standalone Queue Manager with **mqweb** on `https://127.0.0.1:9443` for
`task test:integration` — no kind cluster required.

## Prerequisites

- Docker (or a compatible runtime with `docker compose`)
- Network access to pull `icr.io/ibm-messaging/mq:9.4.2.0-r1` (same tag as
  `hack/kind-cluster`)

## Quick start

```sh
task mq:integration:up
task mq:integration:wait
task test:integration
task mq:integration:down
```

Or one shot:

```sh
task test:integration:local
```

First start can take several minutes while the queue manager initializes.

## Credentials

| Field | Value |
|-------|--------|
| Queue manager | `QM1` |
| Admin user | `admin` |
| Admin password | `passw0rd` |

Do not log passwords in CI or operator logs.

## Troubleshooting

- **Slow / healthcheck failing**: `docker compose -f hack/mq-docker/docker-compose.yml logs -f`
- **Port 9443 in use**: stop the other service or change the host port in
  `docker-compose.yml` and set `KURATOR_INTEGRATION_MQ_ENDPOINT` accordingly.
- **Reuse kind MQ instead**: with `task cluster:up`, point integration tests at the
  NodePort (see [docs/DEVELOPMENT.md](../../docs/DEVELOPMENT.md#test-tiers)).

## Teardown

```sh
task mq:integration:down
```

To remove volumes as well: `docker compose -f hack/mq-docker/docker-compose.yml down -v`
