import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha'
import { Table } from 'aws-cdk-lib/aws-dynamodb'
import { Construct } from 'constructs'
import { LambdaConfig } from '.'

export function createBetLambdas(
  scope: Construct,
  table: Table,
  config: LambdaConfig,
): BetLambdas {
  const createBet = new GoFunction(scope, 'createBetFunction', {
    entry: 'src/main/bet',
    ...config,
  })

  table.grantReadWriteData(createBet)

  return {
    create: createBet,
  }
}

export type BetLambdas = {
  create: GoFunction
}
