package params

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/hjson/hjson-go"
)

// Defines structure modeling the index.hjson file

// MasterParams should mock main file for iGenDec
type MasterParams struct {
	Comment                       string             `json:"Comment"`
	Burnin                        int                `json:"burnin"`
	PlanningHorizon               int                `json:"planningHorizon"`
	Traits                        []string           `json:"Traits"`
	Components                    []string           `json:"Components"`
	Genetic                       [324]float64       `json:"genetic"`
	Residual                      [225]float64       `json:"residual"`
	BreedEffects                  []string           `json:"BreedEffects"`
	HeterosisCodes                []string           `json:"HeterosisCodes"`
	HeterosisValues               []string           `json:"HeterosisValues"`
	BreedTraitSexAod              []string           `json:"BreedTraitSexAod"`
	TraitAgeEffects               []string           `json:"TraitAgeEffects"`
	AgeDist                       []string           `json:"ageDist"`
	NFoundationBulls              int                `json:"nFoundationBulls"`
	MeritFoundationBulls          []float64          `json:"meritFoundationBulls"`
	Herds                         []string           `json:"herds"`
	CalfAum                       float64            `json:"calfAum"`
	CowAum                        float64            `json:"cowAum"`
	CowHerdBreedComposition       []interface{}      `json:"CowHerdBreedComposition"`
	BullBatteryBreedComposition   []interface{}      `json:"BullBatteryBreedComposition"`
	CurrentCalvesBreedComposition []interface{}      `json:"CurrentCalvesBreedComposition"`
	BreedCompositions             []BreedComposition `json:"BreedCompositions"` // Custom field - will be ignored by iGenDec
}

// MasterParamsFromReader parses the content in the reader into ecoparams struct
func MasterParamsFromReader(r io.Reader) (*MasterParams, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var ip = &MasterParams{}
	if err = json.Unmarshal(bytes, ip); err != nil {
		return nil, err
	}
	return ip, nil
}

// Bytes returns the marshalled index params
// If need to change the way we process the params, can easily do here
func (params *MasterParams) Bytes() ([]byte, error) {
	return json.MarshalIndent(params, "", "    ")
}

// ToMap returns the values we need from the struct in a fiber compatible map
// Has to set up some fields to work with the front end
func (params *MasterParams) ToMap(m map[string]interface{}) map[string]interface{} {
	m["PlanningHorizon"] = params.PlanningHorizon
	m["Burnin"] = params.Burnin

	// Split herd up so html template can handle
	var herds [][]string
	var totalCows float64
	for i, v := range params.Herds {
		tokens := strings.Split(v, ",")
		for idx, el := range tokens {
			tokens[idx] = strings.TrimSpace(el)
		}
		herds = append(herds, tokens)
		v, _ := strconv.Atoi(herds[i][1])
		totalCows += float64(v)
	}
	m["Herds"] = herds

	// Store the agedist in a map for easier display
	m["AgeRange"] = len(params.AgeDist) + 1
	agedist := make([]AgeRange, len(params.AgeDist))
	for i, v := range params.AgeDist {
		val, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
		agedist[len(params.AgeDist)-1-i] = AgeRange{len(params.AgeDist) + 1 - i, math.Round(val*10000) / 100}
	}
	m["AgeDist"] = agedist

	m["HerdBreedComposition"] = HerdCompositionType{"CowHerdBreedComposition", parseComp(params.CowHerdBreedComposition)}
	m["BullBreedComposition"] = HerdCompositionType{"BullBatteryBreedComposition", parseComp(params.BullBatteryBreedComposition)}
	m["CurrentCalvesBreedComposition"] = HerdCompositionType{"CurrentCalvesBreedComposition", parseComp(params.CurrentCalvesBreedComposition)}

	// Should be the first field in a comma delimited string
	breeds := make([]string, len(params.HeterosisCodes))
	for idx, code := range params.HeterosisCodes {
		breeds[idx] = strings.TrimSpace(strings.SplitN(code, ",", 2)[0])
	}
	sort.Strings(breeds)
	// Prepate breed compositions
	for i := range params.BreedCompositions {
		bpString := strings.Split(params.BreedCompositions[i].Encoded, ",")
		for j := 0; j < len(bpString); j += 2 {
			params.BreedCompositions[i].BreedProps = append(params.BreedCompositions[i].BreedProps, BreedProp{bpString[j], bpString[j+1], breeds})
		}
	}
	m["BreedCompositions"] = params.BreedCompositions

	var traits []NameVal
	for _, t := range params.Traits {
		tokens := strings.Split(t, ",")
		val, _ := strconv.ParseFloat(strings.TrimSpace(tokens[1]), 64)
		name := strings.TrimSpace(tokens[0])
		nv := NameVal{name, val, true}
		// There is a blacklist of traits they don't want shown and this will override them
		switch name {
		case "USREA", "USIMF", "USFAT", "STAY", "HP", "CD":
			nv.Display = false
		}
		traits = append(traits, nv)
	}
	m["Traits"] = traits

	return m
}

