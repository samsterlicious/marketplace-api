import {
  CorsHttpMethod,
  DomainName,
  HttpApi,
} from '@aws-cdk/aws-apigatewayv2-alpha'
import { HttpJwtAuthorizer } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha'
import { Duration } from 'aws-cdk-lib'
import { Cors } from 'aws-cdk-lib/aws-apigateway'
import {
  Certificate,
  CertificateValidation,
} from 'aws-cdk-lib/aws-certificatemanager'
import { ARecord, HostedZone, RecordTarget } from 'aws-cdk-lib/aws-route53'
import { ApiGatewayv2DomainProperties } from 'aws-cdk-lib/aws-route53-targets'
import { Construct } from 'constructs'
import { Config } from '../../backend-stack'
import { Lambdas } from '../lambdas'
import { createBetResource } from './routes/bet'
import { createBidResource } from './routes/bid'
import { createLeagueResource } from './routes/leagueRoute'
import { createMarketplaceResource } from './routes/marketplace'
import { createOutcomeeResource } from './routes/outcomeRoute'
import { createUserResource } from './routes/user'

export function createApi(
  scope: Construct,
  lambdas: Lambdas,
  context: Config,
): HttpApi {
  const domainName = `${context.domain.prefix}.${context.domain.root}`

  const zone = HostedZone.fromLookup(scope, 'hostedZone', {
    domainName: context.domain.root,
  })

  const certificate = new Certificate(scope, 'certificate', {
    domainName,
    validation: CertificateValidation.fromDns(zone),
  })

  const domain = new DomainName(scope, 'domainName', {
    domainName,
    certificate,
  })

  const api = new HttpApi(scope, 'Api', {
    corsPreflight: {
      allowOrigins: Cors.ALL_ORIGINS,
      allowMethods: [
        CorsHttpMethod.GET,
        CorsHttpMethod.HEAD,
        CorsHttpMethod.OPTIONS,
        CorsHttpMethod.POST,
        CorsHttpMethod.PUT,
      ],
      allowHeaders: Cors.DEFAULT_HEADERS,
      maxAge: Duration.days(10),
    },
    defaultDomainMapping: {
      domainName: domain,
    },
  })

  new ARecord(scope, 'alias', {
    zone,
    recordName: context.domain.prefix,
    target: RecordTarget.fromAlias(
      new ApiGatewayv2DomainProperties(
        domain.regionalDomainName,
        domain.regionalHostedZoneId,
      ),
    ),
  })

  const authorizer = new HttpJwtAuthorizer('authorizer', context.auth0.issuer, {
    jwtAudience: context.auth0.audience,
  })

  createBetResource(api, lambdas.bet, authorizer)
  createBidResource(api, lambdas.bid, authorizer)
  createMarketplaceResource(api, lambdas.marketplace, authorizer)
  createLeagueResource(api, lambdas.league, authorizer)
  createOutcomeeResource(api, lambdas.outcome, authorizer)
  createUserResource(api, lambdas.user, authorizer)
  return api
}
