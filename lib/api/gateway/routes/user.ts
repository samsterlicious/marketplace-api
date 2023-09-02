import { HttpApi, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha'
import { HttpJwtAuthorizer } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha'
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha'
import { UserLambdas } from '../../lambdas/user.'

export function createUserResource(
  api: HttpApi,
  functions: UserLambdas,
  authorizer: HttpJwtAuthorizer,
) {
  const getUserInfoIntegration = new HttpLambdaIntegration(
    'GetUserInfoIntegration',
    functions.get,
  )

  api.addRoutes({
    path: '/user',
    methods: [HttpMethod.GET],
    integration: getUserInfoIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })
}
