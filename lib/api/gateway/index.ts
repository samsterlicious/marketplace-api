import { Cors, RestApi } from 'aws-cdk-lib/aws-apigateway'
import {
  Certificate,
  CertificateValidation,
} from 'aws-cdk-lib/aws-certificatemanager'
import { ARecord, HostedZone, RecordTarget } from 'aws-cdk-lib/aws-route53'
import { ApiGateway } from 'aws-cdk-lib/aws-route53-targets'
import { Construct } from 'constructs'
import { Config } from '../../backend-stack'
import { Lambdas } from '../lambdas'
import { createBetResource } from './routes/bet'
import { createMarketplaceResource } from './routes/marketplace'

export function createApi(
  scope: Construct,
  lambdas: Lambdas,
  context: Config,
): RestApi {
  const domainName = `${context.domain.prefix}.${context.domain.root}`

  const zone = HostedZone.fromLookup(scope, 'hostedZone', {
    domainName: context.domain.root,
  })

  const certificate = new Certificate(scope, 'certificate', {
    domainName,
    validation: CertificateValidation.fromDns(zone),
  })

  const api = new RestApi(scope, 'api', {
    defaultCorsPreflightOptions: {
      allowOrigins: Cors.ALL_ORIGINS,
      allowMethods: Cors.ALL_METHODS,
      allowHeaders: Cors.DEFAULT_HEADERS,
    },
    domainName: {
      domainName,
      certificate,
    },
  })

  new ARecord(scope, 'alias', {
    zone,
    recordName: context.domain.prefix,
    target: RecordTarget.fromAlias(new ApiGateway(api)),
  })

  createBetResource(api, lambdas.bet)
  createMarketplaceResource(api, lambdas.marketplace)

  return api
}
