# No More Stories

A bot to _automagically_ delete _stories_ from Telegram.

## Why

* Group admins cannot configure it so users cannot forward _Stories_
* Telegram bot API offers no information about the content of shared _Stories_ so group admins can automate protectio from spammers and scammers
* Forwarding a _Story_ is a strategy used by spammers and scammers due to the flaws above

## How

1. Add [@nomorestories_bot](https://telegram.me/nomorestories_bot) to your group
1. Make [@nomorestories_bot](https://telegram.me/nomorestories_bot) an admin able to delete _Stories_

## Contributing

### Environment variables


| Name | Required | Description |
|---|---|---|
| `BOT_TOKEN` | ✅ | The token Telegram's `@BotFather` gives you |
| `BOT_URL` | ✅ | The URL where your bot is reachable by webhooks (e.g.: `https://my.bot/webhook`) |
| `PORT` | ⛔️ | The port to listen for HTTP webhooks (defaults to `8000`) |


### Tests and checks

```console
$ gofmt -w .
$ staticcheck ./...
$ go test ./...
```
