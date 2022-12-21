package orders

import (
	"testing"

	assert "github.com/xilapa/go-tiny-projects/test-assertions"
)

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

// maps ensures that tests are independent of each other
// https://github.com/golang/go/wiki/TableDrivenTests#using-a-map-to-store-test-cases
var orderValidationData = map[string]struct {
	order         *Order
	errorExpected bool
}{
	"valid order data": {
		order:         getTestOrder(),
		errorExpected: false,
	},
	"invalid id": {
		order:         getTestOrder(withId("")),
		errorExpected: true,
	},
	"invalid price": {
		order:         getTestOrder(withPrice(0)),
		errorExpected: true,
	},
	"invalid tax": {
		order:         getTestOrder(withTax(0)),
		errorExpected: true,
	},
}

func TestOrderValidation(t *testing.T) {
	t.Parallel()
	var err error
	for i := range orderValidationData {
		err = orderValidationData[i].order.IsValid()
		if orderValidationData[i].errorExpected {
			assert.Error(t, err, i)
			continue
		}
		assert.NoError(t, err, i)
	}
}

func TestCanCreateOrderWithValidParams(t *testing.T) {
	t.Parallel()
	order, err := NewOrder("123", 12.12, 2)
	assert.NoError(t, err, "should be valid")
	assert.NotEqual(t, nil, order)
}

func TestCanCalculatePrice(t *testing.T) {
	t.Parallel()
	order, err := NewOrder("123", 10, 2)
	assert.NoError(t, err)
	order.CalculateFinalPrice()
	assert.Equal(t, 12.0, order.FinalPrice)
}

// go test -fuzz FuzzCanCalutePrice
// https://go.dev/security/fuzz/
func FuzzCanCalutePrice(f *testing.F) {
	// f.Add("abc123", 8.0, 2.1)
	f.Fuzz(func(t *testing.T, id string, price, tax float64) {
		order, err := NewOrder(id, price, tax)
		if err != nil {
			return
		}
		order.CalculateFinalPrice()
		assert.NotEqual(t, price, order.FinalPrice)
	})
}
