# Proveedores De Email Y Plugins

Fecha de referencia: 2026-04-02.

GoFrame incluye una capa de mail desacoplada en `pkg/mail` para soportar:

- drivers builtin (`noop`, `smtp`, `sendgrid`)
- extensiones in-process por registro de proveedor
- plugins externos por ejecutable en `PATH`

## 1. Configuracion

Campos relevantes en `goframe.yaml`:

```yaml
mail_driver: noop
mail_from: noreply@localhost

# SMTP
smtp_host: ""
smtp_port: 587
smtp_user: ""
smtp_pass: ""

# SendGrid
sendgrid_api_key: ""
sendgrid_endpoint: https://api.sendgrid.com/v3/mail/send
```

`mail_driver` controla el proveedor activo.

## 2. Como se conecta con el framework

`app.New(...)` resuelve el proveedor y deja una instancia lista en:

- `app.App.Mailer` (`mail.Sender`)

`goframe sendtestemail` usa el mismo resolver de proveedores y por eso valida de extremo a extremo la configuracion real.

## 3. Extension in-process (registro de proveedor)

Puedes registrar un proveedor propio desde codigo Go, por ejemplo en un paquete cargado al arrancar:

```go
package mymail

import (
	"context"
	"fmt"
	"strings"

	"github.com/jcsvwinston/GoFrame/pkg/mail"
)

type mailgunSender struct {
	apiKey string
}

func (s *mailgunSender) Send(ctx context.Context, msg mail.Message) error {
	_ = ctx
	_ = msg
	// Implementar llamada HTTP al proveedor.
	return nil
}

func init() {
	_ = mail.RegisterProvider("mailgun", func(cfg mail.Config) (mail.Sender, error) {
		key := strings.TrimSpace(cfg.SendGridAPIKey) // ejemplo; usa tu propia clave/campo
		if key == "" {
			return nil, fmt.Errorf("mailgun api key is required")
		}
		return &mailgunSender{apiKey: key}, nil
	})
}
```

Luego configuras:

```yaml
mail_driver: mailgun
```

## 4. Plugin externo (`goframe-mail-<driver>`)

Si no hay proveedor registrado, GoFrame busca un binario:

- nombre: `goframe-mail-<mail_driver>`
- ejemplo: `mail_driver: mailgun` -> `goframe-mail-mailgun`

Contrato de entrada por `stdin` (JSON):

```json
{
  "driver": "mailgun",
  "from": "noreply@example.com",
  "to": ["dev@example.com"],
  "subject": "Hello",
  "body": "Message body",
  "headers": {"X-Correlation-ID": "abc-123"}
}
```

Contrato de salida:

- `exit code 0`: envio aceptado
- `exit code != 0`: error de envio/configuracion
- `stderr`: detalle de error (se propaga al caller)

## 5. Patron recomendado en capas de aplicacion

Haz que tus servicios dependan de la interfaz `mail.Sender` y no de un proveedor concreto:

```go
type NotificationService struct {
	mailer mail.Sender
}

func NewNotificationService(mailer mail.Sender) *NotificationService {
	return &NotificationService{mailer: mailer}
}
```

En el bootstrap:

```go
a, _ := app.New(cfg)
notifications := NewNotificationService(a.Mailer)
```

Con esto, cambiar de SMTP a SendGrid/Mailgun/plugin no requiere tocar la logica de negocio.