// DefaultMasterParams will return the default values for index.hjson file
func DefaultMasterParams() (*MasterParams, error) {
	file, err := os.Open(DefaultMasterPath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	ip := &MasterParams{}

	// Switch on the filetype. If its hjson need to do some marshalling and unmarshalling magic
	// and if its just JSON do a simple unmarshall
	if filepath.Ext(DefaultMasterPath) == ".hjson" {
		m := make(map[string]interface{})
		if err = hjson.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parsing hjson: %w", err)
		}
		data, err = json.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("parsing hjson: %w", err)
		}
		if err = json.Unmarshal(data, ip); err != nil {
			return nil, fmt.Errorf("parsing hjson: %w", err)
		}
	} else if filepath.Ext(DefaultMasterPath) == ".json" {
		if err = json.Unmarshal(data, ip); err != nil {
			return nil, fmt.Errorf("parsing json: %w", err)
		}
	} else {
		return nil, fmt.Errorf("expecting either .json or .hjson file, have %s", DefaultMasterPath)
	}

	// Mine the breed compositions and build some generic names
	breedcomps := make(map[string]BreedComposition)
	allcomps := append(ip.CowHerdBreedComposition, ip.BullBatteryBreedComposition...)
	for idx := 1; idx < len(allcomps); idx += 2 {
		comps := strings.Split(allcomps[idx].(string), ",")
		var name string
		if len(comps) == 2 {
			name = strings.TrimSpace(comps[0])
		} else {
			for i := 0; i < len(comps); i += 2 {
				comps[i/2] = strings.TrimSpace(comps[i]) + strings.TrimSpace(comps[i+1])
			}
			name = strings.Join(comps[:len(comps)/2], " x ")
		}

		breedcomps[name] = BreedComposition{Encoded: strings.ReplaceAll(allcomps[idx].(string), " ", ""), Name: name}
	}
	ip.BreedCompositions = make([]BreedComposition, 0, len(breedcomps))
	for _, bc := range breedcomps {
		ip.BreedCompositions = append(ip.BreedCompositions, bc)
	}
	sort.Slice(ip.BreedCompositions, func(i, j int) bool { return ip.BreedCompositions[i].Name < ip.BreedCompositions[j].Name })

	return ip, err
}

func parseComp(compsIn []interface{}) []HerdComposition {
	comp := make([]HerdComposition, len(compsIn)/2)
	for idx := 0; idx < len(compsIn); idx += 2 {
		// Can't have whitespace - will mess up matching
		comps := strings.Split(compsIn[idx+1].(string), ",")
		for idx, val := range comps {
			comps[idx] = strings.TrimSpace(val)
		}
		comp[idx/2] = HerdComposition{Count: compsIn[idx].(float64), BreedComp: strings.Join(comps, ",")}
	}
	return comp
}

