package models

type Link struct {
	ID      int `gorm:"primaryKey"`
	LinkIdx int
	FromIdx int
	ToIdx   int
}
