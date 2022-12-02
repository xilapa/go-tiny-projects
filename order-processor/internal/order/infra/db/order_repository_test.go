package db

import (
	"database/sql"
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

// TearDownSuite implements suite.TearDownAllSuite
func (s *OrderRepositoryTestSuite) TearDownSuite() {
	s.Db.Close()
}

var _ suite.SetupAllSuite = (*OrderRepositoryTestSuite)(nil)
var _ suite.TearDownAllSuite = (*OrderRepositoryTestSuite)(nil)

func (s *OrderRepositoryTestSuite) TestCanSaveOrderToDb() {
	// arrange
	order, err := orders.NewOrder("123", 12.1, 2)
	s.NoError(err)
	order.CalculateFinalPrice()
	repo := OrderRepository{Db: s.Db}

	// act
	affectedRows, err := repo.Save(order)
	s.NoError(err)
	s.Equal(affectedRows, int64(1))

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

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(OrderRepositoryTestSuite))
}
