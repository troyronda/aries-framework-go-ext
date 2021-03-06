/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// Package newstorage contains common tests for newstorage implementation.
//
package newstorage

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/newstorage"
	"github.com/stretchr/testify/require"
)

// TestAll tests common storage functionality.
func TestAll(t *testing.T, provider newstorage.Provider) {
	t.Helper()

	t.Run("Store Put and Get", func(t *testing.T) {
		TestPutGet(t, provider)
	})
	t.Run("Store GetBulk and Get", func(t *testing.T) {
		TestGetBulk(t, provider)
	})
	t.Run("Delete", func(t *testing.T) {
		TestDelete(t, provider)
	})
	t.Run("Query", func(t *testing.T) {
		TestQuery(t, provider)
	})
	t.Run("Batch", func(t *testing.T) {
		TestBatch(t, provider)
	})
}

// TestPutGet tests common Put and Get functionality.
func TestPutGet(t *testing.T, provider newstorage.Provider) {
	t.Helper()

	const commonKey = "did:example:1"

	data := []byte("value1")

	// Create two different stores for testing.
	store1name := randomStoreName()
	store1, err := provider.OpenStore(store1name)
	require.NoError(t, err)

	store2, err := provider.OpenStore(randomStoreName())
	require.NoError(t, err)

	// Put in store 1.
	err = store1.Put(commonKey, data)
	require.NoError(t, err)

	// Try getting from store 1 - should be found.
	doc, err := store1.Get(commonKey)
	require.NoError(t, err)
	require.NotEmpty(t, doc)
	require.Equal(t, data, doc)

	// Try getting from store 2 - should not be found
	doc, err = store2.Get(commonKey)
	require.True(t, errors.Is(err, newstorage.ErrDataNotFound), "got unexpected error or no error")
	require.Nil(t, doc)

	// Put in store 2.
	err = store2.Put(commonKey, data)
	require.NoError(t, err)

	// Now we should be able to get that value from store 2.
	doc, err = store2.Get(commonKey)
	require.NoError(t, err)
	require.NotEmpty(t, doc)
	require.Equal(t, data, doc)

	// Create store 3 with the same name as store 1.
	store3, err := provider.OpenStore(store1name)
	require.NoError(t, err)
	require.NotNil(t, store3)

	// Since store 3 points to the same underlying database as store 1, the data should be found.
	doc, err = store3.Get(commonKey)
	require.NoError(t, err)
	require.NotEmpty(t, doc)
	require.Equal(t, data, doc)

	tryNilOrBlankValues(t, store1, data, commonKey)
}

// TestGetBulk tests common GetBulk functionality.
func TestGetBulk(t *testing.T, provider newstorage.Provider) { //nolint: funlen // Test file
	t.Run("Success: all values found", func(t *testing.T) {
		store, err := provider.OpenStore(randomStoreName())
		require.NoError(t, err)
		require.NotNil(t, store)

		err = store.Put("key1", []byte("value1"),
			[]newstorage.Tag{
				{Name: "tagName1", Value: "tagValue1"},
				{Name: "tagName2", Value: "tagValue2"},
			}...)
		require.NoError(t, err)

		err = store.Put("key2", []byte("value2"),
			[]newstorage.Tag{
				{Name: "tagName1", Value: "tagValue1"},
				{Name: "tagName2", Value: "tagValue2"},
			}...)
		require.NoError(t, err)

		values, err := store.GetBulk("key1", "key2")
		require.NoError(t, err)
		require.Len(t, values, 2)
		require.Equal(t, "value1", string(values[0]))
		require.Equal(t, "value2", string(values[1]))
	})
	t.Run("Success: one value found, one not", func(t *testing.T) {
		store, err := provider.OpenStore(randomStoreName())
		require.NoError(t, err)
		require.NotNil(t, store)

		err = store.Put("key1", []byte("value1"),
			[]newstorage.Tag{
				{Name: "tagName1", Value: "tagValue1"},
				{Name: "tagName2", Value: "tagValue2"},
			}...)
		require.NoError(t, err)

		values, err := store.GetBulk("key1", "key2")
		require.NoError(t, err)
		require.Len(t, values, 2)
		require.Equal(t, "value1", string(values[0]))
		require.Nil(t, values[1])
	})
	t.Run("Success: one value found, one not because it was deleted", func(t *testing.T) {
		store, err := provider.OpenStore(randomStoreName())
		require.NoError(t, err)
		require.NotNil(t, store)

		err = store.Put("key1", []byte("value1"),
			[]newstorage.Tag{
				{Name: "tagName1", Value: "tagValue1"},
				{Name: "tagName2", Value: "tagValue2"},
			}...)
		require.NoError(t, err)

		err = store.Put("key2", []byte("value2"),
			[]newstorage.Tag{
				{Name: "tagName1", Value: "tagValue1"},
				{Name: "tagName2", Value: "tagValue2"},
			}...)
		require.NoError(t, err)

		err = store.Delete("key2")
		require.NoError(t, err)

		values, err := store.GetBulk("key1", "key2")
		require.NoError(t, err)
		require.Len(t, values, 2)
		require.Equal(t, "value1", string(values[0]))
		require.Nil(t, values[1])
	})
	t.Run("Success: no values found", func(t *testing.T) {
		store, err := provider.OpenStore(randomStoreName())
		require.NoError(t, err)
		require.NotNil(t, store)

		err = store.Put("key1", []byte("value1"),
			[]newstorage.Tag{
				{Name: "tagName1", Value: "tagValue1"},
				{Name: "tagName2", Value: "tagValue2"},
			}...)
		require.NoError(t, err)

		values, err := store.GetBulk("key3", "key4")
		require.NoError(t, err)
		require.Len(t, values, 2)
		require.Nil(t, values[0])
		require.Nil(t, values[1])
	})
	t.Run("Failure: keys string slice cannot be nil", func(t *testing.T) {
		store, err := provider.OpenStore(randomStoreName())
		require.NoError(t, err)
		require.NotNil(t, store)

		values, err := store.GetBulk(nil...)
		require.EqualError(t, err, "keys string slice cannot be nil")
		require.Nil(t, values)
	})
}

