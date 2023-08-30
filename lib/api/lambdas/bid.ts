import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha'
import { Construct } from 'constructs'
import { CreateLambdaParams, LambdaConfig } from '.'

export function createBidLambdas(
  scope: Construct,
  config: LambdaConfig,
  params: CreateLambdaParams,
): BidLambdas {
  const createBid = new GoFunction(scope, 'createBidLambda', {
    entry: 'src/main/bid/create',
    ...config,
  })

  const getByEvent = new GoFunction(scope, 'getByEventLambda', {
    entry: 'src/main/bid/getByEvent',
    ...config,
  })

  const getByUser = new GoFunction(scope, 'getByUserLambda', {
    entry: 'src/main/bid/getByUser',
    ...config,
  })

  params.table.grantReadWriteData(createBid)
  params.table.grantReadWriteData(getByEvent)
  params.table.grantReadWriteData(getByUser)

  return {
    create: createBid,
    getByEvent,
    getByUser,
  }
}

export type BidLambdas = {
  create: GoFunction
  getByEvent: GoFunction
  getByUser: GoFunction
}
