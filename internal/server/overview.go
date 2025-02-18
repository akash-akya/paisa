package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Overview struct {
	Date             time.Time `json:"date"`
	InvestmentAmount float64   `json:"investment_amount"`
	WithdrawalAmount float64   `json:"withdrawal_amount"`
	GainAmount       float64   `json:"gain_amount"`
}

func GetOverview(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%").All()

	postings = service.PopulateMarketPrice(db, postings)
	overviewTimeline := computeOverviewTimeline(db, postings)
	xirr := service.XIRR(db, postings)
	return gin.H{"overview_timeline": overviewTimeline, "xirr": xirr}
}

func computeOverviewTimeline(db *gorm.DB, postings []posting.Posting) []Overview {
	var networths []Overview

	var p posting.Posting
	var pastPostings []posting.Posting

	if len(postings) == 0 {
		return networths
	}

	end := time.Now()
	for start := postings[0].Date; start.Before(end); start = start.AddDate(0, 0, 1) {
		for len(postings) > 0 && (postings[0].Date.Before(start) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			pastPostings = append(pastPostings, p)
		}

		investment := lo.Reduce(pastPostings, func(agg float64, p posting.Posting, _ int) float64 {
			if p.Amount < 0 || service.IsInterest(db, p) {
				return agg
			} else {
				return p.Amount + agg
			}
		}, 0)

		withdrawal := lo.Reduce(pastPostings, func(agg float64, p posting.Posting, _ int) float64 {
			if p.Amount > 0 || service.IsInterest(db, p) {
				return agg
			} else {
				return -p.Amount + agg
			}
		}, 0)

		balance := lo.Reduce(pastPostings, func(agg float64, p posting.Posting, _ int) float64 {
			if service.IsInterest(db, p) {
				return p.Amount + agg
			} else {
				return service.GetMarketPrice(db, p, start) + agg
			}
		}, 0)

		gain := balance + withdrawal - investment
		networths = append(networths, Overview{Date: start, InvestmentAmount: investment, WithdrawalAmount: withdrawal, GainAmount: gain})
	}
	return networths
}
