package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:     1000,
		Status:     ParcelStatusRegistered,
		Address:    "test",
		Created_At: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add (отсутствие ошибки, наличие идентификатора)
	number, err := store.Add(parcel)
	require.NoError(t, err, "failed to add parcel")
	assert.NotEmpty(t, number, "parcel number should not be empty")

	// get (отсутствие ошибки, значения полей в объекте = значения полей в parcel)
	get, err := store.Get(number)
	require.NoError(t, err, "failed to get parcel")
	assert.Equal(t, parcel.Client, get.Client)
	assert.Equal(t, parcel.Address, get.Address)
	assert.Equal(t, parcel.Status, get.Status)
	assert.Equal(t, parcel.Created_At, get.Created_At)

	// delete (отсутствие ошибки, посылку больше нельзя получить из БД)
	err = store.Delete(number)
	require.NoError(t, err, "failed to delete parcel")
	_, err = store.Get(number)
	assert.Error(t, err, "error when getting deleted parcel")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add (отсутствие ошибки, наличие идентификатора)
	number, err := store.Add(parcel)
	require.NoError(t, err, "failed to add parcel")
	assert.NotEmpty(t, number, "parcel number should not be empty")

	// set address (отсутствие ошибки)
	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)
	require.NoError(t, err, "failed to set new address")

	// check
	get, _ := store.Get(number)
	assert.Equal(t, newAddress, get.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add (отсутствие ошибки, наличие идентификатора)
	number, err := store.Add(parcel)
	require.NoError(t, err, "failed to add parcel")
	assert.NotEmpty(t, number, "parcel number should not be empty")

	// set status (отсутствие ошибки)
	newStatus := "Delivered"
	err = store.SetStatus(number, newStatus)
	require.NoError(t, err, "failed to set parcel status")

	// check
	get, _ := store.Get(number)
	assert.Equal(t, newStatus, get.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add (отсутствие ошибки, наличие идентификатора)
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err, "failed to add parcel")
		assert.NotEmpty(t, id, "parcel number should not be empty")

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client (отсутствие ошибки,сравнить количество посылок)
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "failed to get parcels by client")
	assert.Len(t, storedParcels, len(parcels), "number of retrieved parcels should match")

	// check
	for _, parcel := range storedParcels {
		//все посылки из storedParcels есть в parcelMap
		originalParcel, found := parcelMap[parcel.Number]
		require.True(t, found, "retrieved parcel not found in parcelMap")

		//значения полей полученных посылок заполнены верно
		assert.Equal(t, originalParcel, parcel)
	}
}
