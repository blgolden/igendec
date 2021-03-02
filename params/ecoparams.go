package params

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hjson/hjson-go"
)

// Defines structure modeling the ecoIndex.hjson file

// EcoParams should mock the economical optional input file for iGenDec
type EcoParams struct {
	SaleEndpoint        string      `json:"saleEndpoint"`
	IndexTerminal       bool        `json:"indexTerminal"`
	IndexComponents     []string    `json:"indexComponents"`
	TraitSexPricePerCwt []string    `json:"traitSexPricePerCwt"`
	DiscountRate        string      `json:"discountRate"`
	AumCost             [12]float64 `json:"aumCost"`
	BackgroundAumCost   [12]float64 `json:"backgroundAumCost"`
	BackgroundDays      int         `json:"backgroundDays"`
	DaysOnFeed          float64     `json:"daysOnFeed"`
	FeedlotFeedCost     string      `json:"feedlotFeedCost"`
	GridPremiums        []string    `json:"gridPremiums"`
	ProportionInProgram string      `json:"proportionInProgram"`
}

// Bytes returns the marshalled index params
// If need to change the way we process the params, can easily do here
func (params *EcoParams) Bytes() ([]byte, error) {
	return json.MarshalIndent(params, "", "    ")
}

// ToMap returns the values we need from the struct in a fiber compatible map
func (params *EcoParams) ToMap(m map[string]interface{}) map[string]interface{} {

	// Build slice for trait-sex prices
	traitSexPrices := make([]TraitSexPriceField, len(params.TraitSexPricePerCwt))
	for i, s := range params.TraitSexPricePerCwt {
		tokens := strings.Split(s, ",")
		traitSexPrices[i] = TraitSexPriceField{
			Trait: tokens[0],
			Type:  SexMap[tokens[1]],
			Meta:  strings.Join(tokens[1:len(tokens)-1], ","),
		}

		// Ignore errors as would set to zero if they failed anyway
		traitSexPrices[i].WeightLow, _ = strconv.Atoi(tokens[2])
		traitSexPrices[i].WeightHigh, _ = strconv.Atoi(tokens[3])
		traitSexPrices[i].Cost, _ = strconv.ParseFloat(tokens[4], 64)
	}

	m["TraitSexPrice"] = traitSexPrices

	m["SaleEndpoint"] = params.SaleEndpoint
	m["DiscountRate"] = params.DiscountRate
	m["AumCost"] = params.AumCost
	m["BackgroundAumCost"] = params.BackgroundAumCost
	m["BackgroundDays"] = params.BackgroundDays
	m["DaysOnFeed"] = params.DaysOnFeed
	m["FeedlotFeedCost"] = params.FeedlotFeedCost
	m["Months"] = []string{"January", "Febuary", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

	// Add components
	components := make([]Component, len(TraitMap))
	copy(components, TraitMap)
	for idx, t := range components {
		if params.indexcomponentsHas(t.Short) {
			t.Selected = true
			components[idx] = t
		}
	}
	m["Components"] = components
	m["IndexTerminal"] = params.IndexTerminal

	if params.SaleEndpoint == Slaughter.Internal { // Add the grid-premiums
		premiums := make([][]string, len(params.GridPremiums))
		for idx, row := range params.GridPremiums {
			row = strings.ReplaceAll(row, " ", "")
			premiums[idx] = strings.Split(row, ",")
		}
		m["GridPremiums"] = premiums
		m["ProportionInProgram"] = params.ProportionInProgram
	}

	return m
}

// simple helper function for checking if an item is in a list
// not efficient but fine for small slices like we are using
func (params *EcoParams) indexcomponentsHas(item string) bool {
	for _, el := range params.IndexComponents {
		if strings.Join(strings.Fields(el), "") == item {
			return true
		}
	}
	return false
}

// DefaultEcoParams returns the default eco params as seen in ecoIndex.hjson
// Every user will be initilised with this struct
func DefaultEcoParams(endpoint Endpoint, indextype IndexType) (*EcoParams, error) {
	var filename string
	switch endpoint {
	case Weaning:
		if indextype == Terminal {
			filename = DefaultWeaningTerminalPath
		} else {
			filename = DefaultWeaningPath
		}
	case Background:
		if indextype == Terminal {
			filename = DefaultBackgroundTerminalPath
		} else {
			filename = DefaultBackgroundPath
		}
	case Fat:
		if indextype == Terminal {
			filename = DefaultFatTerminalPath
		} else {
			filename = DefaultFatPath
		}
	case Slaughter:
		if indextype == Terminal {
			filename = DefaultSlaughterTerminalPath
		} else {
			filename = DefaultSlaughterPath
		}
	default:
		return nil, fmt.Errorf("endpoint %s is not supported", endpoint)
	}
	return EcoParamsFromFile(filename)
}

// EcoParamsFromFile reads in an eco parameter file
func EcoParamsFromFile(filename string) (*EcoParams, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(file)

	ep := &EcoParams{}

	if filepath.Ext(filename) == ".hjson" {
		m := make(map[string]interface{})
		if err = hjson.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		data, err = json.Marshal(m)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(data, ep); err != nil {
			return nil, err
		}
	} else if filepath.Ext(filename) == ".json" {
		if err = json.Unmarshal(data, ep); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("expecting either json or hjson file, have %s", filename)
	}

	for idx, v := range ep.IndexComponents {
		ep.IndexComponents[idx] = strings.ReplaceAll(v, " ", "")
	}

	return ep, nil
}
