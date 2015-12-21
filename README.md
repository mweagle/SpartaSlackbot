# SpartaSlackbot
Slack bot example application using [Sparta](https://github.com/mweagle/Sparta)

See the [Sparta Docs](http://gosparta.io/docs/apigateway/slack/) for more information, including steps for how to configure Slack.

## Instructions

  1. `make get`
  1. `S3_BUCKET=<MY_S3_BUCKET_NAME> make provision`
  1. Query `curl -vs --data "text=HelloWorld" https://**REST_API_ID**.execute-api.**AWS_REGION**.amazonaws.com/v1/slack`
