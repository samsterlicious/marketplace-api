import { Table } from 'aws-cdk-lib/aws-dynamodb'
import { RetentionDays } from 'aws-cdk-lib/aws-logs'
import { Construct } from 'constructs'
import { BetLambdas, createBetLambdas } from './bets'
import { MarketplaceLambdas, createMarketplaceLambdas } from './marketplace'

export function createLambdas(scope: Construct, table: Table): Lambdas {
  const config: LambdaConfig = {
    environment: { TABLE_NAME: table.tableName },
    logRetention: RetentionDays.THREE_DAYS,
  }
  return {
    bet: createBetLambdas(scope, table, config),
    marketplace: createMarketplaceLambdas(scope, table, config),
  }
}

export type Lambdas = {
  bet: BetLambdas
  marketplace: MarketplaceLambdas
}

export type LambdaConfig = {
  environment: { [key: string]: string }
  logRetention: RetentionDays
}
