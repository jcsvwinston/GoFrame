# Plugin SDK Official Examples

Reference date: 2026-04-05.
Status: Current.

This guide documents the official Plugin SDK v1 examples shipped in repository:

- mail capability example (`mail.send`)
- queue capability example (`queue.publish`)

Source directory:

- `examples/plugins/`

## Included Providers

1. `examplemail`
- binary name: `nucleus-plugin-examplemail`
- capability: `mail.send`
- executable path in repo: `examples/plugins/mail`

2. `examplequeue`
- binary name: `nucleus-plugin-examplequeue`
- capability: `queue.publish`
- executable path in repo: `examples/plugins/queue`

Both examples implement:

- `capabilities`
- `capabilities --json`
- request envelope decoding (`version: v1`)
- response envelope encoding (`ok` + `error` + `metrics`)
- exit-code mapping aligned with `pkg/plugins` semantics

## Build

From repository root:

```bash
mkdir -p .tmp/plugins

go build -o .tmp/plugins/nucleus-plugin-examplemail ./examples/plugins/mail
go build -o .tmp/plugins/nucleus-plugin-examplequeue ./examples/plugins/queue
```

Add binaries to `PATH`:

```bash
export PATH="$(pwd)/.tmp/plugins:$PATH"
```

## Verify with CLI

```bash
nucleus plugin list --config nucleus.yml
nucleus plugin doctor --config nucleus.yml

nucleus plugin test --provider examplemail --capability mail.send --execute
nucleus plugin test --provider examplequeue --capability queue.publish --execute
```

## Mail End-to-End Smoke

Set this in your app config:

```yaml
mail_driver: examplemail
mail_from: noreply@example.com
```

Run:

```bash
nucleus sendtestemail --config nucleus.yml --to dev@example.com --subject "mail plugin smoke"
```

## Queue Envelope Smoke

```bash
cat <<'JSON' | nucleus-plugin-examplequeue
{
  "version": "v1",
  "request_id": "req_queue_demo",
  "timestamp": "2026-04-05T12:00:00Z",
  "capability": "queue.publish",
  "provider": "examplequeue",
  "timeout_ms": 5000,
  "payload": {
    "topic": "events.users",
    "body": {"event": "user.created", "id": 42}
  }
}
JSON
```
