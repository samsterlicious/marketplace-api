import { Table } from 'aws-cdk-lib/aws-dynamodb'
import { RetentionDays } from 'aws-cdk-lib/aws-logs'
import { Construct } from 'constructs'
import { BetLambdas, createBetLambdas } from './bet'
import { BidLambdas, createBidLambdas } from './bid'
import { LeagueLambdas, createLeagueLambdas } from './league'
import { MarketplaceLambdas, createMarketplaceLambdas } from './marketplace'
import { OutcomeLambdas, createOutcomeLambdas } from './outcome'
import { UserLambdas, createUserLambdas } from './user.'

export function createLambdas(
  scope: Construct,
  params: CreateLambdaParams,
): Lambdas {
  const lambdaConfig: LambdaConfig = {
    environment: { TABLE_NAME: params.table.tableName },
    logRetention: RetentionDays.ONE_DAY,
  }

  const betLambdas = createBetLambdas(scope, lambdaConfig, params)
  return {
    bet: betLambdas,
    bid: createBidLambdas(scope, lambdaConfig, params),
    marketplace: createMarketplaceLambdas(scope, lambdaConfig, params),
    league: createLeagueLambdas(scope, lambdaConfig, params),
    outcome: createOutcomeLambdas(scope, lambdaConfig, params),
    user: createUserLambdas(scope, lambdaConfig, params),
  }
}

export type Lambdas = {
  bet: BetLambdas
  bid: BidLambdas
  marketplace: MarketplaceLambdas
  league: LeagueLambdas
  outcome: OutcomeLambdas
  user: UserLambdas
}

export type LambdaConfig = {
  environment: { [key: string]: string }
  logRetention: RetentionDays
}

export type CreateLambdaParams = {
  table: Table
}
