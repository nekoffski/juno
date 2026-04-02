package device

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func insertDevice(ctx context.Context, tx pgx.Tx, addr DeviceAddr, vendor DeviceVendor) (int, string, error) {
	var id int
	err := tx.QueryRow(ctx, `
		INSERT INTO devices (vendor, ip, port, name)
		VALUES ($1, $2, $3, '')
		ON CONFLICT (ip, port) DO NOTHING RETURNING id	
	`, vendor, addr.Ip, addr.Port).Scan(&id)

	if err != nil {
		return 0, "", err
	}

	name := fmt.Sprintf("%s_%d", vendor, id)
	_, err = tx.Exec(ctx, `UPDATE devices SET name = $1 WHERE id = $2`, name, id)
	if err != nil {
		return 0, "", err
	}

	return id, name, nil
}

func deleteDevice(ctx context.Context, pool *pgxpool.Pool, id int) error {
	_, err := pool.Exec(ctx, `DELETE FROM devices WHERE id = $1`, id)
	return err
}

func fetchDevices(ctx context.Context, pool *pgxpool.Pool, callback func(int, DeviceAddr, DeviceVendor, string)) error {
	rows, err := pool.Query(ctx, `SELECT id, ip, port, vendor, name FROM devices`)
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
