package usecases

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	orders "github.com/xilapa/go-tiny-projects/order-processor/internal/order/entity"
	"github.com/xilapa/go-tiny-projects/order-processor/internal/order/infra/database"
	assert "github.com/xilapa/go-tiny-projects/test-assertions"
)

type calculateFinalPriceUseCaseTestSuite struct {
	Db *sql.DB
}

func (s *calculateFinalPriceUseCaseTestSuite) handleReturnNoErrorWithValidCommand(t *testing.T) {
	// arrange
	repo := database.NewOrderRepository(s.Db)
	cmmd := &OrderCommand{ID: "123", Price: 10.3, Tax: 2.2}
	useCase := NewCalculateFinalPriceUseCase(repo)
	expected := &OrderResult{
		ID:         cmmd.ID,
		Price:      cmmd.Price,
		Tax:        cmmd.Tax,
		FinalPrice: (cmmd.Price + cmmd.Tax),
	}

	// act
	res, err := useCase.Handle(cmmd)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, expected, res)

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

	assert.NoError(t, err)
	assert.EqualValues(t, expected, &orderFromDb)
}

func TestCalculatePriceUseCase(t *testing.T) {
	// setup
	t.Parallel()
	db, err := database.InitialiazeDb("")
	assert.NoError(t, err)
	suite := &calculateFinalPriceUseCaseTestSuite{db}

	// teardown
	defer db.Close()

	// run tests
	t.Run("HandleReturnNoErrorWithValidCommand", suite.handleReturnNoErrorWithValidCommand)
}
