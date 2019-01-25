package api

import (
	"fmt"
	"os"

	"github.com/globalsign/mgo"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Handler struct {
	DB *mgo.Session
}

func LaunchServer() {
	server := echo.New()

	server.Use(middleware.Logger())
	server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	session, err := mgo.Dial("mongodb://localhost:27017")
	if err != nil {
		fmt.Printf("Error while connecting database: %s", err.Error())
		os.Exit(-1)
	}

	h := Handler{DB: session}

	server.GET("/", h.Index)
	server.GET("/countries", h.GetCountries)
	server.GET("/regions", h.GetRegionsNames)
	server.GET("/indicators/list", h.ListIndicatorsNames)
	server.GET("/incomeGroups/list", h.GetIncomeGroups)
	server.GET("/incomeGroups/:incomeGroup/countryCount", h.GetCountryCountInIncomeGroup)
	server.GET("/indicators/count", h.GetIndicatorsCount)
	server.GET("/indicators/:indicatorId/info", h.GetIndicatorInfo)
	server.GET("/indicators/:indicatorId/source", h.GetIndicatorSource)
	server.GET("/indicators/:indicatorId/worldAvg", h.GetWorldAverageForIndicator)
	server.GET("/indicators/:indicatorId/bestworst/:year", h.GetBestAndWorstCountriesForIndicatorAndYear)
	server.GET("/indicators/bestworst/:region/:year", h.GetBestAndWorstCountriesForEachIndFixedYear)
	server.GET("/indicators/:indicatorId/countryCount", h.GetCountriesCountForIndicator)
	server.GET("/indicators/:indicatorId/incomeGroup/:incomeGroup/values", h.GetCountriesValuesFromIncomeGroup)
	server.GET("/indicators/:indicatorId/:countryCode/values", h.GetCountryValues)

	server.Logger.Fatal(server.Start(":1323"))
}
