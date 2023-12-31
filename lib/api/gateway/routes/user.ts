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

  const updateUsernameIntegration = new HttpLambdaIntegration(
    'UpdateUsernameIntegration',
    functions.updateName,
  )

  api.addRoutes({
    path: '/user',
    methods: [HttpMethod.GET],
    integration: getUserInfoIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })

  api.addRoutes({
    path: '/user',
    methods: [HttpMethod.PUT],
    integration: updateUsernameIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })
}
