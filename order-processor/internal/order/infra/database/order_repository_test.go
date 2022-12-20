package database

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	orders "github.com/xilapa/go-tiny-projects/order-processor/internal/order/entity"
	assert "github.com/xilapa/go-tiny-projects/test-assertions"
)

type orderRepositoryTestSuite struct {
	Db *sql.DB
}

func (s *orderRepositoryTestSuite) canSaveOrderToDb(t *testing.T) {
	// arrange
	order, err := orders.NewOrder("123", 12.1, 2)
	assert.NoError(t, err)
	order.CalculateFinalPrice()
	repo := NewOrderRepository(s.Db)

	// act
	affectedRows, err := repo.Save(order)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), affectedRows)

	// assert
	var res orders.Order
	err = s.Db.QueryRow(
		` SELECT
					id, price, tax, final_price
				FROM
					orders
				WHERE
					id = ?`,
		order.ID).
		Scan(&res.ID, &res.Price, &res.Tax, &res.FinalPrice)
	assert.NoError(t, err)
	assert.Equal(t, &res, order)
}

func (s *orderRepositoryTestSuite) canCountTotalOrders(t *testing.T) {
	// Arrange
	repo := NewOrderRepository(s.Db)

	// assert there is no orders on db before the test starts
	total, err := repo.GetTotal()
	assert.NoError(t, err)
	assert.Equal(t, 0, total)

	expectedTotal := 5
	for i := 1; i <= expectedTotal; i++ {
		order, err := orders.NewOrder(fmt.Sprintf("%d", i), float64(i*5.0), float64(i*1.0))
		assert.NoError(t, err)
		repo.Save(order)
	}

	// Act
	total, err = repo.GetTotal()
	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
}

func TestOrderRepository(t *testing.T) {
	// setup
	t.Parallel() // run parallel with other tests, not with each other
	db, err := InitialiazeDb("")
	assert.NoError(t, err)
	suite := &orderRepositoryTestSuite{db}

	// teardown
	defer db.Close()

	// run tests
	tests := map[string]func(t *testing.T){
		"CanSaveOrderToDb":    suite.canSaveOrderToDb,
		"CanCountTotalOrders": suite.canCountTotalOrders,
	}

	for i := range tests {
		t.Run(i, tests[i])
		// do cleanup between tests
		ClearOrders(db)
	}
}
