package api

import (
	"net/http"
	"strconv"

	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
)

func (h *Handler) Index(c echo.Context) error {
	return c.String(http.StatusOK, "Hello world !")
}

func (h *Handler) GetCountryCountInIncomeGroup(c echo.Context) error {
	pipeline := []bson.M{
		bson.M{"$match": bson.M{"id": 1}},
		bson.M{"$unwind": "$regions"},
		bson.M{"$unwind": "$regions.countries"},
		bson.M{"$match": bson.M{"regions.countries.income_group": c.Param("incomeGroup")}},
		bson.M{"$count": "country_name"},
	}

	var result bson.M
	h.DB.Copy().DB("Countries").C("indicators").Pipe(pipeline).One(&result)

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetIncomeGroups(c echo.Context) error {
	var result []string
	h.DB.Copy().DB("Countries").C("indicators").Find(bson.M{"id": 1}).Distinct("regions.countries.income_group", &result)

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetIndicatorsCount(c echo.Context) error {
	count, err := h.DB.Copy().DB("Countries").C("indicators").Count()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err})
	}

	return c.JSON(http.StatusOK, echo.Map{"count": count})
}

func (h *Handler) GetCountriesCountForIndicator(c echo.Context) error {
	indicatorID, err := strconv.Atoi(c.Param("indicatorId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err})
	}

	pipeline := []bson.M{
		bson.M{"$match": bson.M{"id": indicatorID}},
		bson.M{"$unwind": "$regions"},
		bson.M{"$unwind": "$regions.countries"},
		bson.M{"$count": "country_name"},
	}

	var result bson.M
	h.DB.Copy().DB("Countries").C("indicators").Pipe(pipeline).One(&result)

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetIndicatorInfo(c echo.Context) error {
	indicatorID, err := strconv.Atoi(c.Param("indicatorId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err})
	}
	db := h.DB.Copy()
	var result bson.M
	db.DB("Countries").C("indicators").Find(bson.M{"id": indicatorID}).Select(bson.M{"indicator_name": 1, "source_note": 1, "source_organization": 1}).One(&result)

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetCountriesValuesFromIncomeGroup(c echo.Context) error {
	indicatorID, err := strconv.Atoi(c.Param("indicatorId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err})
	}
	incomeGroup := c.Param("incomeGroup")

	pipeline := []bson.M{
		bson.M{"$match": bson.M{"id": indicatorID}},
		bson.M{"$unwind": "$regions"},
		bson.M{"$unwind": "$regions.countries"},
		bson.M{"$match": bson.M{"regions.countries.income_group": incomeGroup}},
		bson.M{"$project": bson.M{"regions.countries.country_name": 1, "regions.countries.values": 1}},
	}

	var result []bson.M
	db := h.DB.Copy()
	db.DB("Countries").C("indicators").Pipe(pipeline).All(&result)
	var m []bson.M
	for i := 0; i < len(result); i++ {
		m = append(m, result[i])
	}

	return c.JSON(http.StatusOK, m)
}

func (h *Handler) ListIndicatorsNames(c echo.Context) error {
	pipeline := []bson.M{
		bson.M{"$project": bson.M{"indicator_name": 1, "id": 1}},
	}

	var result []bson.M
	db := h.DB.Copy()
	db.DB("Countries").C("indicators").Pipe(pipeline).All(&result)
	var m []bson.M
	for i := 0; i < len(result); i++ {
		m = append(m, bson.M{"id": result[i]["id"].(int), "indicator_name": result[i]["indicator_name"].(string)})
	}

	return c.JSON(http.StatusOK, m)
}

func (h *Handler) FindRegionNameByCountryCode(c string) string {
	db := h.DB.Copy()
	pipeline := []bson.M{
		bson.M{"$match": bson.M{"id": 1}},
		bson.M{"$unwind": "$regions"},
		bson.M{"$match": bson.M{"regions.countries.country_code": c}},
	}

	var result bson.M
	db.DB("Countries").C("indicators").Pipe(pipeline).One(&result)

	return (result["regions"].(bson.M))["region_name"].(string)
}

func (h *Handler) GetCountryValues(c echo.Context) error {

	indicatorID, err := strconv.Atoi(c.Param("indicatorId"))
	countryCode := c.Param("countryCode")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err})
	}

	pipeline := []bson.M{
		bson.M{"$match": bson.M{"id": indicatorID}},
		bson.M{"$unwind": "$regions"},
		bson.M{"$match": bson.M{"regions.region_name": h.FindRegionNameByCountryCode(countryCode)}},
		bson.M{"$unwind": "$regions.countries"},
		bson.M{"$match": bson.M{"regions.countries.country_code": countryCode}},
	}

	var result bson.M
	db := h.DB.Copy()
	db.DB("Countries").C("indicators").Pipe(pipeline).One(&result)

	return c.JSON(http.StatusOK, result)
}
