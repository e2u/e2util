package e2dynamodb

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type demoItem struct {
	PK   string `dynamodbav:"pk"`
	SK   string `dynamodbav:"sk"`
	Name string `dynamodbav:"name"`
}

type failAV struct{}

func (failAV) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return nil, assert.AnError
}

func testConfig() aws.Config {
	return aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
	}
}

func newTestDynamoDB(t *testing.T, handler http.HandlerFunc) *DynamoDB {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New("demo-table", testConfig(), func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(srv.URL)
		o.HTTPClient = srv.Client()
		o.Retryer = aws.NopRetryer{}
	})
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

func TestNew(t *testing.T) {
	d := New("demo", testConfig(), func(o *dynamodb.Options) {
		o.Region = "eu-west-1"
	})
	require.NotNil(t, d)
	assert.Equal(t, "demo", aws.ToString(d.tableName))
	assert.NotNil(t, d.client)
	assert.Equal(t, "eu-west-1", d.client.Options().Region)
}

func TestBuildKeyValue(t *testing.T) {
	d := &DynamoDB{}

	t.Run("string", func(t *testing.T) {
		av := d.BuildKeyValue(&Key{Type: KeyTypeString, Name: "pk", Value: "hello"})
		s, ok := av.(*types.AttributeValueMemberS)
		require.True(t, ok)
		assert.Equal(t, "hello", s.Value)
	})

	t.Run("default string", func(t *testing.T) {
		av := d.BuildKeyValue(&Key{Type: "unknown", Name: "pk", Value: "fallback"})
		s, ok := av.(*types.AttributeValueMemberS)
		require.True(t, ok)
		assert.Equal(t, "fallback", s.Value)
	})

	t.Run("number", func(t *testing.T) {
		av := d.BuildKeyValue(&Key{Type: KeyTypeNumber, Name: "n", Value: "42"})
		n, ok := av.(*types.AttributeValueMemberN)
		require.True(t, ok)
		assert.Equal(t, "42", n.Value)
	})

	t.Run("bool", func(t *testing.T) {
		av := d.BuildKeyValue(&Key{Type: KeyTypeBool, Name: "b", Value: true})
		b, ok := av.(*types.AttributeValueMemberBOOL)
		require.True(t, ok)
		assert.True(t, b.Value)
	})

	t.Run("binary", func(t *testing.T) {
		av := d.BuildKeyValue(&Key{Type: KeyTypeBinary, Name: "bin", Value: []byte("abc")})
		b, ok := av.(*types.AttributeValueMemberB)
		require.True(t, ok)
		assert.Equal(t, []byte("abc"), b.Value)
	})

	t.Run("binary array", func(t *testing.T) {
		av := d.BuildKeyValue(&Key{Type: KeyTypeBinaryArray, Name: "bs", Value: [][]byte{[]byte("a"), []byte("b")}})
		bs, ok := av.(*types.AttributeValueMemberBS)
		require.True(t, ok)
		assert.Equal(t, [][]byte{[]byte("a"), []byte("b")}, bs.Value)
	})

	t.Run("null", func(t *testing.T) {
		av := d.BuildKeyValue(&Key{Type: KeyTypeNull, Name: "n", Value: true})
		n, ok := av.(*types.AttributeValueMemberNULL)
		require.True(t, ok)
		assert.True(t, n.Value)
	})

	t.Run("string array from []string", func(t *testing.T) {
		av := d.BuildKeyValue(&Key{Type: KeyTypeStringArray, Name: "ss", Value: []string{"a", "b"}})
		ss, ok := av.(*types.AttributeValueMemberSS)
		require.True(t, ok)
		assert.Equal(t, []string{"a", "b"}, ss.Value)
	})

	t.Run("string array from []*string skips nil", func(t *testing.T) {
		av := d.BuildKeyValue(&Key{Type: KeyTypeStringArray, Name: "ss", Value: []*string{aws.String("a"), nil, aws.String("b")}})
		ss, ok := av.(*types.AttributeValueMemberSS)
		require.True(t, ok)
		assert.Equal(t, []string{"a", "b"}, ss.Value)
	})

	t.Run("number array from []string", func(t *testing.T) {
		av := d.BuildKeyValue(&Key{Type: KeyTypeNumberArray, Name: "ns", Value: []string{"1", "2"}})
		ns, ok := av.(*types.AttributeValueMemberNS)
		require.True(t, ok)
		assert.Equal(t, []string{"1", "2"}, ns.Value)
	})

	t.Run("number array from []*string", func(t *testing.T) {
		av := d.BuildKeyValue(&Key{Type: KeyTypeNumberArray, Name: "ns", Value: []*string{aws.String("1"), aws.String("2")}})
		ns, ok := av.(*types.AttributeValueMemberNS)
		require.True(t, ok)
		assert.Equal(t, []string{"1", "2"}, ns.Value)
	})

	t.Run("list and map", func(t *testing.T) {
		list := []types.AttributeValue{&types.AttributeValueMemberS{Value: "x"}}
		av := d.BuildKeyValue(&Key{Type: KeyTypeList, Name: "l", Value: list})
		l, ok := av.(*types.AttributeValueMemberL)
		require.True(t, ok)
		assert.Equal(t, list, l.Value)

		m := map[string]types.AttributeValue{"k": &types.AttributeValueMemberS{Value: "v"}}
		av = d.BuildKeyValue(&Key{Type: KeyTypeMap, Name: "m", Value: m})
		mm, ok := av.(*types.AttributeValueMemberM)
		require.True(t, ok)
		assert.Equal(t, m, mm.Value)
	})

	t.Run("string set panics on unsupported type", func(t *testing.T) {
		assert.Panics(t, func() {
			d.BuildKeyValue(&Key{Type: KeyTypeStringArray, Name: "ss", Value: 123})
		})
	})
}

