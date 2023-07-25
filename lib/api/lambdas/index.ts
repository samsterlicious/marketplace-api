import { Table } from 'aws-cdk-lib/aws-dynamodb'
import { Construct } from 'constructs'
import { BetLambdas, createBetLambdas } from './bets'

export function createLambdas(scope: Construct, table: Table): Lambdas {
  return {
    bet: createBetLambdas(scope, table),
  }
}

export type Lambdas = {
  bet: BetLambdas
}
