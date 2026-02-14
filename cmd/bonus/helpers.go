package bonus

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

func validateBonusEntry(id, amount string) error {
	if id == "" {
		return fmt.Errorf("id (participant or submission) must not be empty")
	}

	val, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("invalid amount '%s': must be a number", amount)
	}

	if math.IsNaN(val) || math.IsInf(val, 0) {
		return fmt.Errorf("invalid amount '%s': must be a finite number", amount)
	}

	if val <= 0 {
		return fmt.Errorf("invalid amount '%s': must be greater than zero", amount)
	}

	return nil
}

// parseBonusEntries converts --bonus flag values to a csv_bonuses string.
func parseBonusEntries(flags []string) (string, error) {
	var lines []string

	for _, entry := range flags {
		parts := strings.SplitN(entry, ",", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid bonus entry '%s': expected format 'id,amount'", entry)
		}

		id := strings.TrimSpace(parts[0])
		amount := strings.TrimSpace(parts[1])

		if err := validateBonusEntry(id, amount); err != nil {
			return "", err
		}

		lines = append(lines, fmt.Sprintf("%s,%s", id, amount))
	}

	return strings.Join(lines, "\n"), nil
}

// parseBonusFile reads and validates a CSV file, returning a csv_bonuses string.
func parseBonusFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to read file: %w", err)
	}

	content := strings.TrimRight(string(data), "\n\r ")
	if content == "" {
		return "", fmt.Errorf("file is empty: %s", filePath)
	}

	rawLines := strings.Split(content, "\n")
	var entries []string

	for i, line := range rawLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ",", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("line %d: invalid format '%s': expected 'id,amount'", i+1, line)
		}

		id := strings.TrimSpace(parts[0])
		amount := strings.TrimSpace(parts[1])

		if err := validateBonusEntry(id, amount); err != nil {
			return "", fmt.Errorf("line %d: %w", i+1, err)
		}

		entries = append(entries, fmt.Sprintf("%s,%s", id, amount))
	}

	if len(entries) == 0 {
		return "", fmt.Errorf("no valid entries found in file: %s", filePath)
	}

	return strings.Join(entries, "\n"), nil
}

func confirmPayment(bonusID string, nonInteractive bool, r io.Reader, w io.Writer) (bool, error) {
	if nonInteractive {
		return true, nil
	}

	fmt.Fprintf(w, "You are about to pay bonus %s. This action cannot be undone. Proceed? [y/N]: ", bonusID)

	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes", nil
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading input: %w", err)
	}

	return false, nil
}
