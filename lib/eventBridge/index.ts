import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha'
import { Rule, Schedule } from 'aws-cdk-lib/aws-events'
import { LambdaFunction } from 'aws-cdk-lib/aws-events-targets'
import { Construct } from 'constructs'

export function createEventBridge(scope: Construct, lambda: GoFunction): void {
  new Rule(scope, 'MarketplacePopulatorRule', {
    schedule: Schedule.cron({ minute: '0', hour: '0' }),
    targets: [new LambdaFunction(lambda)],
  })
}