// TestDelete tests common Delete functionality.
func TestDelete(t *testing.T, provider newstorage.Provider) {
	t.Helper()

	const commonKey = "did:example:1234"

	data := []byte("value1")

	store, err := provider.OpenStore(randomStoreName())
	require.NoError(t, err)

	// Put in store 1
	err = store.Put(commonKey, data)
	require.NoError(t, err)

	// Try getting from store 1 - should be found.
	doc, err := store.Get(commonKey)
	require.NoError(t, err)
	require.NotEmpty(t, doc)
	require.Equal(t, data, doc)

	// Delete an existing key - should succeed.
	err = store.Delete(commonKey)
	require.NoError(t, err)

	// Delete a key which never existed. Should not throw any error.
	err = store.Delete("k1")
	require.NoError(t, err)

	// Try to get the value stored under the deleted key - should fail.
	doc, err = store.Get(commonKey)
	require.True(t, errors.Is(err, newstorage.ErrDataNotFound), "got unexpected error or no error")
	require.Empty(t, doc)

	// Try Delete with an blank key - should fail.
	err = store.Delete("")
	require.Error(t, err)
}

// TestQuery tests common Query functionality.
func TestQuery(t *testing.T, provider newstorage.Provider) { // nolint: funlen // Test file
	t.Run("Success - tag name only query - 2 values found", func(t *testing.T) {
		keysToPut := []string{"key1", "key2", "key3"}
		valuesToPut := [][]byte{[]byte("value1"), []byte("value2"), []byte("value3")}
		tagsToPut := [][]newstorage.Tag{
			{{Name: "tagName1", Value: "tagValue1"}, {Name: "tagName2", Value: "tagValue2"}},
			{{Name: "tagName3", Value: "tagValue"}, {Name: "tagName4"}},
			{{Name: "tagName3", Value: "tagValue2"}},
		}

		expectedKeys := []string{keysToPut[1], keysToPut[2]}
		expectedValues := [][]byte{valuesToPut[1], valuesToPut[2]}
		expectedTags := [][]newstorage.Tag{tagsToPut[1], tagsToPut[2]}

		queryExpression := "tagName3"

		t.Run("Default page setting", func(t *testing.T) {
			storeName := randomStoreName()

			store, err := provider.OpenStore(storeName)
			require.NoError(t, err)
			require.NotNil(t, store)

			err = provider.SetStoreConfig(storeName,
				newstorage.StoreConfiguration{TagNames: []string{"tagName1", "tagName2", "tagName3", "tagName4"}})
			require.NoError(t, err)

			putData(t, store, keysToPut, valuesToPut, tagsToPut)

			iterator, err := store.Query(queryExpression)
			require.NoError(t, err)

			verifyExpectedIterator(t, iterator, expectedKeys, expectedValues, expectedTags)
		})
		t.Run("Page size 2", func(t *testing.T) {
			storeName := randomStoreName()

			store, err := provider.OpenStore(storeName)
			require.NoError(t, err)
			require.NotNil(t, store)

			err = provider.SetStoreConfig(storeName,
				newstorage.StoreConfiguration{TagNames: []string{"tagName1", "tagName2", "tagName3", "tagName4"}})
			require.NoError(t, err)

			putData(t, store, keysToPut, valuesToPut, tagsToPut)

			//nolint:gomnd // Test file
			iterator, err := store.Query(queryExpression, newstorage.WithPageSize(2))
			require.NoError(t, err)

			verifyExpectedIterator(t, iterator, expectedKeys, expectedValues, expectedTags)
		})
		t.Run("Page size 1", func(t *testing.T) {
			storeName := randomStoreName()

			store, err := provider.OpenStore(storeName)
			require.NoError(t, err)
			require.NotNil(t, store)

			err = provider.SetStoreConfig(storeName,
				newstorage.StoreConfiguration{TagNames: []string{"tagName1", "tagName2", "tagName3", "tagName4"}})
			require.NoError(t, err)

			putData(t, store, keysToPut, valuesToPut, tagsToPut)

			iterator, err := store.Query(queryExpression, newstorage.WithPageSize(1))
			require.NoError(t, err)

			verifyExpectedIterator(t, iterator, expectedKeys, expectedValues, expectedTags)
		})
		t.Run("Page size 100", func(t *testing.T) {
			storeName := randomStoreName()

			store, err := provider.OpenStore(storeName)
			require.NoError(t, err)
			require.NotNil(t, store)

			err = provider.SetStoreConfig(storeName,
				newstorage.StoreConfiguration{TagNames: []string{"tagName1", "tagName2", "tagName3", "tagName4"}})
			require.NoError(t, err)

			putData(t, store, keysToPut, valuesToPut, tagsToPut)

			//nolint:gomnd // Test file
			iterator, err := store.Query(queryExpression, newstorage.WithPageSize(100))
			require.NoError(t, err)

			verifyExpectedIterator(t, iterator, expectedKeys, expectedValues, expectedTags)
		})
	})
	t.Run("Success - tag name only query - 0 values found", func(t *testing.T) {
		keysToPut := []string{"key1", "key2", "key3"}
		valuesToPut := [][]byte{[]byte("value1"), []byte("value2"), []byte("value3")}
		tagsToPut := [][]newstorage.Tag{
			{{Name: "tagName1", Value: "tagValue1"}, {Name: "tagName2", Value: "tagValue2"}},
			{{Name: "tagName3", Value: "tagValue"}, {Name: "tagName4"}},
			{{Name: "tagName3", Value: "tagValue2"}},
		}

		queryExpression := "nonExistentTagName"

		t.Run("Default page setting", func(t *testing.T) {
			storeName := randomStoreName()

			store, err := provider.OpenStore(storeName)
			require.NoError(t, err)
			require.NotNil(t, store)

			err = provider.SetStoreConfig(storeName,
				newstorage.StoreConfiguration{TagNames: []string{"tagName1", "tagName2", "tagName3", "tagName4"}})
			require.NoError(t, err)

			putData(t, store, keysToPut, valuesToPut, tagsToPut)

			iterator, err := store.Query(" ")
			require.NoError(t, err)

			verifyExpectedIterator(t, iterator, nil, nil, nil)
		})
		t.Run("Page size 2", func(t *testing.T) {
			storeName := randomStoreName()

			store, err := provider.OpenStore(storeName)
			require.NoError(t, err)
			require.NotNil(t, store)

			err = provider.SetStoreConfig(storeName,
				newstorage.StoreConfiguration{TagNames: []string{"tagName1", "tagName2", "tagName3", "tagName4"}})
			require.NoError(t, err)

			putData(t, store, keysToPut, valuesToPut, tagsToPut)

			//nolint:gomnd // Test file
			iterator, err := store.Query(queryExpression, newstorage.WithPageSize(2))
			require.NoError(t, err)

			verifyExpectedIterator(t, iterator, nil, nil, nil)
		})
		t.Run("Page size 1", func(t *testing.T) {
			storeName := randomStoreName()

			store, err := provider.OpenStore(storeName)
			require.NoError(t, err)
			require.NotNil(t, store)

			err = provider.SetStoreConfig(storeName,
				newstorage.StoreConfiguration{TagNames: []string{"tagName1", "tagName2", "tagName3", "tagName4"}})
			require.NoError(t, err)

			putData(t, store, keysToPut, valuesToPut, tagsToPut)

			iterator, err := store.Query(queryExpression, newstorage.WithPageSize(1))
			require.NoError(t, err)

			verifyExpectedIterator(t, iterator, nil, nil, nil)
		})
		t.Run("Page size 100", func(t *testing.T) {
			storeName := randomStoreName()

			store, err := provider.OpenStore(storeName)
			require.NoError(t, err)
			require.NotNil(t, store)

			err = provider.SetStoreConfig(storeName,
				newstorage.StoreConfiguration{TagNames: []string{"tagName1", "tagName2", "tagName3", "tagName4"}})
			require.NoError(t, err)

			putData(t, store, keysToPut, valuesToPut, tagsToPut)

			//nolint:gomnd // Test file
			iterator, err := store.Query(queryExpression, newstorage.WithPageSize(100))
			require.NoError(t, err)

			verifyExpectedIterator(t, iterator, nil, nil, nil)
		})
	})
	t.Run("Success - tag name and value query - 2 values found", func(t *testing.T) {
		keysToPut := []string{"key1", "key2", "key3", "key4"}
		valuesToPut := [][]byte{[]byte("value1"), []byte("value2"), []byte("value3"), []byte("value4")}
		tagsToPut := [][]newstorage.Tag{
			{{Name: "tagName1", Value: "tagValue1"}, {Name: "tagName2", Value: "tagValue2"}},
			{{Name: "tagName3", Value: "tagValue1"}, {Name: "tagName4"}},
			{{Name: "tagName3", Value: "tagValue2"}},
			{{Name: "tagName3", Value: "tagValue1"}},
		}

		expectedKeys := []string{keysToPut[1], keysToPut[3]}
		expectedValues := [][]byte{valuesToPut[1], valuesToPut[3]}
		expectedTags := [][]newstorage.Tag{tagsToPut[1], tagsToPut[3]}

		queryExpression := "tagName3:tagValue1"

		t.Run("Default page setting", func(t *testing.T) {
			storeName := randomStoreName()

			store, err := provider.OpenStore(storeName)
			require.NoError(t, err)
			require.NotNil(t, store)

			err = provider.SetStoreConfig(storeName,
				newstorage.StoreConfiguration{TagNames: []string{"tagName1", "tagName2", "tagName3", "tagName4"}})
			require.NoError(t, err)

			putData(t, store, keysToPut, valuesToPut, tagsToPut)

			iterator, err := store.Query(queryExpression)
			require.NoError(t, err)

			verifyExpectedIterator(t, iterator, expectedKeys, expectedValues, expectedTags)
		})
		t.Run("Page size 2", func(t *testing.T) {
			storeName := randomStoreName()

			store, err := provider.OpenStore(storeName)
			require.NoError(t, err)
			require.NotNil(t, store)

			err = provider.SetStoreConfig(storeName,
				newstorage.StoreConfiguration{TagNames: []string{"tagName1", "tagName2", "tagName3", "tagName4"}})
			require.NoError(t, err)

			putData(t, store, keysToPut, valuesToPut, tagsToPut)

			//nolint:gomnd // Test file
			iterator, err := store.Query(queryExpression, newstorage.WithPageSize(2))
			require.NoError(t, err)

			verifyExpectedIterator(t, iterator, expectedKeys, expectedValues, expectedTags)
		})
		t.Run("Page size 1", func(t *testing.T) {
			storeName := randomStoreName()

			store, err := provider.OpenStore(storeName)
			require.NoError(t, err)
			require.NotNil(t, store)

			err = provider.SetStoreConfig(storeName,
				newstorage.StoreConfiguration{TagNames: []string{"tagName1", "tagName2", "tagName3", "tagName4"}})
			require.NoError(t, err)

			putData(t, store, keysToPut, valuesToPut, tagsToPut)

			iterator, err := store.Query(queryExpression, newstorage.WithPageSize(1))
			require.NoError(t, err)

			verifyExpectedIterator(t, iterator, expectedKeys, expectedValues, expectedTags)
		})
		t.Run("Page size 100", func(t *testing.T) {
			storeName := randomStoreName()

			store, err := provider.OpenStore(storeName)
			require.NoError(t, err)
			require.NotNil(t, store)

			err = provider.SetStoreConfig(storeName,
				newstorage.StoreConfiguration{TagNames: []string{"tagName1", "tagName2", "tagName3", "tagName4"}})
			require.NoError(t, err)

			putData(t, store, keysToPut, valuesToPut, tagsToPut)

			//nolint:gomnd // Test file
			iterator, err := store.Query(queryExpression, newstorage.WithPageSize(100))
			require.NoError(t, err)

			verifyExpectedIterator(t, iterator, expectedKeys, expectedValues, expectedTags)
		})
	})
	t.Run("Failure - invalid expression formats", func(t *testing.T) {
		storeName := randomStoreName()

		store, err := provider.OpenStore(storeName)
		require.NoError(t, err)
		require.NotNil(t, store)

		t.Run("Blank expression", func(t *testing.T) {
			iterator, err := store.Query("")
			require.EqualError(t, err, "invalid expression format. it must be in the following format: TagName:TagValue")
			require.Empty(t, iterator)
		})
		t.Run("Too many colon-separated parts", func(t *testing.T) {
			iterator, err := store.Query("name:value:somethingElse")
			require.EqualError(t, err, "invalid expression format. it must be in the following format: TagName:TagValue")
			require.Empty(t, iterator)
		})
	})
}

