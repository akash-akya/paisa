package posting

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type Posting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Date      time.Time `json:"date"`
	Payee     string    `json:"payee"`
	Account   string    `json:"account"`
	Commodity string    `json:"commodity"`
	Quantity  float64   `json:"quantity"`
	Amount    float64   `json:"amount"`

	MarketAmount float64 `gorm:"-:all" json:"market_amount"`
}

func (p *Posting) Price() float64 {
	return p.Amount / p.Quantity
}

func (p *Posting) AddQuantity(quantity float64) {
	price := p.Price()
	p.Quantity += quantity
	p.Amount = p.Quantity * price
}

func UpsertAll(db *gorm.DB, postings []*Posting) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec("DELETE FROM postings").Error
		if err != nil {
			return err
		}
		for _, posting := range postings {
			err := tx.Create(posting).Error
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func GroupByMonth(postings []Posting) map[string][]Posting {
	grouped := make(map[string][]Posting)
	for _, p := range postings {
		key := p.Date.Format("2006-01")
		ps, ok := grouped[key]
		if ok {
			grouped[key] = append(ps, p)
		} else {
			grouped[key] = []Posting{p}
		}

	}
	return grouped
}