func TestStringSet(t *testing.T) {
	assert.Equal(t, []string{"a"}, stringSet([]string{"a"}))
	assert.Equal(t, []string{"a"}, stringSet([]*string{aws.String("a"), nil}))
	assert.Panics(t, func() { stringSet(true) })
}

func TestGetByPK(t *testing.T) {
	t.Run("nil partition key", func(t *testing.T) {
		d := New("demo", testConfig())
		err := d.GetByPK(nil, &demoItem{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "partition key must not nil")
	})

	t.Run("unmarshals item", func(t *testing.T) {
		var gotBody map[string]any
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DynamoDB_20120810.GetItem", r.Header.Get("X-Amz-Target"))
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
			writeJSON(w, http.StatusOK, `{"Item":{"pk":{"S":"user-1"},"name":{"S":"alice"}}}`)
		})

		var item demoItem
		err := d.GetByPK(&Key{Type: KeyTypeString, Name: "pk", Value: "user-1"}, &item)
		require.NoError(t, err)
		assert.Equal(t, "user-1", item.PK)
		assert.Equal(t, "alice", item.Name)
		assert.Equal(t, "demo-table", gotBody["TableName"])
		key := gotBody["Key"].(map[string]any)
		assert.Equal(t, "user-1", key["pk"].(map[string]any)["S"])
		assert.Nil(t, key["sk"])
	})

	t.Run("passes extra get options", func(t *testing.T) {
		var gotBody map[string]any
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
			writeJSON(w, http.StatusOK, `{"Item":{"pk":{"S":"user-1"}}}`)
		})
		var item demoItem
		err := d.GetByPK(&Key{Type: KeyTypeString, Name: "pk", Value: "user-1"}, &item, &dynamodb.GetItemInput{
			ConsistentRead: aws.Bool(true),
		})
		require.NoError(t, err)
		assert.Equal(t, true, gotBody["ConsistentRead"])
	})

	t.Run("api error", func(t *testing.T) {
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"ResourceNotFoundException","message":"missing"}`)
		})
		err := d.GetByPK(&Key{Type: KeyTypeString, Name: "pk", Value: "x"}, &demoItem{})
		require.Error(t, err)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, `{"Item":{"pk":{"S":"user-1"}}}`)
		})
		err := d.GetByPK(&Key{Type: KeyTypeString, Name: "pk", Value: "user-1"}, demoItem{})
		require.Error(t, err)
	})
}

func TestGetByPKAndSK(t *testing.T) {
	var gotBody map[string]any
	d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		writeJSON(w, http.StatusOK, `{"Item":{"pk":{"S":"user-1"},"sk":{"S":"profile"},"name":{"S":"bob"}}}`)
	})

	var item demoItem
	err := d.GetByPKAndSK(
		&Key{Type: KeyTypeString, Name: "pk", Value: "user-1"},
		&Key{Type: KeyTypeString, Name: "sk", Value: "profile"},
		&item,
	)
	require.NoError(t, err)
	assert.Equal(t, "bob", item.Name)
	key := gotBody["Key"].(map[string]any)
	assert.Equal(t, "user-1", key["pk"].(map[string]any)["S"])
	assert.Equal(t, "profile", key["sk"].(map[string]any)["S"])
}

func TestPut(t *testing.T) {
	t.Run("writes marshaled item without extra options", func(t *testing.T) {
		var gotBody map[string]any
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
			writeJSON(w, http.StatusOK, `{}`)
		})
		err := d.Put(demoItem{PK: "user-1", Name: "carol"})
		require.NoError(t, err)
		assert.Equal(t, "demo-table", gotBody["TableName"])
		assert.Nil(t, gotBody["ConditionExpression"])
	})

	t.Run("writes marshaled item", func(t *testing.T) {
		var gotBody map[string]any
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DynamoDB_20120810.PutItem", r.Header.Get("X-Amz-Target"))
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
			writeJSON(w, http.StatusOK, `{}`)
		})
		err := d.Put(demoItem{PK: "user-1", SK: "profile", Name: "alice"}, &dynamodb.PutItemInput{
			ConditionExpression: aws.String("attribute_not_exists(pk)"),
		})
		require.NoError(t, err)
		assert.Equal(t, "demo-table", gotBody["TableName"])
		assert.Equal(t, "attribute_not_exists(pk)", gotBody["ConditionExpression"])
		item := gotBody["Item"].(map[string]any)
		assert.Equal(t, "user-1", item["pk"].(map[string]any)["S"])
		assert.Equal(t, "alice", item["name"].(map[string]any)["S"])
	})

	t.Run("marshal error", func(t *testing.T) {
		d := New("demo", testConfig())
		err := d.Put(struct {
			Bad failAV `dynamodbav:"bad"`
		}{})
		require.Error(t, err)
	})

	t.Run("api error", func(t *testing.T) {
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"ValidationException","message":"bad item"}`)
		})
		err := d.Put(demoItem{PK: "user-1"})
		require.Error(t, err)
	})
}

