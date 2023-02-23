package db

import (
	"os"

	"github.com/ONSdigital/dp-nlp-search-scrubber/config"
	"github.com/gocarina/gocsv"
)

type Industry struct {
	Code string `csv:"SIC Code"`
	Name string `csv:"Description"`
}

func getIndustry(cfg *config.Config) ([]*Industry, error) {
	file, err := os.Open(cfg.IndustryDataFile)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	ir := []*Industry{}

	if err := gocsv.UnmarshalFile(file, &ir); err != nil {
		return nil, err
	}

	return ir, nil
}