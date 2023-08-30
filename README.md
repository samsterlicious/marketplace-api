# Welcome to your CDK TypeScript project

This is a blank project for CDK development with TypeScript.

The `cdk.json` file tells the CDK Toolkit how to execute your app.

## Useful commands

- `npm run build` compile typescript to js
- `npm run watch` watch for changes and compile
- `npm run test` perform the jest unit tests
- `cdk deploy` deploy this stack to your default AWS account/region
- `cdk diff` compare deployed stack with current state
- `cdk synth` emits the synthesized CloudFormation template

## DynamoDB Structure

|                      ID                      |                                 SortKey                                 |   GSI1_ID   |              GSI1_SortKey               | Spread |     TTL      | HomeAmount | AwayAmount | Amount |
| :------------------------------------------: | :---------------------------------------------------------------------: | :---------: | :-------------------------------------: | :----: | :----------: | :--------: | :--------: | :----: |
|                      MK                      |                 sport\|league\|date\|awayTeam\|homeTeam                 |             |                                         | spread | date + 1 day | HomeAmount | AwayAmount |        |
| BID\|sport\|league\|date\|awayTeam\|homeTeam |                      chosenTeam\|user\|createDate                       |  BID\|user  |                 amount                  | spread | date + 1 day |            |            |        |
|                  BET\|user                   | sport\|league\|date\|awayTeam\|homeTeam\|chosenTeam\|user1\|date of bet | BET\|Status | sport\|league\|date\|awayTeam\|homeTeam | spread | date + 1 day |            |            | amount |

## Access patternz
