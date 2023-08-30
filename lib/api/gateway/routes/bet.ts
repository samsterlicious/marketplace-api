import { HttpApi, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha'
import { HttpJwtAuthorizer } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha'
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha'
import { BetLambdas } from '../../lambdas/bet'

export function createBetResource(
  api: HttpApi,
  functions: BetLambdas,
  authorizer: HttpJwtAuthorizer,
) {
  const getIntegration = new HttpLambdaIntegration(
    'GetBetsIntegration',
    functions.get,
  )

  api.addRoutes({
    path: '/bet',
    methods: [HttpMethod.GET],
    integration: getIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })

  api.addRoutes({
    path: '/bet/date/{date}',
    methods: [HttpMethod.GET],
    integration: getIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })
}
