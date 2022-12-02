package db

import (
	"database/sql"

	orders "github.com/xilapa/go-tiny-projects/order-processor/internal/order/entity"
)

type OrderRepository struct {
	Db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{Db: db}
}

var _ orders.OrderRepo = (*OrderRepository)(nil)

func (r *OrderRepository) Save(o *orders.Order) (int64, error) {
	stmt, err := r.Db.Prepare("INSERT INTO orders (id, price, tax, final_price) VALUES (?,?,?,?)")
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(o.ID, o.Price, o.Tax, o.FinalPrice)
	if err != nil {
		return 0, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rows, nil
}
