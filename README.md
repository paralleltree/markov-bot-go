# markov-bot-go

An application posting generated text from posts on Mastodon using the Markov chain.

## Build and Run locally

    $ docker-compose build app
    $ docker compose run --rm app /app/bot run

Run `docker-compose run --rm app /app/bot run --help` to view help.
You can also pass the required arguments as environment variables.

## Build as AWS Lambda Function

1. Create ECR repository to upload container image and make note of the repository url.
1. Enter to `cmd/lambda`.
1. Run `sam build` to build container image.
1. Run `sam deploy --guided` and pass the repository url to create new lambda function.

To use the function, follow steps below:

* Set event to specify which S3 bucket and key path will be used to place configuration and model file.
  * See `PostEvent` struct in `cmd/lambda/main.go`.
* Put configration file on S3.
  * See `Config` struct in `cmd/lambda/main.go`.
