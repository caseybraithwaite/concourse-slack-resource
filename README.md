# Concourse Slack Resource

Resource for sending and updating in Slack from Concourse. There's a few other resources available but as far as I'm aware this is the only one that supports sending blocks.

## Configuration

### Source

*channel* - **required** - the channel id to send the message to - either this or params.channel must be provided

*bot_token* - **required** - the OAUTH token for the Slack bot to use

*debug* - **optional** - include verbose status messages - do not use in production, this will leak credentials.

### Params

*blocks* - **required** - the blocks to send to slack - see the example below

*blocks_file* - **required** - a file containing the blocks to send to slack - required if not setting the blocks option

*timestamp* - **optional** - the timestamp of a previous message to update

*thread_ts* - **optional** - the timestamp of a previous message to start a thread under

*text* - **optional** - text to send, if provided with blocks then this will be used as fallback text (for example, in notifications where blocks can't be rendered)

*channel* - **optional** - the channel id to send the message to - will override source.channel if provided

### A note on channels

The Slack APIs for posting messages and updating messages are slightly different. When posting a message you can use the name of the channel, whereas when updating a message with the `chat.update` API you must provide the channel ID. It's recommended that you just specify the channel ID in the source configuration for this resource since that works everywhere.

### Environment variable interpolation

Concourse provides a few environment variables to `put` steps that contain metadata about the build. This resource will interpolate environment variables given they're in the `{{ $ENV_VAR }}` format.

For example,

```yaml
- put: slack-general
  params:
  blocks: |
    [
      {
        "type": "section",
        "text": {
            "type": "mrkdwn",
            "text": "Concourse available at {{ $ATC_EXTERNAL_URL }}"
        }
      }
    ]
```

### No Get

Since v7.9.0, Concourse has supported an option called `no_get` in put steps. This skips the implicit get step right after doing a put. You can use `no_get` with this resource as long as you don't need the message timestamp later on.

## An example

```yaml
jobs:
  - name: slack
    plan:
      # Send a message
      - put: slack-general
        params:
          text: Fallback text displaed when blocks can't be rendered in notifications!
          blocks: |
            [
              {
                "type": "section",
                "text": {
                  "type": "mrkdwn",
                  "text": "Hello from <{{ $ATC_EXTERNAL_URL }}|Concourse>!"
                }
              }
            ]

      - load_var: timestamp
        file: slack-general/ts

      - task: wait
        config:
          platform: linux
          image_resource:
            type: registry-image
            source: {repository: alpine}
          outputs:
            - name: msg
          run:
            path: sh
            args:
              - -exc
              - |
                sleep 5

                echo -n '[
                  {
                    "type": "section",
                    "text": {
                      "type": "mrkdwn",
                      "text": "Updated"
                    }
                  }
                ]' > msg/blocks
      
      # Update the message from earlier using the contents of a file
      - put: slack-general
        inputs:
          - msg
        params:
          timestamp: ((.:timestamp))
          blocks_file: msg/blocks
      
      # Send a thread
      - put: slack-general
        inputs: []
        params:
          thread_ts: ((.:timestamp))
          blocks: |
            [
              {
                "type": "section",
                "text": {
                  "type": "mrkdwn",
                  "text": "I'm a thread message!"
                }
              }
            ]

resources:
  - name: slack-general
    check_every: never
    type: slack
    source:
      channel: "((secret/slack/channel))"
      bot_token: "((secret/slack))"

resource_types:
  - name: slack
    type: registry-image
    source:
      repository: ca5ey32/concourse-slack-resource
      tag: 0.1.0
```