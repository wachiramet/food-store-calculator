package calculator

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
)

var (
	ErrUnknownItem     = errors.New("unknown item")
	ErrInvalidQuantity = errors.New("invalid quantity")
	ErrInvalidDiscount = errors.New("invalid discount")
)

type Order struct {
	Items         map[Item]int
	HasMemberCard bool
}

type Receipt struct {
	Subtotal  Money
	Discounts []AppliedDiscount
	Total     Money
}

type Calculator struct {
	menu      Menu
	discounts []Discount
}

func New(menu Menu, discounts ...Discount) (*Calculator, error) {
	for _, discount := range discounts {
		if err := discount.Validate(); err != nil {
			return nil, err
		}
	}

	return &Calculator{menu: menu, discounts: discounts}, nil
}

func NewDefault() (*Calculator, error) {
	return New(DefaultMenu(), DefaultDiscounts()...)
}

func (c *Calculator) Price(order Order) (Money, error) {
	receipt, err := c.Receipt(order)
	if err != nil {
		return 0, err
	}

	return receipt.Total, nil
}

func (c *Calculator) Receipt(order Order) (Receipt, error) {
	bill, err := c.newBill(order)
	if err != nil {
		return Receipt{}, err
	}

	receipt := Receipt{Subtotal: bill.Total}
	for _, discount := range c.discounts {
		for _, applied := range discount.Apply(bill) {
			if applied.Amount == 0 {
				continue
			}

			receipt.Discounts = append(receipt.Discounts, applied)
			bill.Total -= applied.Amount
		}
	}
	receipt.Total = bill.Total

	return receipt, nil
}

func (c *Calculator) newBill(order Order) (Bill, error) {
	bill := Bill{HasMemberCard: order.HasMemberCard}

	for item, quantity := range order.Items {
		price, ok := c.menu[item]
		if !ok {
			return Bill{}, fmt.Errorf("%w: %s", ErrUnknownItem, item)
		}
		if quantity < 0 {
			return Bill{}, fmt.Errorf("%w: %s x %d", ErrInvalidQuantity, item, quantity)
		}
		if quantity == 0 {
			continue
		}

		bill.Lines = append(bill.Lines, Line{Item: item, Quantity: quantity, UnitPrice: price})
		bill.Total += price * Money(quantity)
	}

	slices.SortFunc(bill.Lines, func(a, b Line) int {
		return cmp.Compare(a.Item, b.Item)
	})

	return bill, nil
}
