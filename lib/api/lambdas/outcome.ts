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

  params.table.grantReadData(getByUser)

  return {
    getByUser,
  }
}

export type OutcomeLambdas = {
  getByUser: GoFunction
}
