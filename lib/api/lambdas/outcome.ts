import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha'
import { Construct } from 'constructs'
import { CreateLambdaParams, LambdaConfig } from '.'

export function createOutcomeLambdas(
  scope: Construct,
  config: LambdaConfig,
  params: CreateLambdaParams,
): OutcomeLambdas {
  const getByUser = new GoFunction(scope, 'getOutcomesByUserLambda', {
    entry: 'src/main/outcome/getByUser',
    ...config,
  })

  const getRankings = new GoFunction(scope, 'getOutcomeRankingsLambda', {
    entry: 'src/main/outcome/getRankings',
    ...config,
  })

  params.table.grantReadData(getByUser)
  params.table.grantReadData(getRankings)

  return {
    getRankings,
    getByUser,
  }
}

export type OutcomeLambdas = {
  getRankings: GoFunction
  getByUser: GoFunction
}
