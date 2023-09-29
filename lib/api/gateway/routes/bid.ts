import { HttpApi, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha'
import { HttpJwtAuthorizer } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha'
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha'
import { BidLambdas } from '../../lambdas/bid'

export function createBidResource(
  api: HttpApi,
  functions: BidLambdas,
  authorizer: HttpJwtAuthorizer,
) {
  const createIntegration = new HttpLambdaIntegration(
    'CreateIntegration',
    functions.create,
  )

  const getByEventIntegration = new HttpLambdaIntegration(
    'GetByEventIntegration',
    functions.getByEvent,
  )

  const getByUserIntegration = new HttpLambdaIntegration(
    'GetByUserIntegration',
    functions.getByUser,
  )

  const updateBidIntegration = new HttpLambdaIntegration(
    'UpdateBidIntegration',
    functions.update,
  )

  api.addRoutes({
    path: '/bid',
    methods: [HttpMethod.POST],
    integration: createIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })

  api.addRoutes({
    path: '/bid/event/{event}',
    methods: [HttpMethod.GET],
    integration: getByEventIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })

  api.addRoutes({
    path: '/bid',
    methods: [HttpMethod.GET],
    integration: getByUserIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })

  api.addRoutes({
    path: '/bid',
    methods: [HttpMethod.PUT],
    integration: updateBidIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })
}
