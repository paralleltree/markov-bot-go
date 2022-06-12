# markov-bot-go

An application posting generated text from posts on Mastodon using the Markov chain.

## Build and Run

    $ docker-compose build app
    $ docker compose run --rm app /app/bot run

Run `docker-compose run --rm app /app/bot help` to view help.
You can also pass the required arguments as environment variables.