// TestBatch tests common Batch functionality.
func TestBatch(t *testing.T, provider newstorage.Provider) { // nolint:funlen // Test file
	t.Run("Success: put three new values", func(t *testing.T) {
		store, err := provider.OpenStore(randomStoreName())
		require.NoError(t, err)
		require.NotNil(t, store)

		operations := []newstorage.Operation{
			{Key: "key1", Value: []byte("value1"), Tags: []newstorage.Tag{{Name: "tagName1"}}},
			{Key: "key2", Value: []byte("value2"), Tags: []newstorage.Tag{{Name: "tagName2"}}},
			{Key: "key3", Value: []byte("value3"), Tags: []newstorage.Tag{{Name: "tagName3"}}},
		}

		err = store.Batch(operations)
		require.NoError(t, err)

		// Check and make sure all values and tags were stored

		value, err := store.Get("key1")
		require.NoError(t, err)
		require.Equal(t, "value1", string(value))

		value, err = store.Get("key2")
		require.NoError(t, err)
		require.Equal(t, "value2", string(value))

		value, err = store.Get("key3")
		require.NoError(t, err)
		require.Equal(t, "value3", string(value))

		tags, err := store.GetTags("key1")
		require.NoError(t, err)
		require.Len(t, tags, 1)
		require.Equal(t, "tagName1", tags[0].Name)

		tags, err = store.GetTags("key2")
		require.NoError(t, err)
		require.Len(t, tags, 1)
		require.Equal(t, "tagName2", tags[0].Name)

		tags, err = store.GetTags("key3")
		require.NoError(t, err)
		require.Len(t, tags, 1)
		require.Equal(t, "tagName3", tags[0].Name)
	})
	t.Run("Success: update three different previously-stored values via Batch", func(t *testing.T) {
		store, err := provider.OpenStore(randomStoreName())
		require.NoError(t, err)
		require.NotNil(t, store)

		err = store.Put("key1", []byte("value1"), []newstorage.Tag{{Name: "tagName1", Value: "tagValue1"}}...)
		require.NoError(t, err)

		err = store.Put("key2", []byte("value2"), []newstorage.Tag{{Name: "tagName2", Value: "tagValue2"}}...)
		require.NoError(t, err)

		err = store.Put("key3", []byte("value3"), []newstorage.Tag{{Name: "tagName3", Value: "tagValue3"}}...)
		require.NoError(t, err)

		operations := []newstorage.Operation{
			{Key: "key1", Value: []byte("value1_new"), Tags: []newstorage.Tag{{Name: "tagName1"}}},
			{Key: "key2", Value: []byte("value2_new"), Tags: []newstorage.Tag{{Name: "tagName2_new", Value: "tagValue2"}}},
			{Key: "key3", Value: []byte("value3_new"), Tags: []newstorage.Tag{{Name: "tagName3_new", Value: "tagValue3_new"}}},
		}

		err = store.Batch(operations)
		require.NoError(t, err)

		// Check and make sure all values and tags were stored

		value, err := store.Get("key1")
		require.NoError(t, err)
		require.Equal(t, "value1_new", string(value))

		value, err = store.Get("key2")
		require.NoError(t, err)
		require.Equal(t, "value2_new", string(value))

		value, err = store.Get("key3")
		require.NoError(t, err)
		require.Equal(t, "value3_new", string(value))

		tags, err := store.GetTags("key1")
		require.NoError(t, err)
		require.Len(t, tags, 1)
		require.Equal(t, "tagName1", tags[0].Name)
		require.Equal(t, "", tags[0].Value)

		tags, err = store.GetTags("key2")
		require.NoError(t, err)
		require.Len(t, tags, 1)
		require.Equal(t, "tagName2_new", tags[0].Name)
		require.Equal(t, "tagValue2", tags[0].Value)

		tags, err = store.GetTags("key3")
		require.NoError(t, err)
		require.Len(t, tags, 1)
		require.Equal(t, "tagName3_new", tags[0].Name)
		require.Equal(t, "tagValue3_new", tags[0].Value)
	})
	t.Run("Success: Delete three different previously-stored values via Batch", func(t *testing.T) {
		store, err := provider.OpenStore(randomStoreName())
		require.NoError(t, err)
		require.NotNil(t, store)

		err = store.Put("key1", []byte("value1"), []newstorage.Tag{{Name: "tagName1", Value: "tagValue1"}}...)
		require.NoError(t, err)

		err = store.Put("key2", []byte("value2"), []newstorage.Tag{{Name: "tagName2", Value: "tagValue2"}}...)
		require.NoError(t, err)

		err = store.Put("key3", []byte("value3"), []newstorage.Tag{{Name: "tagName3", Value: "tagValue3"}}...)
		require.NoError(t, err)

		operations := []newstorage.Operation{
			{Key: "key1", Value: nil, Tags: nil},
			{Key: "key2", Value: nil, Tags: nil},
			{Key: "key3", Value: nil, Tags: nil},
		}

		err = store.Batch(operations)
		require.NoError(t, err)

		// Check and make sure values can't be found now

		value, err := store.Get("key1")
		require.True(t, errors.Is(err, newstorage.ErrDataNotFound), "got unexpected error or no error")
		require.Nil(t, value)

		value, err = store.Get("key2")
		require.True(t, errors.Is(err, newstorage.ErrDataNotFound), "got unexpected error or no error")
		require.Nil(t, value)

		value, err = store.Get("key3")
		require.True(t, errors.Is(err, newstorage.ErrDataNotFound), "got unexpected error or no error")
		require.Nil(t, value)

		tags, err := store.GetTags("key1")
		require.True(t, errors.Is(err, newstorage.ErrDataNotFound), "got unexpected error or no error")
		require.Empty(t, tags)

		tags, err = store.GetTags("key2")
		require.True(t, errors.Is(err, newstorage.ErrDataNotFound), "got unexpected error or no error")
		require.Empty(t, tags)

		tags, err = store.GetTags("key3")
		require.True(t, errors.Is(err, newstorage.ErrDataNotFound), "got unexpected error or no error")
		require.Empty(t, tags)
	})
	t.Run("Success: Put value and then delete it in the same Batch call", func(t *testing.T) {
		store, err := provider.OpenStore(randomStoreName())
		require.NoError(t, err)
		require.NotNil(t, store)

		operations := []newstorage.Operation{
			{Key: "key1", Value: []byte("value1"), Tags: []newstorage.Tag{{Name: "tagName1", Value: "tagValue1"}}},
			{Key: "key1", Value: nil, Tags: nil},
		}

		err = store.Batch(operations)
		require.NoError(t, err)

		// Check and make sure that the delete effectively "overrode" the put in the Batch call.

		value, err := store.Get("key1")
		require.True(t, errors.Is(err, newstorage.ErrDataNotFound), "got unexpected error or no error")
		require.Nil(t, value)

		tags, err := store.GetTags("key1")
		require.True(t, errors.Is(err, newstorage.ErrDataNotFound), "got unexpected error or no error")
		require.Empty(t, tags)
	})
	t.Run("Success: Put value and update it in the same Batch call", func(t *testing.T) {
		store, err := provider.OpenStore(randomStoreName())
		require.NoError(t, err)
		require.NotNil(t, store)

		operations := []newstorage.Operation{
			{Key: "key1", Value: []byte("value1"), Tags: []newstorage.Tag{{Name: "tagName1", Value: "tagValue1"}}},
			{Key: "key1", Value: []byte("value2"), Tags: []newstorage.Tag{{Name: "tagName2", Value: "tagValue2"}}},
		}

		err = store.Batch(operations)
		require.NoError(t, err)

		// Check and make sure that the second put effectively "overrode" the first put in the Batch call.

		value, err := store.Get("key1")
		require.NoError(t, err)
		require.Equal(t, "value2", string(value))

		tags, err := store.GetTags("key1")
		require.NoError(t, err)
		require.Len(t, tags, 1)
		require.Equal(t, "tagName2", tags[0].Name)
		require.Equal(t, "tagValue2", tags[0].Value)
	})
}