// 	var ip = &MasterParams{
// 		Comment:         "no comment",
// 		Burnin:          10,
// 		PlanningHorizon: 10,
// 		Traits: []string{
// 			"USREA, 9.166",
// 			"USIMF, 3.10",
// 			"USFAT, .18",
// 			"HCW, 791.91",
// 			"REA, 12.6254",
// 			"FAT, .57",
// 			"MS, .429",
// 			"BW, 85",
// 			"WW, 545.32",
// 			"YW, 850",
// 			"FI, 10.0",
// 			"MW, 1000",
// 			"STAY, .00",
// 			"HP, .0",
// 			"CD, 0.0"},
// 		Components: []string{
// 			"USREA, D",
// 			"USIMF, D",
// 			"USFAT, D",
// 			"HCW, D",
// 			"REA, D",
// 			"FAT, D",
// 			"MS, D",
// 			"BW, D",
// 			"WW, D",
// 			"WW, M",
// 			"YW, D",
// 			"YW, M",
// 			"FI, D",
// 			"MW, D",
// 			"STAY, D",
// 			"HP, D",
// 			"CD, D",
// 			"CD, M"},

// 		Genetic: [324]float64{
// 			0.402438, -0.05362, -0.00072, 12.96646, 0.248737, -0.0051, -0.19477, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// 			-0.05362, 0.213923, -0.0009, 5.020534, 0.009616, 0.008778, 0.118634, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// 			-0.00072, -0.0009, 0.00751, 0.176298, -0.00299, 0.008045, 0.008979, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// 			12.96646, 5.020534, 0.176298, 2353.502, 5.404427, -0.17691, 4.919037, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// 			0.248737, 0.009616, -0.00299, 5.404427, 0.326089, 0.014164, -0.19403, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// 			-0.0051, 0.008778, 0.008045, -0.17691, 0.014164, 0.038003, 0.005167, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// 			-0.19477, 0.118634, 0.008979, 4.919037, -0.19403, 0.005167, 0.39584, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// 			0, 0, 0, 0, 0, 0, 0, 52.10422, 89.1737, 0, 183.8029, 0, 0.151885, 194.0954, 0, 0, 7.848975, -0.93975,
// 			0, 0, 0, 0, 0, 0, 0, 89.1737, 632.2023, 0, 500, 0, 17.49085, 771.8583, 0, 0, 5.9379, -4.3124,
// 			0, 0, 0, 0, 0, 0, 0, 0, 0, 475.7031, 0, 0, 0, 0, 0, 0, 0, 0,
// 			0, 0, 0, 0, 0, 0, 0, 183.8029, 500, 0, 2342.501, 0, 25.80119, 1863.275, 0, 0, 26.69594, -11.2532,
// 			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 475.7031, 0, 0, 0, 0, 0, 0,
// 			0, 0, 0, 0, 0, 0, 0, 0.151885, 17.49085, 0, 25.80119, 0, 4.511744, 143.6348, 0, 0, 1.167006, -0.64718,
// 			0, 0, 0, 0, 0, 0, 0, 194.0954, 771.8583, 0, 1863.275, 0, 143.6348, 17531.16, 0, 0, -0.39355, 0.207366,
// 			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.006754156667, 0.00441522713, 0, 0,
// 			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.00441522713, 0.00675, 0, 0,
// 			0, 0, 0, 0, 0, 0, 0, 7.848975, 5.9379, 0, 26.69594, 0, 1.167006, -0.39355, 0, 0, 3.546032, -0.91232,
// 			0, 0, 0, 0, 0, 0, 0, -0.93975, -4.3124, 0, -11.2532, 0, -0.64718, 0.207366, 0, 0, -0.91232, 1.449777},
// 		Residual: [225]float64{
// 			0.747384, -0.00113, -0.01095, 38.80353, 0.562754, -0.00628, 0.008275, 1.403767, 7.392704, 12.0799, -0.02127, -0.08117, 0, 0, 0.031628,
// 			-0.00113, 0.4992, 0.040246, -0.00176, 0.001051, -0.002, 0.40926, -0.0429, 0.009546, 0.006614, 0.312568, 0.001852, 0, 0, -0.01087,
// 			-0.01095, 0.040246, 0.017524, -0.00911, -0.00679, -0.00383, 0.010127, 0.006882, 0.022446, 0.008829, 0.055746, -0.0022, 0, 0, -0.00416,
// 			38.80353, -0.00176, -0.00911, 7060.507, 32.62738, 9.880296, 9.886022, 289.7524, 2004.151, 4494.983, 129.7273, 1.639172, 0, 0, -0.24511,
// 			0.562754, 0.001051, -0.00679, 32.62738, 0.450313, 0.007369, -0.00548, 1.653857, 8.945172, 2.684706, -0.1038, -0.10368, 0, 0, 0.003748,
// 			-0.00628, -0.002, -0.00383, 9.880296, 0.007369, 0.048374, 0.066052, 0.546484, 2.645192, 5.011243, 0.267454, -0.07103, 0, 0, 0.016609,
// 			0.008275, 0.40926, 0.010127, 9.886022, -0.00548, 0.066052, 0.6458, 0.096473, 0.076456, 13.48783, 0.723065, -0.01175, 0, 0, -0.00213,
// 			1.403767, -0.0429, 0.006882, 289.7524, 1.653857, 0.546484, 0.096473, 121.5765, 163.9552, 36.39806, -0.79115, 666.4849, 0, 0, -30.0132,
// 			7.392704, 0.009546, 0.022446, 2004.151, 8.945172, 2.645192, 0.076456, 163.9552, 1896.607, 439.9289, 63.10466, 3463.235, 0, 0, 24.21195,
// 			12.0799, 0.006614, 0.008829, 4494.983, 2.684706, 5.011243, 13.48783, 36.39806, 439.9289, 6667.119, 146.1314, -87.0394, 0, 0, 29.9586,
// 			-0.02127, 0.312568, 0.055746, 129.7273, -0.1038, 0.267454, 0.723065, -0.79115, 63.10466, 146.1314, 8.758191, 115.8469, 0, 0, 0.275332,
// 			-0.08117, 0.001852, -0.0022, 1.639172, -0.10368, -0.07103, -0.01175, 666.4849, 3463.235, -87.0394, 115.8469, 17531.16, 0, 0, -0.06618,
// 			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.06078741, 0, 0,
// 			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.06078, 0,
// 			0.031628, -0.01087, -0.00416, -0.24511, 0.003748, 0.016609, -0.00213, -30.0132, 24.21195, 29.9586, 0.275332, -0.06618, 0, 0, 31.9143},

