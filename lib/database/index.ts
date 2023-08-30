import { AttributeType, BillingMode, Table } from 'aws-cdk-lib/aws-dynamodb'
import { Construct } from 'constructs'

export function createDynamoDatabase(scope: Construct): Table {
  const table = new Table(scope, 'table', {
    partitionKey: { name: 'id', type: AttributeType.STRING },
    sortKey: { name: 'sortKey', type: AttributeType.STRING },
    billingMode: BillingMode.PROVISIONED,
    readCapacity: 1,
    writeCapacity: 1,
    timeToLiveAttribute: 'ttl',
  })

  table.addGlobalSecondaryIndex({
    indexName: 'gsi1',
    readCapacity: 1,
    writeCapacity: 1,
    partitionKey: { name: 'gsi1_id', type: AttributeType.STRING },
    sortKey: { name: 'gsi1_sortKey', type: AttributeType.STRING },
  })

  table.addGlobalSecondaryIndex({
    indexName: 'gsi2',
    readCapacity: 1,
    writeCapacity: 1,
    partitionKey: { name: 'gsi2_id', type: AttributeType.STRING },
    sortKey: { name: 'gsi2_sortKey', type: AttributeType.STRING },
  })

  table.addGlobalSecondaryIndex({
    indexName: 'gsi3',
    readCapacity: 1,
    writeCapacity: 1,
    partitionKey: { name: 'gsi3_id', type: AttributeType.STRING },
    sortKey: { name: 'gsi3_sortKey', type: AttributeType.STRING },
  })

  return table
}
