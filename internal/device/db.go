package device

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DeviceRepository interface {
	InsertDevice(ctx context.Context, addr DeviceAddr, vendor DeviceVendor) (id int, name string, err error)
	DeleteDevice(ctx context.Context, id int) error
	FetchDevices(ctx context.Context, callback func(int, DeviceAddr, DeviceVendor, string)) error
}

type pgxpoolRepository struct {
	pool *pgxpool.Pool
}

func NewPgxRepository(pool *pgxpool.Pool) DeviceRepository {
	return &pgxpoolRepository{pool: pool}
}

func (r *pgxpoolRepository) InsertDevice(ctx context.Context, addr DeviceAddr, vendor DeviceVendor) (int, string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback(ctx)

	var id int
	err = tx.QueryRow(ctx, `
		INSERT INTO devices (vendor, ip, port, name)
		VALUES ($1, $2, $3, '')
		ON CONFLICT (ip, port) DO NOTHING RETURNING id
	`, vendor, addr.Ip, addr.Port).Scan(&id)
	if err != nil {
		return 0, "", err
	}

	name := fmt.Sprintf("%s_%d", vendor, id)
	if _, err = tx.Exec(ctx, `UPDATE devices SET name = $1 WHERE id = $2`, name, id); err != nil {
		return 0, "", err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, "", err
	}
	return id, name, nil
}

func (r *pgxpoolRepository) DeleteDevice(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM devices WHERE id = $1`, id)
	return err
}

func (r *pgxpoolRepository) FetchDevices(ctx context.Context, callback func(int, DeviceAddr, DeviceVendor, string)) error {
	rows, err := r.pool.Query(ctx, `SELECT id, ip, port, vendor, name FROM devices`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var ip, vendor, name string
		var port int
		if err := rows.Scan(&id, &ip, &port, &vendor, &name); err != nil {
			return err
		}
		callback(id, DeviceAddr{Ip: ip, Port: port}, DeviceVendor(vendor), name)
	}
	return rows.Err()
}
