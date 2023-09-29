import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha'
import { Construct } from 'constructs'
import { CreateLambdaParams, LambdaConfig } from '.'

export function createLeagueLambdas(
  scope: Construct,
  config: LambdaConfig,
  params: CreateLambdaParams,
): LeagueLambdas {
  const getUsers = new GoFunction(scope, 'getUsersInLeagueLambda', {
    entry: 'src/main/league/getUsersInLeague',
    ...config,
  })

  params.table.grantReadWriteData(getUsers)

  return {
    getUsers,
  }
}

export type LeagueLambdas = {
  getUsers: GoFunction
}
