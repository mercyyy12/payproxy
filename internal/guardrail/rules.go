package guardrail

import (
	"fmt"
	"strconv"
	"sync"
)

type Manager struct {
	maxSingleTxn float64
	dailyBudget  float64
	spentToday   float64
	mu           sync.Mutex
}

func NewManager(maxSingleTxn, dailyBudget float64) *Manager {
	return &Manager{
		maxSingleTxn: maxSingleTxn,
		dailyBudget:  dailyBudget,
	}
}

func (m *Manager) CheckAndReserve(amountStr string) error {
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return fmt.Errorf("can't convert the amount into float")
	}

	if amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if amount > m.maxSingleTxn {
		return fmt.Errorf("transaction amount $%.2f exceeds single limit of $%.2f", amount, m.maxSingleTxn)
	}

	if m.spentToday+amount > m.dailyBudget {
		return fmt.Errorf("exceeds daily budget limit of $%.2f", m.dailyBudget)
	}

	m.spentToday += amount
	return nil
}
