## CashBot

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/71a01800fa1e44118a0ff8fa3471b867)](https://app.codacy.com/app/washingt0/cashbot?utm_source=github.com&utm_medium=referral&utm_content=washingt0/cashbot&utm_campaign=Badge_Grade_Dashboard)
[![Go Report Card](https://goreportcard.com/badge/github.com/washingt0/cashbot)](https://goreportcard.com/report/github.com/washingt0/cashbot)

### Requirements
*   MongoDB
*   Redis
*   GoLang

### Running
See releases for the latest stable version already build.

You have to set some environment variables:

*   `CASHBOT_API_TOKEN`: Token supplied by telegram's BotFather
*   `CASHBOT_REDIS_URI`: URI for connect to your running instance of redis
*   `CASHBOT_MONGO_URI`: URI for connect to your running instance of MongoDB
*   `CASHBOT_DEBUG`: If this is set, the bot will run in debug mode

```bash
$ CASHBOT_API_TOKEN=yourawsometoken CASHBOT_REDIS_URI=redis://127.0.0.1:6379 CASHBOT_MONGO_URI=mongodb://127.0.0.1:27017 CASHBOT_DEBUG=1 ./cashbot
```
