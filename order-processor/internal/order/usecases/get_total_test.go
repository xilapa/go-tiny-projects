package usecases

import (
	"database/sql"
	"fmt"
	"testing"

	orders "github.com/xilapa/go-tiny-projects/order-processor/internal/order/entity"
	"github.com/xilapa/go-tiny-projects/order-processor/internal/order/infra/database"
	assert "github.com/xilapa/go-tiny-projects/test-assertions"
)

type getTotalUseCaseTestSuite struct {
	Db *sql.DB
}

func (s *getTotalUseCaseTestSuite) canHandleGetTotal(t *testing.T) {
	// arrange
	repo := database.NewOrderRepository(s.Db)
	useCase := NewGetTotalUseCase(repo)

	// assert no orders
	res, err := useCase.Handle()
	assert.NoError(t, err)
	assert.Equal(t, 0, res.Total)

	// assert five orders
	expectedTotal := 5
	for i := 1; i <= expectedTotal; i++ {
		order, err := orders.NewOrder(fmt.Sprintf("%d", i), float64(i*5.0), float64(i*1.0))
		assert.NoError(t, err)
		repo.Save(order)
	}

	res, err = useCase.Handle()
	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, res.Total)
}

func TestGetTotalUseCase(t *testing.T) {
	// setup
	t.Parallel()
	db, err := database.InitialiazeDb("")
	assert.NoError(t, err)
	testSuite := &getTotalUseCaseTestSuite{db}

	// teardown
	defer db.Close()

	// run tests
	t.Run("CanHandleGetTotal", testSuite.canHandleGetTotal)
}
