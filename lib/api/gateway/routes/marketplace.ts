import { HttpApi, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha'
import { HttpJwtAuthorizer } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha'
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha'
import { MarketplaceLambdas } from '../../lambdas/marketplace'

export function createMarketplaceResource(
  api: HttpApi,
  functions: MarketplaceLambdas,
  authorizer: HttpJwtAuthorizer,
) {
  const getMarketplaceIntegration = new HttpLambdaIntegration(
    'CreateIntegration',
    functions.get,
  )

  api.addRoutes({
    path: '/marketplace',
    methods: [HttpMethod.GET],
    integration: getMarketplaceIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })
}
