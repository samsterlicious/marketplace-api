import { HttpApi, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha'
import { HttpJwtAuthorizer } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha'
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha'
import { LeagueLambdas } from '../../lambdas/league'

export function createLeagueResource(
  api: HttpApi,
  functions: LeagueLambdas,
  authorizer: HttpJwtAuthorizer,
) {
  const getUserInfoIntegration = new HttpLambdaIntegration(
    'GetUserInfoIntegration',
    functions.getUsers,
  )
  api.addRoutes({
    path: '/league/users/{league}',
    methods: [HttpMethod.GET],
    integration: getUserInfoIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })
}
