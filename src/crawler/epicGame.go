package crawler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Response struct {
	Data Data `json:"data"`
}

type Data struct {
	Catalog Catalog `json:"Catalog"`
}

type Catalog struct {
	SearchStore SearchStore `json:"searchStore"`
}

type SearchStore struct {
	Elements []Element `json:"elements"`
}

type Element struct {
	Title                string     `json:"title"`
	ID                   string     `json:"id"`
	Namespace            string     `json:"namespace"`
	Description          string     `json:"description"`
	EffectiveDate        *time.Time `json:"effectiveDate"`
	OfferType            string     `json:"offerType"`
	ExpiryDate           *time.Time `json:"expiryDate"`
	ViewableDate         time.Time  `json:"viewableDate"`
	Status               string     `json:"status"`
	IsCodeRedemptionOnly bool       `json:"isCodeRedemptionOnly"`
	KeyImages            []KeyImage `json:"keyImages"`
	Seller               Seller     `json:"seller"`
	ProductSlug          string     `json:"productSlug"`
	URLSlug              string     `json:"urlSlug"`
	URL                  string     `json:"url"`
	CatalogNs            CatalogNs  `json:"catalogNs"`
	OfferMappings        []Mapping  `json:"offerMappings"`
	Price                Price      `json:"price"`
}

type KeyImage struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Seller struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CatalogNs struct {
	Mappings []Mapping `json:"mappings"`
}

type Mapping struct {
	PageSlug string `json:"pageSlug"`
	PageType string `json:"pageType"`
}

/*
type DiscountSetting struct {
	DiscountType       string  `json:"discountType"`
	DiscountPercentage float64 `json:"discountPercentage"`
}

type PromotionalOffer struct {
	StartDate       time.Time       `json:"startDate"`
	EndDate         time.Time       `json:"endDate"`
	DiscountSetting DiscountSetting `json:"discountSetting"`
}

type PromotionalOffers struct {
	PromotionalOffers []PromotionalOffer `json:"promotionalOffers"`
}

type Promotions struct {
	PromotionalOffers PromotionalOffers `json:"promotionalOffers"`
}
*/

type TotalPrice struct {
	DiscountPrice   float64 `json:"discountPrice"`
	OriginalPrice   float64 `json:"originalPrice"`
	VoucherDiscount float64 `json:"voucherDiscount"`
	Discount        float64 `json:"discount"`
	CurrencyCode    string  `json:"currencyCode"`
	CurrencyInfo    struct {
		Decimals int `json:"decimals"`
	} `json:"currencyInfo"`
	FmtPrice struct {
		OriginalPrice     string `json:"originalPrice"`
		DiscountPrice     string `json:"discountPrice"`
		IntermediatePrice string `json:"intermediatePrice"`
	} `json:"fmtPrice"`
}

type Price struct {
	TotalPrice TotalPrice `json:"totalPrice"`
}

type FreeGameType struct {
	Name      string
	StartDate string
	EndDate   string
	Path      string
	Desc      string
}

func GetEpicGameData() []FreeGameType {
	url := "https://store-site-backend-static-ipv4.ak.epicgames.com/freeGamesPromotions?locale=zh-Hant&country=HK&allowCountries=HK"

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making API request:", err)
		return nil
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading API response:", err)
		return nil
	}

	var result Response
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Error unmarshaling API response:", err)
		return nil
	}

	// Access the response dynamically
	var itemList []FreeGameType

	for _, item := range result.Data.Catalog.SearchStore.Elements {
		// layout := "2006-01-02"
		// t, err := time.Parse(layout, item.EffectiveDate)
		tempStartDate := ""
		tempExpiryDate := ""

		if item.ExpiryDate != nil {
			tempExpiryDate = item.ExpiryDate.Format("2006-01-02")
		}
		if item.EffectiveDate != nil {
			tempStartDate = item.EffectiveDate.Format("2006-01-02")
		}

		temp := FreeGameType{Name: item.Title, StartDate: tempStartDate, EndDate: tempExpiryDate, Path: "", Desc: ""}
		itemList = append(itemList, temp)
	}

	return itemList

	// resData := range result.Data.Catalog.SearchStore.Elements

	// return resData
}
