import { LambdaIntegration, RestApi } from 'aws-cdk-lib/aws-apigateway'
import { MarketplaceLambdas } from '../../lambdas/marketplace'

export function createMarketplaceResource(
  api: RestApi,
  functions: MarketplaceLambdas,
) {
  const betResource = api.root.addResource('marketplace')

  betResource.addMethod('GET', new LambdaIntegration(functions.get))
}
