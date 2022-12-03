package usecases

import orders "github.com/xilapa/go-tiny-projects/order-processor/internal/order/entity"

type GetTotalResult struct {
	Total int
}

type GetTotalUseCase struct {
	OrderRepository orders.OrderRepo
}

func NewGetTotalUseCase(orderRepo orders.OrderRepo) *GetTotalUseCase {
	return &GetTotalUseCase{OrderRepository: orderRepo}
}

func (u *GetTotalUseCase) Handle() (*GetTotalResult, error) {
	total, err := u.OrderRepository.GetTotal()
	if err != nil {
		return nil, err
	}
	return &GetTotalResult{total}, nil
}
