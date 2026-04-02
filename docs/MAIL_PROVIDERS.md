# Mail Providers and Plugins

Reference date: 2026-04-02.

GoFrame includes a pluggable mail layer in `pkg/mail`.

## Supported Drivers

Built-in drivers:

- `noop`
- `smtp`
- `sendgrid`

Extensibility options:

- in-process registration via `mail.RegisterProvider(...)`
- external binary plugin `goframe-mail-<driver>` on `PATH`

## Configuration

Typical keys in `goframe.yaml`:

```yaml
mail_driver: noop
mail_from: noreply@localhost

smtp_host: ""
smtp_port: 587
smtp_user: ""
smtp_pass: ""

sendgrid_api_key: ""
sendgrid_endpoint: https://api.sendgrid.com/v3/mail/send
```

## Operational Commands

```bash
goframe sendtestemail --config goframe.yaml --to dev@example.com --dry-run
goframe sendtestemail --config goframe.yaml --driver sendgrid --to dev@example.com --dry-run
goframe mailproviders --config goframe.yaml
goframe mailproviders --config goframe.yaml --json
```

## External Plugin Contract

If `mail_driver: mailgun`, GoFrame looks for `goframe-mail-mailgun`.

Input is sent over `stdin` as JSON:

- `driver`
- `from`
- `to`
- `subject`
- `body`
- `headers`

Exit code contract:

- `0`: accepted
- non-zero: failed
