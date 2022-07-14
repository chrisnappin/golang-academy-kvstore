package kvstore_test

import (
	"store/pkg/kvstore"
	"testing"
)

const key1 = "key1"
const key2 = "key2"
const value1 = "ABC"
const value2 = "DEF"
const user1 = "user1"
const user2 = "user2"

func TestEmptyStoreRead(t *testing.T) {
	store := kvstore.NewKVStore()

	value, ok := kvstore.Read(store, key1)
	if ok {
		t.Fatalf("Should have been empty but was: %t value %s", ok, value)
	}

	kvstore.Close(store)
}

func TestSimpleReadAndWrite(t *testing.T) {
	store := kvstore.NewKVStore()

	err := kvstore.Write(store, key1, value1, user1)
	if err != nil {
		t.Fatal("Write should have been successful but got:", err)
	}

	value, ok := kvstore.Read(store, key1)
	if !ok {
		t.Fatalf("Key should have been present but was: %t (value %s)", ok, value)
	}
	if value != value1 {
		t.Fatalf("Key value should have been %s but was: %s", value1, value)
	}

	kvstore.Close(store)
}

func TestUpdateByOwner(t *testing.T) {
	store := kvstore.NewKVStore()

	err := kvstore.Write(store, key1, value1, user1)
	if err != nil {
		t.Fatal("Initial write should have been successful but got:", err)
	}

	err = kvstore.Write(store, key1, value2, user1) // same user
	if err != nil {
		t.Fatal("Second write should have been successful but got:", err)
	}

	value, ok := kvstore.Read(store, key1)
	if !ok {
		t.Fatalf("Key should have been present but was: %t (value %s)", ok, value)
	}
	if value != value2 {
		t.Fatalf("Key value should have been %s but was: %s", value2, value)
	}

	kvstore.Close(store)
}

func TestUpdateByOtherUser(t *testing.T) {
	store := kvstore.NewKVStore()

	err := kvstore.Write(store, key1, value1, user1)
	if err != nil {
		t.Fatal("Initial write should have been successful but got:", err)
	}

	err = kvstore.Write(store, key1, value2, user2) // different user
	if err == nil {
		t.Fatal("Second write should have failed")
	}

	value, ok := kvstore.Read(store, key1)
	if !ok {
		t.Fatalf("Key should have been present but was: %t (value %s)", ok, value)
	}
	if value != value1 {
		t.Fatalf("Key value should have been %s but was: %s", value1, value)
	}

	kvstore.Close(store)
}

func TestEmptyStoreDelete(t *testing.T) {
	store := kvstore.NewKVStore()

	ok, err := kvstore.Delete(store, key1, user1)
	if ok {
		t.Fatalf("Should have been not deleted but was: %t value %s", ok, err)
	}

	kvstore.Close(store)
}

func TestDeleteByOwner(t *testing.T) {
	store := kvstore.NewKVStore()

	err := kvstore.Write(store, key1, value1, user1)
	if err != nil {
		t.Fatal("Initial write should have been successful but got:", err)
	}

	deleted, err := kvstore.Delete(store, key1, user1) // as same user
	if !deleted || err != nil {
		t.Fatal("Delete should have been successful but got:", deleted, err)
	}

	value, ok := kvstore.Read(store, key1)
	if ok {
		t.Fatalf("Key should not be present but was: %t (value %s)", ok, value)
	}

	kvstore.Close(store)
}

func TestDeleteByOtherUser(t *testing.T) {
	store := kvstore.NewKVStore()

	err := kvstore.Write(store, key1, value1, user1)
	if err != nil {
		t.Fatal("Initial write should have been successful but got:", err)
	}

	deleted, err := kvstore.Delete(store, key1, user2) // different user
	if deleted {
		t.Fatal("Delete should not be successful but got:", deleted, err)
	}

	value, ok := kvstore.Read(store, key1)
	if !ok {
		t.Fatalf("Key should still be present but was: %t (value %s)", ok, value)
	}

	kvstore.Close(store)
}

func TestSimpleListKey(t *testing.T) {
	store := kvstore.NewKVStore()

	err := kvstore.Write(store, key1, value1, user1)
	if err != nil {
		t.Fatal("Write should have been successful but got:", err)
	}

	entryInfo := kvstore.List(store, key1)
	if entryInfo == nil {
		t.Fatal("Key should have been present but was not")
	}
	if entryInfo.Key != key1 {
		t.Fatalf("Key should have been %s but was: %s", key1, entryInfo.Key)
	}
	if entryInfo.Owner != user1 {
		t.Fatalf("Key owner should have been %s but was: %s", user1, entryInfo.Owner)
	}

	kvstore.Close(store)
}

func TestEmptyListKey(t *testing.T) {
	store := kvstore.NewKVStore()

	entry := kvstore.List(store, key1)
	if entry != nil {
		t.Fatal("Key should have not been present but was:", entry)
	}

	kvstore.Close(store)
}

func TestEmptyListAll(t *testing.T) {
	store := kvstore.NewKVStore()

	entries := kvstore.ListAll(store)
	if len(entries) != 0 {
		t.Fatal("ListAll should have been empty but was:", entries)
	}

	kvstore.Close(store)
}

func TestSimpleListAll(t *testing.T) {
	store := kvstore.NewKVStore()

	err := kvstore.Write(store, key1, value1, user1)
	if err != nil {
		t.Fatal("First write should have been successful but got:", err)
	}

	err = kvstore.Write(store, key2, value2, user2)
	if err != nil {
		t.Fatal("Second write should have been successful but got:", err)
	}

	entries := kvstore.ListAll(store)
	if len(entries) != 2 {
		t.Fatal("ListAll should have 2 entries but was:", entries)
	}

	kvstore.Close(store)
}
