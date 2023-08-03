import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha'
import { Table } from 'aws-cdk-lib/aws-dynamodb'
import { Construct } from 'constructs'
import { LambdaConfig } from '.'

export function createMarketplaceLambdas(
  scope: Construct,
  table: Table,
  config: LambdaConfig,
): MarketplaceLambdas {
  const getAvailableEvents = new GoFunction(
    scope,
    'getAvailableEventsFunction',
    {
      entry: 'src/main/marketplace/get',
      ...config,
    },
  )

  const createEvents = new GoFunction(scope, 'createEventsFunction', {
    entry: 'src/main/marketplace/create',
    ...config,
  })

  table.grantReadWriteData(createEvents)
  table.grantReadData(getAvailableEvents)

  return {
    create: createEvents,
    get: getAvailableEvents,
  }
}

export type MarketplaceLambdas = {
  create: GoFunction
  get: GoFunction
}
