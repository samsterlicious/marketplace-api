import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha'
import { Rule, Schedule } from 'aws-cdk-lib/aws-events'
import { LambdaFunction } from 'aws-cdk-lib/aws-events-targets'
import { Effect, PolicyStatement } from 'aws-cdk-lib/aws-iam'
import { Construct } from 'constructs'
import { CreateLambdaParams, LambdaConfig } from '.'

export function createMarketplaceLambdas(
  scope: Construct,
  config: LambdaConfig,
  params: CreateLambdaParams,
): MarketplaceLambdas {
  const getAvailableEvents = new GoFunction(scope, 'getAvailableEventsLambda', {
    entry: 'src/main/marketplace/get',
    ...config,
  })

  const createEvents = new GoFunction(scope, 'createEventsLambda', {
    entry: 'src/main/marketplace/create',
    ...config,
    environment: {
      ...config.environment,
    },
  })

  const getEspnInfo = new GoFunction(scope, 'getEspnInfo', {
    entry: 'src/main/marketplace/getEspnInfo',
    ...config,
  })

  params.table.grantReadWriteData(createEvents)
  params.table.grantReadData(getAvailableEvents)

  createEvents.addToRolePolicy(
    new PolicyStatement({
      actions: ['events:PutRule', 'events:PutTargets'],
      resources: [
        `arn:aws:events:${process.env.CDK_DEFAULT_REGION}:${process.env.CDK_DEFAULT_ACCOUNT}:rule/*`,
        `arn:aws:events:${process.env.CDK_DEFAULT_REGION}:${process.env.CDK_DEFAULT_ACCOUNT}:event-bus/default`,
      ],
      effect: Effect.ALLOW,
    }),
  )

  // const eventBus = EventBus.fromEventBusName(scope, 'DefaultBus', 'default')

  new Rule(scope, 'MarketplacePopulatorRule', {
    schedule: Schedule.cron({ minute: '0', hour: '4', weekDay: '3' }),
    targets: [new LambdaFunction(createEvents)],
  })

  return {
    create: createEvents,
    getEspnInfo,
    get: getAvailableEvents,
  }
}

export type MarketplaceLambdas = {
  create: GoFunction
  get: GoFunction
  getEspnInfo: GoFunction
}
