package models

type Account struct {
	ID      int `gorm:"primaryKey"`
	Path    string
	EthAddr string
	Sign    bool
	Ay      string
	Balance string
	Score   int
}
