package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
)

func (h *Handler) Index(c echo.Context) error {
	return c.String(http.StatusOK, "Hello world !")
}

func (h *Handler) GetCountries(c echo.Context) error {
	var result []bson.M
	h.DB.Copy().DB("Countries").C("indicators").Find(bson.M{"id": 1}).Distinct("regions.countries", &result)

	return c.JSON(http.StatusOK, result)
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

func (h *Handler) GetWorldAverageForIndicator(c echo.Context) error {
	indicatorID, err := strconv.Atoi(c.Param("indicatorId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	mapReduce := &mgo.MapReduce{
		Map: fmt.Sprintf(`function() {
					annee = 1960;
					for (annee; annee <= 2010; annee++) {
						moyenne = 0;
						nbPays = 247;
						for (var i = 0; i < this.regions.length; i++) {
						for (var j = 0; j < this.regions[i].countries.length; j++) {
							for (var k = 0; k < this.regions[i].countries[j].values.length; k++) {
							if (this.regions[i].countries[j].values[k].year == annee) {
								if (
								this.regions[i].countries[j].values[k].indicator_value != null
								) {
								moyenne += Number(
									this.regions[i].countries[j].values[k].indicator_value
								);
								} else {
								nbPays -= 1;
								}
							}
							}
						}
						}
						if (nbPays != 0) {
						moyenne /= nbPays;
						}
						emit(annee.toString(), {moyenne: moyenne, nbPays: nbPays });
					}
					}`),
		Reduce: "function () {}",
	}

	var result []bson.M
	_, err = h.DB.Copy().DB("Countries").C("indicators").Find(bson.M{"id": indicatorID}).MapReduce(mapReduce, &result)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetBestAndWorstCountriesForIndicatorAndYear(c echo.Context) error {
	indicatorID, err := strconv.Atoi(c.Param("indicatorId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	mapReduce := &mgo.MapReduce{
		Map: fmt.Sprintf(`function() {
				year = %d;
				countriesValues = [];
				for (let region of this.regions) {
					for (let country of region.countries) {
					let v1 = 0;
					for (let value of country.values) {
						if (value.year === year) v1 = value.indicator_value;
					}
					if (v1 === Infinity || v1 === -Infinity || !v1) continue;
					
					countriesValues.push({ name: country.country_name, value: v1 });
					}
				}
				countriesValues.sort((a, b) => a.value < b.value);

				emit(countriesValues[0].name, countriesValues[0].value);
				emit(
					countriesValues[countriesValues.length - 1].name,
					countriesValues[countriesValues.length - 1].value
				);
			}`, year),
		Reduce: "function() {}",
	}

	var result []struct {
		ID    string  `bson:"_id" json:"id"`
		Value float64 `json:"value"`
	}
	_, err = h.DB.Copy().DB("Countries").C("indicators").Find(bson.M{"id": indicatorID}).MapReduce(mapReduce, &result)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, result)

}

func (h *Handler) GetBestAndWorstCountriesForEachIndFixedYear(c echo.Context) error {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	mapReduce := &mgo.MapReduce{
		Map: fmt.Sprintf(`function() {
			var year = %d;
			var regionName = "%s";
			var region = this.regions.find(r => r.region_name === regionName);
			var maxCode = "";
			var maxValue = -Infinity;
			var minCode = "";
			var minValue = Infinity;
			region.countries.forEach(c => {
			  var countryV = c.values.find(v => v.year === year).indicator_value;
			  if (countryV && countryV > maxValue) {
				maxValue = countryV;
				maxCode = c.country_code;
			  }
			  if (countryV && countryV < minValue) {
				minValue = countryV;
				minCode = c.country_code;
			  }
			});
			if (!(maxCode === "" || maxValue === -Infinity))
			  emit(this.id, {
				max: { code: maxCode, value: maxValue },
				min: { code: minCode, value: minValue }
			  });
		  }`, year, c.Param("region")),
		Reduce: "function() {}",
	}

	var result []interface{}
	_, err = h.DB.Copy().DB("Countries").C("indicators").Find(bson.M{}).MapReduce(mapReduce, &result)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, result)

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

func (h *Handler) GetRegionsNames(c echo.Context) error {
	var result struct {
		ID      bson.Binary `bson:"_id" json:"_id"`
		Regions []struct {
			RegionName string `bson:"region_name" json:"region_name"`
		}
	}
	h.DB.Copy().DB("Countries").C("indicators").Find(bson.M{"id": 1}).Select(bson.M{"regions.region_name": 1}).One(&result)
	var formatted []string
	for i := 0; i < len(result.Regions); i++ {
		formatted = append(formatted, result.Regions[i].RegionName)
	}
	return c.JSON(http.StatusOK, formatted)
}

func (h *Handler) GetIndicatorSource(c echo.Context) error {
	indicatorID, err := strconv.Atoi(c.Param("indicatorId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	var result bson.M
	h.DB.Copy().DB("Countries").C("indicators").Find(bson.M{"id": indicatorID}).Select(bson.M{"indicator_name": 1, "source_note": 1, "source_organization": 1}).One(&result)

	return c.JSON(http.StatusOK, result)
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