// 		BreedEffects: []string{
// 			"Trait,EffectType,Angus,RedAngus,Simmental,Hereford,Gelbvieh,Limousin,Brahman,Shorthorn,Charolais",
// 			"USREA,D,0,-0.9364990157,1.707011748,-0.7584009266,1.458398861,2.402155827,-1.49536492,-0.6337285618,0",
// 			"USIMF,D,0,-0.2266041419,-0.4305481398,-0.4253816485,-0.4353039057,-0.6517777256,-0.7328419182,-0.3971064582,0",
// 			"USFAT,D,0,-0.0071250441,-0.0394066021,-0.0122126674,-0.0352813686,-0.0447333408,-0.0313310252,-0.0295139463,0",
// 			"HCW,D,0,-72,-34,-94,-46,-52,-110,-86,-29.8",
// 			"REA,D,0,-0.75883,1.383164,-0.61452,1.181717,1.946428,-1.21167,-0.5135,1.84",
// 			"FAT,D,0,-0.026,-0.254,0.026,-0.172,-0.21,-0.31709,-0.204,-0.412",
// 			"MS,D,0,-0.4,-1.22,-1.52,-1.4,-1.54,-2.71208,-1.26,-1.46",
// 			"BW,D,0,0,6.8,6.2,4,4,21.2,9.8,11.2",
// 			"WW,D,0,-30.6,8.4,-36.2,-11,-13.8,36.2,-70.8,18.2",
// 			"WW,M,0,2.8,1.2,-26.8,16.4,-11,-1.8,-8.4,-14.2",
// 			"YW,D,0,-69.6,-21.8,-110.2,-54.2,-103.8,-99.2,-127.4,-33.4",
// 			"YW,M,0,2.8,1.2,-26.8,16.4,-11,-1.8,-8.4,-14.2",
// 			"FI,D,0,-0.34,-0.045,-0.87,-1.16,-1.365,-1.455,-1.1,-0.575",
// 			"MW,D,0,-60.7,-75.2,-15.4,-121.4,-96.4,-64.7,-124.9,-16.7",
// 			"STAY,D,0,0,0,0,0,0,0,0,0",
// 			"HP,D,0,0,0,0,0,0,0,0,0",
// 			"CD,D,0,2.31,1.45,-0.69,0.86,4.14,3.68,3.81,4.08",
// 			"CD,M,0,1.72,-1.45,0.22,0.86,-0.91,-0.43,-3.33,-1.18"},
// 		HeterosisCodes: []string{
// 			"Angus, B",
// 			"RedAngus, B",
// 			"Simmental, C",
// 			"Hereford, B",
// 			"Gelbvieh, C",
// 			"Limousin, C",
// 			"Brahman, Z",
// 			"Shorthorn, B",
// 			"Charolais, C"},

