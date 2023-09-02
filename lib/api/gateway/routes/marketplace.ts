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
    'GetMarketplaceIntegration',
    functions.get,
  )

  const getEspnInfoIntegration = new HttpLambdaIntegration(
    'GetEspnInfoIntegration',
    functions.getEspnInfo,
  )

  api.addRoutes({
    path: '/marketplace',
    methods: [HttpMethod.GET],
    integration: getMarketplaceIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })

  api.addRoutes({
    path: '/espn-info',
    methods: [HttpMethod.GET],
    integration: getEspnInfoIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })
}
