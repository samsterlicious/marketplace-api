import { Duration } from 'aws-cdk-lib'
import { Table } from 'aws-cdk-lib/aws-dynamodb'
import { Runtime } from 'aws-cdk-lib/aws-lambda'
import { NodejsFunction } from 'aws-cdk-lib/aws-lambda-nodejs'
import { Construct } from 'constructs'

export function createBetLambdas(scope: Construct, table: Table): BetLambdas {
  const createBet = new NodejsFunction(scope, 'createBet', {
    entry: 'src/bet/create.ts',
    runtime: Runtime.NODEJS_18_X,
    timeout: Duration.seconds(5),
    environment: { TABLE_NAME: table.tableName },
  })

  table.grantWriteData(createBet)

  return {
    create: createBet,
  }
}

export type BetLambdas = {
  create: NodejsFunction
}
