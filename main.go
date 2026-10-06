package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"foodstore/calculator"
)

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(in io.Reader, out io.Writer) error {
	menu := calculator.DefaultMenu()
	discounts := calculator.DefaultDiscounts()

	calc, err := calculator.New(menu, discounts...)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(in)

	fmt.Fprintf(out, "Menu: %s\n", describeMenu(menu))
	for _, discount := range discounts {
		if bundle, ok := discount.(calculator.BundleDiscount); ok {
			fmt.Fprintf(out, "Bundle: %s\n", describeBundle(bundle))
		}
	}
	fmt.Fprintln(out, `Enter "<item> [quantity]", empty line to finish.`)

	items := readItems(scanner, out, menu)
	hasMemberCard := readMemberCard(scanner, out)
	if err := scanner.Err(); err != nil {
		return err
	}

	receipt, err := calc.Receipt(calculator.Order{Items: items, HasMemberCard: hasMemberCard})
	if err != nil {
		return err
	}

	fmt.Fprintln(out)
	printReceipt(out, receipt)
	return nil
}

func printReceipt(out io.Writer, receipt calculator.Receipt) {
	printAmount(out, "Subtotal", receipt.Subtotal)
	for _, discount := range receipt.Discounts {
		printAmount(out, discount.Name, -discount.Amount)
	}
	printAmount(out, "Total", receipt.Total)
}

func printAmount(out io.Writer, label string, amount calculator.Money) {
	fmt.Fprintf(out, "%-26s %10s THB\n", label+":", amount)
}

func describeMenu(menu calculator.Menu) string {
	items := make([]calculator.Item, 0, len(menu))
	for item := range menu {
		items = append(items, item)
	}
	slices.Sort(items)

	entries := make([]string, len(items))
	for i, item := range items {
		entries[i] = fmt.Sprintf("%s %s", item, menu[item])
	}
	return strings.Join(entries, ", ")
}

func describeBundle(bundle calculator.BundleDiscount) string {
	items := make([]string, len(bundle.Items))
	for i, item := range bundle.Items {
		items[i] = string(item)
	}
	return fmt.Sprintf("%s (%d%% off every %d of the same item)", strings.Join(items, ", "), bundle.Percent, bundle.Size)
}

func readItems(scanner *bufio.Scanner, out io.Writer, menu calculator.Menu) map[calculator.Item]int {
	items := make(map[calculator.Item]int)
	for {
		line, ok := prompt(scanner, out, "> ")
		if !ok || line == "" {
			return items
		}

		item, quantity, err := parseEntry(menu, line)
		if err != nil {
			fmt.Fprintln(out, err)
			continue
		}
		items[item] += quantity
	}
}

func readMemberCard(scanner *bufio.Scanner, out io.Writer) bool {
	answer, _ := prompt(scanner, out, "Member card? (y/n): ")
	return strings.EqualFold(answer, "y") || strings.EqualFold(answer, "yes")
}

func prompt(scanner *bufio.Scanner, out io.Writer, label string) (string, bool) {
	fmt.Fprint(out, label)
	if !scanner.Scan() {
		return "", false
	}
	return strings.TrimSpace(scanner.Text()), true
}

func parseEntry(menu calculator.Menu, line string) (calculator.Item, int, error) {
	fields := strings.Fields(line)
	if len(fields) > 2 {
		return "", 0, fmt.Errorf("invalid entry %q, expected \"<item> [quantity]\"", line)
	}

	item := calculator.Item(strings.ToLower(fields[0]))
	if _, ok := menu[item]; !ok {
		return "", 0, fmt.Errorf("%w: %s", calculator.ErrUnknownItem, item)
	}

	if len(fields) == 1 {
		return item, 1, nil
	}

	quantity, err := strconv.Atoi(fields[1])
	if err != nil || quantity < 1 {
		return "", 0, fmt.Errorf("%w: %s", calculator.ErrInvalidQuantity, fields[1])
	}
	return item, quantity, nil
}
