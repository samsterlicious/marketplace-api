import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha'
import { Construct } from 'constructs'
import { CreateLambdaParams, LambdaConfig } from '.'

export function createUserLambdas(
  scope: Construct,
  config: LambdaConfig,
  params: CreateLambdaParams,
): UserLambdas {
  const get = new GoFunction(scope, 'getUserInfoLambda', {
    entry: 'src/main/user/getUser',
    ...config,
  })

  params.table.grantReadWriteData(get)

  return {
    get,
  }
}

export type UserLambdas = {
  get: GoFunction
}
