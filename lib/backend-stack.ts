import * as cdk from 'aws-cdk-lib'
import { Construct } from 'constructs'
import { createApi } from './api/gateway'
import { createLambdas } from './api/lambdas'
import { createDynamoDatabase } from './database'
// import * as sqs from 'aws-cdk-lib/aws-sqs';

export class BackendStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props)
    const environment = this.node.tryGetContext('config')
    if (!environment) {
      throw new Error(
        'Environment variable must be passed to cdk: ` cdk -c config=XXX`',
      )
    }

    const config = this.node.getContext(environment)
    const table = createDynamoDatabase(this)

    const lambdas = createLambdas(this, { table })
    createApi(this, lambdas, config)
  }
}

export type Config = {
  auth0: {
    audience: string[]
    issuer: string
  }
  domain: {
    prefix: string
    root: string
  }
}