func tryNilOrBlankValues(t *testing.T, store newstorage.Store, data []byte, commonKey string) {
	// Try getting blank key
	_, err := store.Get("")
	require.Error(t, err)

	// Try putting with empty key
	err = store.Put("", data)
	require.Error(t, err)

	// Try putting nil value
	err = store.Put(commonKey, nil)
	require.Error(t, err)
}

func randomStoreName() string {
	return "store-" + uuid.New().String()
}

func putData(t *testing.T, store newstorage.Store, keys []string, values [][]byte, tags [][]newstorage.Tag) {
	for i := 0; i < len(keys); i++ {
		err := store.Put(keys[i], values[i], tags[i]...)
		require.NoError(t, err)
	}
}

func verifyExpectedIterator(t *testing.T, // nolint:gocyclo,funlen // Test file
	actualResultsItr newstorage.Iterator,
	expectedKeys []string, expectedValues [][]byte, expectedTags [][]newstorage.Tag) {
	if len(expectedValues) != len(expectedKeys) || len(expectedTags) != len(expectedKeys) {
		require.FailNow(t,
			"invalid test case. Expected keys, values and tags slices must be the same length")
	}

	var dataChecklist struct {
		keys     []string
		values   [][]byte
		tags     [][]newstorage.Tag
		received []bool
	}

	dataChecklist.keys = expectedKeys
	dataChecklist.values = expectedValues
	dataChecklist.tags = expectedTags
	dataChecklist.received = make([]bool, len(expectedKeys))

	moreResultsToCheck, err := actualResultsItr.Next()
	require.NoError(t, err)

	for moreResultsToCheck {
		dataReceivedCount := 0

		for _, received := range dataChecklist.received {
			if received {
				dataReceivedCount++
			}
		}

		if dataReceivedCount == len(dataChecklist.received) {
			require.FailNow(t, "query returned more results than expected")
		}

		var itrErr error
		receivedKey, itrErr := actualResultsItr.Key()
		require.NoError(t, itrErr)

		receivedValue, itrErr := actualResultsItr.Value()
		require.NoError(t, itrErr)

		receivedTags, itrErr := actualResultsItr.Tags()
		require.NoError(t, itrErr)

		for i := 0; i < len(dataChecklist.keys); i++ {
			if receivedKey == dataChecklist.keys[i] {
				if string(receivedValue) == string(dataChecklist.values[i]) {
					if equalTags(receivedTags, dataChecklist.tags[i]) {
						dataChecklist.received[i] = true

						break
					}
				}
			}
		}

		moreResultsToCheck, err = actualResultsItr.Next()
		require.NoError(t, err)
	}

	err = actualResultsItr.Release()
	require.NoError(t, err)

	for _, received := range dataChecklist.received {
		if !received {
			require.FailNow(t, "received unexpected query results")
		}
	}
}

func equalTags(tags1, tags2 []newstorage.Tag) bool { //nolint:gocyclo // Test file
	if len(tags1) != len(tags2) {
		return false
	}

	matchedTags1 := make([]bool, len(tags1))
	matchedTags2 := make([]bool, len(tags2))

	for i, tag1 := range tags1 {
		for j, tag2 := range tags2 {
			if matchedTags2[j] {
				continue // This tag has already found a match. Tags can only have one match!
			}

			if tag1.Name == tag2.Name && tag1.Value == tag2.Value {
				matchedTags1[i] = true
				matchedTags2[j] = true

				break
			}
		}

		if !matchedTags1[i] {
			return false
		}
	}

	for _, matchedTag := range matchedTags1 {
		if !matchedTag {
			return false
		}
	}

	for _, matchedTag := range matchedTags2 {
		if !matchedTag {
			return false
		}
	}

	return true
}
