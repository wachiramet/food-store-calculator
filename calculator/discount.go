package calculator

import "fmt"

type Line struct {
	Item      Item
	Quantity  int
	UnitPrice Money
}

type Bill struct {
	Lines         []Line
	HasMemberCard bool
	Total         Money
}

type AppliedDiscount struct {
	Name   string
	Amount Money
}

type Discount interface {
	Validate() error
	Apply(bill Bill) []AppliedDiscount
}

func DefaultDiscounts() []Discount {
	return []Discount{
		BundleDiscount{Items: []Item{Orange, Pink, Green}, Size: 2, Percent: 5},
		MemberDiscount{Percent: 10},
	}
}

type BundleDiscount struct {
	Items   []Item
	Size    int
	Percent int64
}

func (d BundleDiscount) Validate() error {
	if d.Size < 1 {
		return fmt.Errorf("%w: bundle size %d", ErrInvalidDiscount, d.Size)
	}
	return validatePercent(d.Percent)
}

func (d BundleDiscount) Apply(bill Bill) []AppliedDiscount {
	var applied []AppliedDiscount
	for _, line := range bill.Lines {
		if !d.applies(line.Item) {
			continue
		}
		bundles := line.Quantity / d.Size
		bundlePrice := line.UnitPrice * Money(d.Size)
		applied = append(applied, AppliedDiscount{
			Name:   fmt.Sprintf("Bundle %s x%d (%d%%)", line.Item, bundles, d.Percent),
			Amount: bundlePrice.Percent(d.Percent) * Money(bundles),
		})
	}
	return applied
}

func (d BundleDiscount) applies(item Item) bool {
	for _, i := range d.Items {
		if i == item {
			return true
		}
	}
	return false
}

type MemberDiscount struct {
	Percent int64
}

func (d MemberDiscount) Validate() error {
	return validatePercent(d.Percent)
}

func (d MemberDiscount) Apply(bill Bill) []AppliedDiscount {
	if !bill.HasMemberCard {
		return nil
	}
	return []AppliedDiscount{{
		Name:   fmt.Sprintf("Member card (%d%%)", d.Percent),
		Amount: bill.Total.Percent(d.Percent),
	}}
}

func validatePercent(percent int64) error {
	if percent < 0 || percent > 100 {
		return fmt.Errorf("%w: percent %d", ErrInvalidDiscount, percent)
	}
	return nil
}
