package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ONSdigital/dp-nlp-search-scrubber/db"
	"github.com/ONSdigital/dp-nlp-search-scrubber/models"
	"github.com/ONSdigital/log.go/v2/log"
)

type HelloResponse struct {
	Message string `json:"message,omitempty"`
}

// HelloHandler returns function containing a simple hello world example of an api handler
func PrefixSearchHandler(ctx context.Context, scrubberDB *db.ScrubberDB) http.HandlerFunc {
	log.Info(ctx, "api contains /scrubber/search endpoint which return a list of possible locations and industries based on OAC and SIC")
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		start := time.Now()
		log.Namespace = "search-scrubber-handler"
		ctx := context.Background()

		scrubberParams := models.GetScrubberParams(r.URL.Query())
		querySl := strings.Split(scrubberParams.Query, " ")

		matchingAreas := getAllMatchingAreas(querySl, scrubberDB)
		matchingIndustries := getAllMatchingIndustries(querySl, scrubberDB)

		scrubberResp := models.ScrubberResp{
			Time:  fmt.Sprint(time.Since(start).Microseconds(), "µs"),
			Query: scrubberParams.Query,
			Results: models.Results{
				Areas:      matchingAreas,
				Industries: matchingIndustries,
			},
		}

		resp, err := json.Marshal(scrubberResp)
		if err != nil {
			log.Fatal(ctx, "Error marshaling JSON: %v", err)
		}

		w.Write(resp)
	}
}

func getAllMatchingAreas(querySl []string, ScrubberDB *db.ScrubberDB) []*models.AreaResp {
	var matchingAreas []*models.AreaResp
	areaRespMap := make(map[string]*models.AreaResp)
	for _, q := range querySl {
		matchingRecords := ScrubberDB.AreasPFM.GetByPrefix(strings.ToUpper(q))
		for _, rData := range matchingRecords {
			area := rData.(*db.Area)
			key := area.LAName + area.RegionName + area.RegionCode
			if _, found := areaRespMap[key]; found {
				areaRespMap[key].Codes[area.OutputAreaCode] = area.OutputAreaCode
			} else {
				areaResp := &models.AreaResp{
					Name:       area.LAName,
					Region:     area.RegionName,
					RegionCode: area.RegionCode,
					Codes: map[string]string{
						area.OutputAreaCode: area.OutputAreaCode,
					},
				}

				areaRespMap[key] = areaResp
				matchingAreas = append(matchingAreas, areaRespMap[key])
			}
		}
	}

	return matchingAreas
}

func getAllMatchingIndustries(querySl []string, ScrubberDB *db.ScrubberDB) []*models.IndustryResp {
	var matchingIndustries []*models.IndustryResp
	validation := make(map[string]string)
	for _, q := range querySl {
		matchingRecords := ScrubberDB.IndustriesPFM.GetByPrefix(strings.ToUpper(q))
		for _, rData := range matchingRecords {
			industry := rData.(*db.Industry)
			if _, valid := validation[industry.Code]; !valid {
				industryResp := &models.IndustryResp{
					Code: industry.Code,
					Name: industry.Name,
				}
				matchingIndustries = append(matchingIndustries, industryResp)
			}
			validation[industry.Code] = industry.Name
		}
	}

	return matchingIndustries
}
