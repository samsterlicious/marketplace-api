import { HttpApi, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha'
import { HttpJwtAuthorizer } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha'
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha'
import { OutcomeLambdas } from '../../lambdas/outcome'

export function createOutcomeeResource(
  api: HttpApi,
  functions: OutcomeLambdas,
  authorizer: HttpJwtAuthorizer,
) {
  const getOutcomesByUserIntegration = new HttpLambdaIntegration(
    'OutcomesByUserIntegration',
    functions.getByUser,
  )

  const getOutcomesRankingsIntegration = new HttpLambdaIntegration(
    'OutcomesRankingsIntegration',
    functions.getRankings,
  )

  api.addRoutes({
    path: '/leaderboard',
    methods: [HttpMethod.GET],
    integration: getOutcomesRankingsIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })

  api.addRoutes({
    path: '/my-outcomes',
    methods: [HttpMethod.GET],
    integration: getOutcomesByUserIntegration,
    authorizer,
    authorizationScopes: ['openid'],
  })
}