// 		HeterosisValues: []string{
// 			"Trait,Dir_or_Maternal,BxB,BxC,CxC,BxZ,CxZ",
// 			"USREA,D,0,0,0,0,0",
// 			"USIMF,D,0,0,0,0,0",
// 			"USFAT,D,0,0,0,0,0",
// 			"HCW,D,22.7958,28.9467,36.244,92.6822,54.2998",
// 			"REA,D,0.372,0.4061,0.4929,1.01835,0.68665",
// 			"FAT,D,0.0433071,-0.00787402,-0.00393701,0.0787402,0.0629922",
// 			"MS,D,0.17,0.06,-0.05,0.09,0.3",
// 			"BW,D,1.0361714,1.653465,1.6093726,5.3572266,4.40924",
// 			"WW,D,14.1757066,19.069963,12.5222416,50.7503524,57.1657966",
// 			"WW,M,0,0,0,0,0",
// 			"YW,D,0,0,0,0,0",
// 			"FI,D,0,0,0,0,0",
// 			"MW,D,0,0,0,0,0",
// 			"STAY,D,0,0,0,0,0",
// 			"SC,D,0,0,0,0,0",
// 			"HP,D,0,0,0,0,0",
// 			"CD,D,0,0,0,0,0",
// 			"CD,M,0,0,0,0,0"},

