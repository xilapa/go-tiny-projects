package orders

import (
	"testing"

	assert "github.com/xilapa/go-tiny-projects/test-assertions"
)

// TODO 2: add fuzzy tests

type orderOption func(*Order)

func getTestOrder(opts ...orderOption) *Order {
	o := &Order{
		ID:    "123",
		Price: 10.0,
		Tax:   1,
	}

	for i := range opts {
		opts[i](o)
	}

	return o
}

func withId(id string) orderOption {
	return func(o *Order) {
		o.ID = id
	}
}

func withPrice(price float64) orderOption {
	return func(o *Order) {
		o.Price = price
	}
}

func withTax(tax float64) orderOption {
	return func(o *Order) {
		o.Tax = tax
	}
}

var orderValidationData = []struct {
	testName      string
	order         *Order
	errorExpected bool
}{
	{
		testName:      "valid order data",
		order:         getTestOrder(),
		errorExpected: false,
	},
	{
		testName:      "invalid id",
		order:         getTestOrder(withId("")),
		errorExpected: true,
	},
	{
		testName:      "invalid price",
		order:         getTestOrder(withPrice(0)),
		errorExpected: true,
	},
	{
		testName:      "invalid tax",
		order:         getTestOrder(withTax(0)),
		errorExpected: true,
	},
}

func TestOrderValidation(t *testing.T) {
	var err error
	for i := range orderValidationData {
		err = orderValidationData[i].order.IsValid()
		if orderValidationData[i].errorExpected {
			assert.Error(t, err, orderValidationData[i].testName)
			continue
		}
		assert.NoError(t, err, orderValidationData[i].testName)
	}
}

func TestCanCreateOrderWithValidParams(t *testing.T) {
	order, err := NewOrder("123", 12.12, 2)
	assert.NoError(t, err, "should be valid")
	assert.NotEqual(t, nil, order)
}

func TestCanCalculatePrice(t *testing.T) {
	order, err := NewOrder("123", 10, 2)
	assert.NoError(t, err)
	order.CalculateFinalPrice()
	assert.Equal(t, 12.0, order.FinalPrice)
}
