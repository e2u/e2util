package e2dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	KeyTypeBinary      = "B"
	KeyTypeBool        = "BOOL"
	KeyTypeBinaryArray = "BS"
	KeyTypeList        = "L"
	KeyTypeMap         = "M"
	KeyTypeNumber      = "N"
	KeyTypeNumberArray = "NS"
	KeyTypeNull        = "NULL"
	KeyTypeString      = "S"
	KeyTypeStringArray = "SS"
)

type DynamoDB struct {
	tableName *string
	client    *dynamodb.Client
}

// New NewDynamoDB
func New(tableName string, cfg aws.Config, optFns ...func(*dynamodb.Options)) *DynamoDB {
	return &DynamoDB{
		tableName: aws.String(tableName),
		client:    dynamodb.NewFromConfig(cfg, optFns...),
	}
}

type Key struct {
	Type  string
	Name  string
	Value any
}

func (d *DynamoDB) BuildKeyValue(k *Key) types.AttributeValue {
	switch k.Type {
	case KeyTypeBinary:
		return &types.AttributeValueMemberB{Value: k.Value.([]byte)}
	case KeyTypeBool:
		return &types.AttributeValueMemberBOOL{Value: k.Value.(bool)}
	case KeyTypeBinaryArray:
		return &types.AttributeValueMemberBS{Value: k.Value.([][]byte)}
	case KeyTypeList:
		return &types.AttributeValueMemberL{Value: k.Value.([]types.AttributeValue)}
	case KeyTypeMap:
		return &types.AttributeValueMemberM{Value: k.Value.(map[string]types.AttributeValue)}
	case KeyTypeNumber:
		return &types.AttributeValueMemberN{Value: k.Value.(string)}
	case KeyTypeNumberArray:
		return &types.AttributeValueMemberNS{Value: stringSet(k.Value)}
	case KeyTypeNull:
		return &types.AttributeValueMemberNULL{Value: k.Value.(bool)}
	case KeyTypeString:
		return &types.AttributeValueMemberS{Value: k.Value.(string)}
	case KeyTypeStringArray:
		return &types.AttributeValueMemberSS{Value: stringSet(k.Value)}
	default:
		return &types.AttributeValueMemberS{Value: k.Value.(string)}
	}
}

func stringSet(v any) []string {
	switch x := v.(type) {
	case []string:
		return x
	case []*string:
		out := make([]string, 0, len(x))
		for _, s := range x {
			if s != nil {
				out = append(out, *s)
			}
		}
		return out
	default:
		return v.([]string)
	}
}

// GetByPK 根據 PK 獲取數據
func (d *DynamoDB) GetByPK(partitionKey *Key, outputItem any, opts ...*dynamodb.GetItemInput) error {
	return d.GetByPKAndSK(partitionKey, nil, outputItem, opts...)
}

// GetByPKAndSK 根據 PK 和 Sort Key 獲取數據
func (d *DynamoDB) GetByPKAndSK(partitionKey *Key, sortKey *Key, outputItem any, opts ...*dynamodb.GetItemInput) error {
	if partitionKey == nil {
		return fmt.Errorf("partition key must not nil")
	}
	gi := &dynamodb.GetItemInput{}
	if len(opts) > 0 {
		gi = opts[0]
	}
	km := make(map[string]types.AttributeValue)
	km[partitionKey.Name] = d.BuildKeyValue(partitionKey)
	if sortKey != nil {
		km[sortKey.Name] = d.BuildKeyValue(sortKey)
	}
	gi.TableName = d.tableName
	gi.Key = km
	out, err := d.client.GetItem(context.Background(), gi)
	if err != nil {
		return err
	}
	if err := attributevalue.UnmarshalMap(out.Item, outputItem); err != nil {
		return err
	}
	return nil
}

// Put 寫入一條數據
func (d *DynamoDB) Put(ar any, opts ...*dynamodb.PutItemInput) error {
	av, err := attributevalue.MarshalMap(ar)
	if err != nil {
		return err
	}
	pi := &dynamodb.PutItemInput{}
	if len(opts) > 0 {
		pi = opts[0]
	}
	pi.TableName = d.tableName
	pi.Item = av
	_, err = d.client.PutItem(context.Background(), pi)
	return err
}

func (d *DynamoDB) DeleteByPKAndSK(partitionKey *Key, sortKey *Key, opts ...*dynamodb.DeleteItemInput) error {
	di := &dynamodb.DeleteItemInput{}
	if len(opts) > 0 {
		di = opts[0]
	}
	di.TableName = d.tableName
	km := make(map[string]types.AttributeValue)
	km[partitionKey.Name] = d.BuildKeyValue(partitionKey)
	if sortKey != nil {
		km[sortKey.Name] = d.BuildKeyValue(sortKey)
	}
	di.Key = km
	_, err := d.client.DeleteItem(context.Background(), di)

	return err
}

func (d *DynamoDB) DeleteByPK(partitionKey *Key, opts ...*dynamodb.DeleteItemInput) error {
	return d.DeleteByPKAndSK(partitionKey, nil, opts...)
}

func (d *DynamoDB) ScanPages(fn func(page *dynamodb.ScanOutput, lastPage bool) bool, opts ...*dynamodb.ScanInput) error {
	si := &dynamodb.ScanInput{}
	if len(opts) > 0 {
		si = opts[0]
	}
	si.TableName = d.tableName
	p := dynamodb.NewScanPaginator(d.client, si)
	for p.HasMorePages() {
		page, err := p.NextPage(context.Background())
		if err != nil {
			return err
		}
		if !fn(page, !p.HasMorePages()) {
			return nil
		}
	}
	return nil
}

func (d *DynamoDB) QueryPages(fn func(page *dynamodb.QueryOutput, lastPage bool) bool, opts ...*dynamodb.QueryInput) error {
	qi := &dynamodb.QueryInput{}
	if len(opts) > 0 {
		qi = opts[0]
	}
	qi.TableName = d.tableName
	p := dynamodb.NewQueryPaginator(d.client, qi)
	for p.HasMorePages() {
		page, err := p.NextPage(context.Background())
		if err != nil {
			return err
		}
		if !fn(page, !p.HasMorePages()) {
			return nil
		}
	}
	return nil
}
