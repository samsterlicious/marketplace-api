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

  const updateName = new GoFunction(scope, 'updateNameLambda', {
    entry: 'src/main/user/updateUsername',
    ...config,
  })

  params.table.grantReadWriteData(get)
  params.table.grantReadWriteData(updateName)

  return {
    get,
    updateName,
  }
}

export type UserLambdas = {
  get: GoFunction
  updateName: GoFunction
}