func TestDeleteByPKAndSK(t *testing.T) {
	t.Run("with sort key and options", func(t *testing.T) {
		var gotBody map[string]any
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DynamoDB_20120810.DeleteItem", r.Header.Get("X-Amz-Target"))
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
			writeJSON(w, http.StatusOK, `{}`)
		})
		err := d.DeleteByPKAndSK(
			&Key{Type: KeyTypeString, Name: "pk", Value: "user-1"},
			&Key{Type: KeyTypeString, Name: "sk", Value: "profile"},
			&dynamodb.DeleteItemInput{ConditionExpression: aws.String("attribute_exists(pk)")},
		)
		require.NoError(t, err)
		assert.Equal(t, "demo-table", gotBody["TableName"])
		assert.Equal(t, "attribute_exists(pk)", gotBody["ConditionExpression"])
		key := gotBody["Key"].(map[string]any)
		assert.Equal(t, "user-1", key["pk"].(map[string]any)["S"])
		assert.Equal(t, "profile", key["sk"].(map[string]any)["S"])
	})

	t.Run("without sort key", func(t *testing.T) {
		var gotBody map[string]any
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
			writeJSON(w, http.StatusOK, `{}`)
		})
		err := d.DeleteByPK(&Key{Type: KeyTypeString, Name: "pk", Value: "user-1"})
		require.NoError(t, err)
		key := gotBody["Key"].(map[string]any)
		assert.Equal(t, "user-1", key["pk"].(map[string]any)["S"])
		assert.Nil(t, key["sk"])
	})

	t.Run("api error", func(t *testing.T) {
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"ResourceNotFoundException","message":"missing"}`)
		})
		err := d.DeleteByPK(&Key{Type: KeyTypeString, Name: "pk", Value: "missing"})
		require.Error(t, err)
	})
}

func TestScanPages(t *testing.T) {
	t.Run("paginates until last page", func(t *testing.T) {
		var calls atomic.Int32
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DynamoDB_20120810.Scan", r.Header.Get("X-Amz-Target"))
			body, _ := io.ReadAll(r.Body)
			n := calls.Add(1)
			if n == 1 {
				assert.NotContains(t, string(body), "ExclusiveStartKey")
				writeJSON(w, http.StatusOK, `{"Items":[{"pk":{"S":"1"}}],"Count":1,"LastEvaluatedKey":{"pk":{"S":"1"}}}`)
				return
			}
			assert.Contains(t, string(body), "ExclusiveStartKey")
			writeJSON(w, http.StatusOK, `{"Items":[{"pk":{"S":"2"}}],"Count":1}`)
		})

		var pages []bool
		var count int
		err := d.ScanPages(func(page *dynamodb.ScanOutput, lastPage bool) bool {
			count += len(page.Items)
			pages = append(pages, lastPage)
			return true
		}, &dynamodb.ScanInput{Limit: aws.Int32(1)})
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Equal(t, []bool{false, true}, pages)
		assert.Equal(t, int32(2), calls.Load())
	})

	t.Run("stops when callback returns false", func(t *testing.T) {
		var calls atomic.Int32
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			writeJSON(w, http.StatusOK, `{"Items":[{"pk":{"S":"1"}}],"Count":1,"LastEvaluatedKey":{"pk":{"S":"1"}}}`)
		})
		err := d.ScanPages(func(page *dynamodb.ScanOutput, lastPage bool) bool {
			assert.False(t, lastPage)
			assert.Len(t, page.Items, 1)
			return false
		})
		require.NoError(t, err)
		assert.Equal(t, int32(1), calls.Load())
	})

	t.Run("api error", func(t *testing.T) {
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"ValidationException","message":"bad scan"}`)
		})
		err := d.ScanPages(func(*dynamodb.ScanOutput, bool) bool { return true })
		require.Error(t, err)
	})
}

