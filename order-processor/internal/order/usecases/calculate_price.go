package usecases

import (
	orders "github.com/xilapa/go-tiny-projects/order-processor/internal/order/entity"
)

type OrderCommand struct {
	ID    string
	Price float64
	Tax   float64
}

type OrderResult struct {
	ID         string
	Price      float64
	Tax        float64
	FinalPrice float64
}

type CalculateFinalPriceUseCase struct {
	OrderRepository orders.OrderRepo
}

func NewCalculateFinalPriceUseCase(repo orders.OrderRepo) *CalculateFinalPriceUseCase {
	return &CalculateFinalPriceUseCase{
		OrderRepository: repo,
	}
}

func (u *CalculateFinalPriceUseCase) Handle(cmmd *OrderCommand) (*OrderResult, error) {
	order, err := orders.NewOrder(cmmd.ID, cmmd.Price, cmmd.Tax)
	if err != nil {
		return nil, err
	}

	order.CalculateFinalPrice()

	_, err = u.OrderRepository.Save(order)
	if err != nil {
		return nil, err
	}

	return &OrderResult{
		ID:         order.ID,
		Price:      order.Price,
		Tax:        order.Tax,
		FinalPrice: order.FinalPrice,
	}, nil
}