// 		BreedTraitSexAod: []string{
// 			"Angus,BW,M,-7.38,-3.65,-1.6,0,-0.44",
// 			"Angus,BW,F,-7.16,-3.56,-1.59,0,-0.43",
// 			"RedAngus,BW,M,-6.41,-3.28,-1.31,0,-0.92",
// 			"RedAngus,BW,F,-6.57,-3.31,-1.3,0,-0.72",
// 			"Simmental,BW,M,-6.39,-3.81,-1.67,0,0.13",
// 			"Simmental,BW,F,-6.27,-3.57,-1.51,0,0.24",
// 			"Hereford,BW,M,-6.75,-2.79,-0.57,0,-2.7",
// 			"Hereford,BW,F,-5.33,-2.42,-1,0,-1.56",
// 			"Gelbvieh,BW,M,-6.28,-3.43,-1.28,0,-0.36",
// 			"Gelbvieh,BW,F,-6.24,-3.22,-1.24,0,-0.26",
// 			"Limousin,BW,M,-6.36,-3.39,-1.55,0,0.19",
// 			"Limousin,BW,F,-6.04,-3.1,-1.45,0,0.29",
// 			"Brahman,BW,M,0,0,0,0,0",
// 			"Brahman,BW,F,0,0,0,0,0",
// 			"Shorthorn,BW,M,-6.52,-3.27,-1.31,0,-0.17",
// 			"Shorthorn,BW,F,-6.05,-3.47,-1.72,0,-0.21",
// 			"Charolais,BW,M,-6.18,-2.14,0.33,0,-1.43",
// 			"Charolais,BW,F,-7.22,-2.09,0.82,0,-0.3",
// 			"Angus,WW,M,-66.4,-33.59,-14.48,0,-7.65",
// 			"Angus,WW,F,-55.72,-28.18,-11.3,0,-5.06",
// 			"Angus,WW,S,-61.06,-30.88,-12.89,0,-6.36",
// 			"RedAngus,WW,M,-70.13,-39.02,-15.49,0,-11.65",
// 			"RedAngus,WW,F,-57.8,-30.19,-11.12,0,-9.51",
// 			"RedAngus,WW,S,-63.96,-34.6,-13.3,0,-10.58",
// 			"Simmental,WW,M,-66,-35.11,-11.86,0,-8.97",
// 			"Simmental,WW,F,-48.25,-24.76,-7.46,0,-5.74",
// 			"Simmental,WW,S,-57.12,-29.93,-9.66,0,-7.34",
// 			"Hereford,WW,M,-57.93,-32.13,-17.44,0,-25.95",
// 			"Hereford,WW,F,-36.8,-19.52,-9.71,0,-15.97",
// 			"Hereford,WW,S,-47.36,-25.82,-13.57,0,-20.96",
// 			"Gelbvieh,WW,M,-74.53,-37.72,-12.05,0,-11.19",
// 			"Gelbvieh,WW,F,-58.47,-28.27,-8.82,0,-8.46",
// 			"Gelbvieh,WW,S,-66.5,-32.99,-10.43,0,-9.82",
// 			"Limousin,WW,M,-58.06,-32.55,-14.46,0,-3.49",
// 			"Limousin,WW,F,-46.58,-25.05,-11.01,0,-3.8",
// 			"Limousin,WW,S,-52.32,-28.8,-12.73,0,-3.65",
// 			"Brahman,WW,M,0,0,0,0,0",
// 			"Brahman,WW,F,0,0,0,0,0",
// 			"Brahman,WW,S,0,0,0,0,0",
// 			"Shorthorn,WW,M,-62.52,-36.62,-18.21,0,-20.27",
// 			"Shorthorn,WW,F,-53.61,-31.11,-14.42,0,-11.93",
// 			"Shorthorn,WW,S,-58.06,-33.86,-16.31,0,-16.1",
// 			"Charolais,WW,M,-32.69,-18.35,3.82,0,21.37",
// 			"Charolais,WW,F,-13.6,3.33,19.82,0,17.01",
// 			"Charolais,WW,S,-23.14,-7.49,11.82,0,19.19"},
// 		TraitAgeEffects: []string{
// 			"USREA,0.902366895,365",
// 			"USIMF,0.005130255,365",
// 			"USFAT,0.0013832025,365",
// 			"HCW,1.20315586,540",
// 			"REA,0.00684034,540",
// 			"FAT,0.00184427,540",
// 			"MS,0.00630821,540",
// 			"BW,0,0",
// 			"WW,2.245463415,205",
// 			"YW,1.90425,365",
// 			"FI,0,0",
// 			"MW,.2,1735",
// 			"STAY,-0.000068,2190",
// 			"HP,0,447",
// 			"CD,0,730"},

// 		AgeDist: []string{
// 			".0",
// 			".0",
// 			".0",
// 			".0",
// 			".0",
// 			".0",
// 			".0",
// 			".0",
// 			".0",
// 			".0",
// 			".0",
// 			".0",
// 			".0",
// 			".0",
// 			".0",
// 			".0741718939",
// 			".0818201827",
// 			".0896445472",
// 			".0981181908",
// 			".1081985254",
// 			".1180037416",
// 			".1299658853",
// 			".1432816111",
// 			".1577308243",
// 		},
// 		NFoundationBulls: 100,
// 		MeritFoundationBulls: []float64{
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0,
// 			0},
// 		Herds: []string{
// 			"Spring,500,180,60,0.4,0.01",
// 		},
// 		CalfAum: 0.5,
// 		CowAum:  1.0,

// 		BreedCompositions: []BreedComposition{
// 			BreedComposition{Name: "Angus", Encoded: "Angus,100"},
// 			BreedComposition{Name: "Hereford", Encoded: "Hereford,100"},
// 			BreedComposition{Name: "Angus x Hereford", Encoded: "Angus,50,Hereford,50"}},
// 	}
// 	ip.CowHerdBreedComposition = []interface{}{
// 		15.0, ip.BreedCompositions[0].Encoded,
// 		85.0, ip.BreedCompositions[2].Encoded}
// 	ip.BullBatteryBreedComposition = []interface{}{
// 		50.0, ip.BreedCompositions[0].Encoded,
// 		50.0, ip.BreedCompositions[1].Encoded}

// 	return ip
// }