func TestQueryPages(t *testing.T) {
	t.Run("paginates until last page", func(t *testing.T) {
		var calls atomic.Int32
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DynamoDB_20120810.Query", r.Header.Get("X-Amz-Target"))
			body, _ := io.ReadAll(r.Body)
			n := calls.Add(1)
			if n == 1 {
				assert.Contains(t, string(body), "pk = :pk")
				writeJSON(w, http.StatusOK, `{"Items":[{"pk":{"S":"user-1"},"sk":{"S":"a"}}],"Count":1,"LastEvaluatedKey":{"pk":{"S":"user-1"},"sk":{"S":"a"}}}`)
				return
			}
			writeJSON(w, http.StatusOK, `{"Items":[{"pk":{"S":"user-1"},"sk":{"S":"b"}}],"Count":1}`)
		})

		var lastFlags []bool
		err := d.QueryPages(func(page *dynamodb.QueryOutput, lastPage bool) bool {
			lastFlags = append(lastFlags, lastPage)
			return true
		}, &dynamodb.QueryInput{
			KeyConditionExpression: aws.String("pk = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "user-1"},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, []bool{false, true}, lastFlags)
	})

	t.Run("stops when callback returns false", func(t *testing.T) {
		var calls atomic.Int32
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			writeJSON(w, http.StatusOK, `{"Items":[{"pk":{"S":"1"}}],"LastEvaluatedKey":{"pk":{"S":"1"}}}`)
		})
		err := d.QueryPages(func(*dynamodb.QueryOutput, bool) bool { return false })
		require.NoError(t, err)
		assert.Equal(t, int32(1), calls.Load())
	})

	t.Run("api error", func(t *testing.T) {
		d := newTestDynamoDB(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"ValidationException","message":"bad query"}`)
		})
		err := d.QueryPages(func(*dynamodb.QueryOutput, bool) bool { return true })
		require.Error(t, err)
	})
}
