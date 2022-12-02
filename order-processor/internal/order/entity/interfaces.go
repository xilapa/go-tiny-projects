package orders

type OrderRepo interface {
	Save(*Order) (int64, error)
}
