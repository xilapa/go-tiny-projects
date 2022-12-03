package usecases

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	orders "github.com/xilapa/go-tiny-projects/order-processor/internal/order/entity"
	"github.com/xilapa/go-tiny-projects/order-processor/internal/order/infra/db"
)

type GetTotalUseCaseTestSuite struct {
	suite.Suite
	Db *sql.DB
}

// SetupSuite implements suite.SetupAllSuite
func (s *GetTotalUseCaseTestSuite) SetupSuite() {
	db, err := db.InitialiazeDb("")
	s.NoError(err)
	s.Db = db
}

// TearDownSuite implements suite.TearDownAllSuite
func (s *GetTotalUseCaseTestSuite) TearDownSuite() {
	s.Db.Close()
}

var _ suite.SetupAllSuite = (*GetTotalUseCaseTestSuite)(nil)
var _ suite.TearDownAllSuite = (*GetTotalUseCaseTestSuite)(nil)

func (s *GetTotalUseCaseTestSuite) TestHandleGetTotal() {
	// arrange
	repo := db.NewOrderRepository(s.Db)
	useCase := NewGetTotalUseCase(repo)

	// assert no orders
	res, err := useCase.Handle()
	s.NoError(err)
	s.Equal(0, res.Total)

	// assert five orders
	expectedTotal := 5
	for i := 1; i <= expectedTotal; i++ {
		order, err := orders.NewOrder(fmt.Sprintf("%d", i), float64(i*5.0), float64(i*1.0))
		s.NoError(err)
		repo.Save(order)
	}

	res, err = useCase.Handle()
	s.NoError(err)
	s.Equal(5, res.Total)
}

func TestRunTestSuite(t *testing.T) {
	suite.Run(t, new(GetTotalUseCaseTestSuite))
}
