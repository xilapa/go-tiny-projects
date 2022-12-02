package usecases

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/suite"
	orders "github.com/xilapa/go-tiny-projects/order-processor/internal/order/entity"
	"github.com/xilapa/go-tiny-projects/order-processor/internal/order/infra/db"
)

type CalculateFinalPriceUseCaseTestSuite struct {
	suite.Suite
	Db *sql.DB
}

// TearDownSuite implements suite.TearDownAllSuite
func (s *CalculateFinalPriceUseCaseTestSuite) TearDownSuite() {
	s.Db.Close()
}

// SetupSuite implements suite.SetupAllSuite
func (s *CalculateFinalPriceUseCaseTestSuite) SetupSuite() {
	db, err := db.InitialiazeDb("")
	s.NoError(err)
	s.Db = db
}

var _ suite.SetupAllSuite = (*CalculateFinalPriceUseCaseTestSuite)(nil)
var _ suite.TearDownAllSuite = (*CalculateFinalPriceUseCaseTestSuite)(nil)

func (s *CalculateFinalPriceUseCaseTestSuite) TestHandleReturnNoErrorWithValidCommand() {
	// arrange
	repo := db.NewOrderRepository(s.Db)
	cmmd := &OrderCommand{ID: "123", Price: 10.3, Tax: 2.2}
	useCase := CalculateFinalPriceUseCase{repo}
	expected := &OrderResult{
		ID:         cmmd.ID,
		Price:      cmmd.Price,
		Tax:        cmmd.Tax,
		FinalPrice: (cmmd.Price + cmmd.Tax),
	}

	// act
	res, err := useCase.Handle(cmmd)

	// assert
	s.NoError(err)
	s.Equal(expected, res)

	// checking the db
	var orderFromDb orders.Order
	err = s.Db.QueryRow(
		` SELECT
					id, price, tax, final_price
				FROM
					orders
				WHERE
					id = ?`,
		cmmd.ID).
		Scan(
			&orderFromDb.ID, &orderFromDb.Price,
			&orderFromDb.Tax, &orderFromDb.FinalPrice,
		)

	s.NoError(err)
	s.EqualValues(expected, &orderFromDb)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(CalculateFinalPriceUseCaseTestSuite))
}
