package db

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/suite"
	orders "github.com/xilapa/go-tiny-projects/order-processor/internal/order/entity"
)

type OrderRepositoryTestSuite struct {
	suite.Suite
	Db *sql.DB
}

// SetupSuite implements suite.SetupAllSuite
func (s *OrderRepositoryTestSuite) SetupSuite() {
	db, err := InitialiazeDb("")
	s.NoError(err)
	s.Db = db
}

// TearDownTest implements suite.TearDownTestSuite
func (s *OrderRepositoryTestSuite) TearDownTest() {
	// clear the orders after each test
	_, err := s.Db.Exec("DELETE FROM orders")
	s.NoError(err)
}

// TearDownSuite implements suite.TearDownAllSuite
func (s *OrderRepositoryTestSuite) TearDownSuite() {
	s.Db.Close()
}

var _ suite.SetupAllSuite = (*OrderRepositoryTestSuite)(nil)
var _ suite.TearDownAllSuite = (*OrderRepositoryTestSuite)(nil)
var _ suite.TearDownTestSuite = (*OrderRepositoryTestSuite)(nil)

func (s *OrderRepositoryTestSuite) TestCanSaveOrderToDb() {
	// arrange
	order, err := orders.NewOrder("123", 12.1, 2)
	s.NoError(err)
	order.CalculateFinalPrice()
	repo := NewOrderRepository(s.Db)

	// act
	affectedRows, err := repo.Save(order)
	s.NoError(err)
	s.Equal(int64(1), affectedRows)

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
	s.NoError(err)
	s.Equal(&res, order)
}

func (s *OrderRepositoryTestSuite) TestCanCountTotalOrders() {
	// Arrange
	repo := NewOrderRepository(s.Db)

	// assert there is no orders on db before the test starts
	total, err := repo.GetTotal()
	s.NoError(err)
	s.Equal(0, total)

	expectedTotal := 5
	for i := 1; i <= expectedTotal; i++ {
		order, err := orders.NewOrder(fmt.Sprintf("%d", i), float64(i*5.0), float64(i*1.0))
		s.NoError(err)
		repo.Save(order)
	}

	// Act
	total, err = repo.GetTotal()
	s.NoError(err)
	s.Equal(expectedTotal, total)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(OrderRepositoryTestSuite))
}
