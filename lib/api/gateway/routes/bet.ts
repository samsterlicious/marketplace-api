import { LambdaIntegration, RestApi } from 'aws-cdk-lib/aws-apigateway'
import { BetLambdas } from '../../lambdas/bets'

export function createBetResource(api: RestApi, functions: BetLambdas) {
  const betResource = api.root.addResource('bet')

  betResource.addMethod('POST', new LambdaIntegration(functions.create))
}
