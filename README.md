## CashBot

### Requirements
* MongoDB
* Redis
* GoLang

### Running
See releases for the latest stable version already build.

You have to set some environment variables:

* `CASHBOT_API_TOKEN`: Token supplied by telegram's BotFather
* `CASHBOT_REDIS_URI`: URI for connect to your running instance of redis
* `CASHBOT_MONGO_URI`: URI for connect to your running instance of MongoDB
* `CASHBOT_DEBUG`: If this is set, the bot will run in debug mode

```bash
$ CASHBOT_API_TOKEN=yourawsometoken CASHBOT_REDIS_URI=redis://127.0.0.1:6379 CASHBOT_MONGO_URI=mongodb://127.0.0.1:27017 CASHBOT_DEBUG=1 ./cashbot
```
