package e2dynamodb

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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
	dy        *dynamodb.DynamoDB
}

// New NewDynamoDB
func New(tableName string, sess *session.Session) *DynamoDB {
	return &DynamoDB{
		tableName: aws.String(tableName),
		dy:        dynamodb.New(sess),
	}
}

type Key struct {
	Type  string
	Name  string
	Value interface{}
}

func (d *DynamoDB) BuildKeyValue(k *Key) *dynamodb.AttributeValue {
	switch k.Type {
	case KeyTypeBinary:
		return &dynamodb.AttributeValue{B: k.Value.([]byte)}
	case KeyTypeBool:
		return &dynamodb.AttributeValue{BOOL: aws.Bool(k.Value.(bool))}
	case KeyTypeBinaryArray:
		return &dynamodb.AttributeValue{BS: k.Value.([][]byte)}
	case KeyTypeList:
		return &dynamodb.AttributeValue{L: k.Value.([]*dynamodb.AttributeValue)}
	case KeyTypeMap:
		return &dynamodb.AttributeValue{M: k.Value.(map[string]*dynamodb.AttributeValue)}
	case KeyTypeNumber:
		return &dynamodb.AttributeValue{N: aws.String(k.Value.(string))}
	case KeyTypeNumberArray:
		return &dynamodb.AttributeValue{NS: k.Value.([]*string)}
	case KeyTypeNull:
		return &dynamodb.AttributeValue{NULL: aws.Bool(k.Value.(bool))}
	case KeyTypeString:
		return &dynamodb.AttributeValue{S: aws.String(k.Value.(string))}
	case KeyTypeStringArray:
		return &dynamodb.AttributeValue{SS: k.Value.([]*string)}
	default:
		return &dynamodb.AttributeValue{S: aws.String(k.Value.(string))}
	}
}

// GetByPK 根據 PK 獲取數據
func (d *DynamoDB) GetByPK(partitionKey *Key, outputItem interface{}, opts ...*dynamodb.GetItemInput) error {
	return d.GetByPKAndSK(partitionKey, nil, outputItem, opts...)
}

// GetByPKAndSK 根據 PK 和 Sort Key 獲取數據
func (d *DynamoDB) GetByPKAndSK(partitionKey *Key, sortKey *Key, outputItem interface{}, opts ...*dynamodb.GetItemInput) error {
	if partitionKey == nil {
		return fmt.Errorf("partition key must not nil")
	}
	gi := &dynamodb.GetItemInput{}
	if len(opts) > 0 {
		gi = opts[0]
	}
	km := make(map[string]*dynamodb.AttributeValue)
	km[partitionKey.Name] = d.BuildKeyValue(partitionKey)
	if sortKey != nil {
		km[sortKey.Name] = d.BuildKeyValue(sortKey)
	}
	gi.TableName = d.tableName
	gi.Key = km
	out, err := d.dy.GetItem(gi)
	if err != nil {
		return err
	}
	if err := dynamodbattribute.UnmarshalMap(out.Item, outputItem); err != nil {
		return err
	}
	return nil
}

// Put 寫入一條數據
func (d *DynamoDB) Put(ar interface{}, opts ...*dynamodb.PutItemInput) error {
	av, err := dynamodbattribute.MarshalMap(ar)
	if err != nil {
		return err
	}
	pi := &dynamodb.PutItemInput{}
	if len(opts) > 0 {
		pi = opts[0]
	}
	pi.TableName = d.tableName
	pi.Item = av
	_, err = d.dy.PutItem(pi)
	return err
}

func (d *DynamoDB) DeleteByPKAndSK(partitionKey *Key, sortKey *Key, opts ...*dynamodb.DeleteItemInput) error {
	di := &dynamodb.DeleteItemInput{}
	if len(opts) > 0 {
		di = opts[0]
	}
	di.TableName = d.tableName
	km := make(map[string]*dynamodb.AttributeValue)
	km[partitionKey.Name] = d.BuildKeyValue(partitionKey)
	if sortKey != nil {
		km[sortKey.Name] = d.BuildKeyValue(sortKey)
	}
	di.Key = km
	_, err := d.dy.DeleteItem(di)

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
	return d.dy.ScanPages(si, fn)
}

func (d *DynamoDB) QueryPages(fn func(page *dynamodb.QueryOutput, lastPage bool) bool, opts ...*dynamodb.QueryInput) error {
	qi := &dynamodb.QueryInput{}
	if len(opts) > 0 {
		qi = opts[0]
	}
	qi.TableName = d.tableName
	return d.dy.QueryPages(qi, fn)
}
