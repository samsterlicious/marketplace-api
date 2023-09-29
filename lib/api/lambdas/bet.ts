import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha'
import { Rule, Schedule } from 'aws-cdk-lib/aws-events'
import { LambdaFunction } from 'aws-cdk-lib/aws-events-targets'
import { Construct } from 'constructs'
import { CreateLambdaParams, LambdaConfig } from '.'

export function createBetLambdas(
  scope: Construct,
  config: LambdaConfig,
  params: CreateLambdaParams,
): BetLambdas {
  const get = new GoFunction(scope, 'getBetsLambda', {
    entry: 'src/main/bet/get',
    ...config,
  })

  const resolve = new GoFunction(scope, 'resolveBetsLambda', {
    entry: 'src/main/bet/resolve',
    ...config,
  })

  params.table.grantReadData(get)
  params.table.grantReadWriteData(resolve)

  new Rule(scope, 'ResolveBetsRule', {
    schedule: Schedule.cron({ minute: '0', hour: '4' }),
    targets: [new LambdaFunction(resolve)],
  })

  return {
    get,
    resolve,
  }
}

export type BetLambdas = {
  get: GoFunction
  resolve: GoFunction
}
