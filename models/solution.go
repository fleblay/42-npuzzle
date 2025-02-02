package models

import (
	"gorm.io/gorm"
)

type Solution struct {
	gorm.Model
	Size int `json:"size"`
	Hash string `json:"hash"`
	Solvable bool `json:"solvable"`
	Path string `json:"path"`
	Length int `json:"length"`
	Algo string `json:"algo"`
	Workers int `json:"workers"`
	Split int `json:"split"`
	Disposition string `json:"disposition"`
	ComputeMs int64 `json:"computeMs"`
}

func (solution *Solution) GetSolutions(db *gorm.DB)(*[]Solution, error) {
	var solutions []Solution
	err := db.Model(&Solution{}).Find(&solutions).Error
	return &solutions, err
}

func (solution *Solution) GetCount(db *gorm.DB)(int64, error) {
	var count int64
	res := db.Model(&Solution{}).Count(&count)
	return count, res.Error
}

func (solution *Solution) GetCountBySize(db *gorm.DB, size int) (int64, error) {
	var count int64
	res := db.Model(&Solution{}).Where("size = ?", size).Count(&count)
	return count, res.Error
}

func (solution *Solution) GetSolutionById(db *gorm.DB, id uint) error {
	return db.Model(&Solution{}).First(solution, id).Error
}

func (solution *Solution) GetSolutionBySize(db *gorm.DB, size int) (*[]Solution, error) {
	var solutions []Solution
	err := db.Model(&Solution{}).Where("size = ?", size).Find(&solutions).Error
	return &solutions, err
}

func (solution *Solution) GetRandomSolutionBySize(db *gorm.DB, size int) error {
	return db.Model(&Solution{}).Order("RANDOM()").Where("size = ?", size).Limit(1).First(&solution).Error
}

func (solution *Solution) GetSolutionByHash(db *gorm.DB, hash string, disposition string) error {
	return db.Model(&Solution{}).Where("hash = ?", hash).Where("disposition = ?", disposition).First(solution).Error
}

func (solution *Solution) UpdateOrCreateSolution(db *gorm.DB) error {
	return db.Save(solution).Error
}

func (solution *Solution) DeleteSolution(db *gorm.DB, id uint) error {
	return db.Delete(solution, id).Error
}
