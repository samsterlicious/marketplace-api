import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha'
import { Rule, Schedule } from 'aws-cdk-lib/aws-events'
import { LambdaFunction } from 'aws-cdk-lib/aws-events-targets'
import { Effect, PolicyStatement, ServicePrincipal } from 'aws-cdk-lib/aws-iam'
import { Construct } from 'constructs'
import { CreateLambdaParams, LambdaConfig } from '.'

export function createBetLambdas(
  scope: Construct,
  config: LambdaConfig,
  params: CreateLambdaParams,
): BetLambdas {
  const createBet = new GoFunction(scope, 'createBetLambda', {
    entry: 'src/main/bet/create',
    ...config,
  })

  const get = new GoFunction(scope, 'getBetsLambda', {
    entry: 'src/main/bet/get',
    ...config,
  })

  const resolve = new GoFunction(scope, 'resolveBetsLambda', {
    entry: 'src/main/bet/resolve',
    ...config,
  })

  params.table.grantReadWriteData(createBet)
  params.table.grantReadData(get)
  params.table.grantReadWriteData(resolve)

  createBet.addToRolePolicy(
    new PolicyStatement({
      actions: ['events:DeleteRule', 'events:RemoveTargets'],
      resources: [
        `arn:aws:events:${process.env.CDK_DEFAULT_REGION}:${process.env.CDK_DEFAULT_ACCOUNT}:rule/*`,
      ],
      effect: Effect.ALLOW,
    }),
  )

  createBet.addPermission('rulePermission', {
    action: 'lambda:InvokeFunction',
    principal: new ServicePrincipal('events.amazonaws.com'),
    sourceArn: `arn:aws:events:${process.env.CDK_DEFAULT_REGION}:${process.env.CDK_DEFAULT_ACCOUNT}:rule/*`,
  })
  new LambdaFunction(createBet)

  new Rule(scope, 'ResolveBetsRule', {
    schedule: Schedule.cron({ minute: '0', hour: '4' }),
    targets: [new LambdaFunction(resolve)],
  })

  return {
    get,
    create: createBet,
    resolve,
  }
}

export type BetLambdas = {
  create: GoFunction
  get: GoFunction
  resolve: GoFunction
}
