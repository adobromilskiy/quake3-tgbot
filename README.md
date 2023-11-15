# quake3-tgbot

Telegram bot parse info from [quake3-stats](https://github.com/adobromilskiy/quake3-stats).

You can ask bot about players stats or last matches info. To get info from bot your message must start from 'q3bot ...'

Example:

> "q3bot show me last match results"
>
> "q3bot what's stats for <player_name>"
>
> "q3bot get info about last 5 matches"

## Usage

Launch quake3-tgbot with next parameters:

| parameter | description |
|-----------|-------------|
| -v         | enable debug mode |
| --telegram | telegram bot token from BotFather |
| --openai   | OpenAI token |
| --server   | [quake3-stats](https://github.com/adobromilskiy/quake3-stats) server url |

Example:

```sh
./quake3-tgbot -v --telegram=<tg-bot-token> --openai=<openai-token> --server=<quake3-stats-server-url>
```

or

```sh
docker run -ti --rm adobromilskiy/quake3-tgbot:latest -v --telegram=<tg-bot-token> --openai=<openai-token> --server=<quake3-stats-server-url>
```