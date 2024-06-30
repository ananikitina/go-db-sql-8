package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add добавляет строку с информацией о посылке
func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.Created_At))
	if err != nil {
		return 0, fmt.Errorf("failed to add parcel info: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last inserted id: %w", err)
	}

	return int(id), nil
}

// Get реализует чтение одной строки из таблицы parcel по заданному number
func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}
	err := s.db.QueryRow("SELECT number, client,status,address,created_at FROM parcel WHERE number = ?", number).Scan(
		&p.Number, &p.Client, &p.Status, &p.Address, &p.Created_At)
	if err != nil {
		return p, fmt.Errorf("failed to query parcel info with number %d: %w", number, err)
	}
	return p, nil
}

// GetByClient реализует чтение строк из таблицы parcel по заданному client
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	var res []Parcel
	rows, err := s.db.Query("SELECT number, client,status,address,created_at FROM parcel WHERE client = ?", client)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, fmt.Errorf("parcel with client %d does not exist", client)
		}
		return res, fmt.Errorf("failed to query parcel info with client %d: %w", client, err)
	}
	defer rows.Close()

	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.Created_At)
		if err != nil {
			return res, fmt.Errorf("failed to read parcel info with client %d: %w", client, err)
		}
		res = append(res, p)
	}

	if err = rows.Err(); err != nil {
		return res, fmt.Errorf("error during rows iteration: %w", err)
	}
	return res, nil
}

// SetStatus реализует обновление статуса в таблице parcel
func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)
	if err != nil {
		return fmt.Errorf("failed to update status for number %d: %w", number, err)
	}
	return nil
}

// SetAddress реализет обновление адреса в таблице parcel при статусе registered
func (s ParcelStore) SetAddress(number int, address string) error {
	var p Parcel
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&p.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parcel with number %d does not exist", number)
		}
		return fmt.Errorf("failed to query parcel with number %d: %w", number, err)
	}

	if p.Status == ParcelStatusRegistered {
		_, err := s.db.Exec("UPDATE parcel SET address = ? WHERE number = ?", address, number)
		if err != nil {
			return fmt.Errorf("failed to update address for number %d: %w", number, err)
		}
	} else {
		return fmt.Errorf("parcel with number %d is not registered", number)
	}

	return nil
}

// Delete реализует удаление строки из таблицы parcel при статусе registered
func (s ParcelStore) Delete(number int) error {
	var p Parcel
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&p.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parcel with number %d does not exist", number)
		}
		return fmt.Errorf("failed to query parcel with number %d: %w", number, err)
	}

	if p.Status == ParcelStatusRegistered {
		_, err := s.db.Exec("DELETE FROM parcel WHERE number = ?", number)
		if err != nil {
			return fmt.Errorf("failed to delete parcel info with number %d: %w", number, err)
		}
	} else {
		return fmt.Errorf("parcel with number %d is not registered", number)
	}

	return nil
}
