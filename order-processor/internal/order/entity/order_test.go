package orders

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCannotCreatOrderWithEmptyId(t *testing.T) {
	order := Order{}
	assert.Error(t, order.IsValid(), "invalid id")
}

func TestCannotCreateOrderWithInvalidPrice(t *testing.T) {
	order := Order{ID: "123"}
	assert.Error(t, order.IsValid(), "invalid price")
}

func TestCannotCreateOrderWithInvalidTax(t *testing.T) {
	order := Order{ID: "123", Price: 12.12}
	assert.Error(t, order.IsValid(), "invalid tax")
}

func TestCanCreateOrderWithValidParams(t *testing.T) {
	order, err := NewOrder("123", 12.12, 2)
	assert.NoError(t, err, "should be valid")
	assert.NotNil(t, order)
}

func TestCanCalculatePrice(t *testing.T) {
	order, err := NewOrder("123", 10, 2)
	assert.NoError(t, err)
	order.CalculateFinalPrice()
	assert.Equal(t, 12.0, order.FinalPrice)
}
