package params

// Default filepaths the the files to treat as default
var (
	DefaultMasterPath = "./defaultMaster.hjson"

	DefaultWeaningPath    = "./defaultEcoWeaning.hjson"
	DefaultBackgroundPath = "./defaultEcoBackground.hjson"
	DefaultFatPath        = "./defaultEcoFatcattle.hjson"
	DefaultSlaughterPath  = "./defaultEcoSlaughtercattle.hjson"

	DefaultWeaningTerminalPath    = "./defaultEcoWeaningTerm.hjson"
	DefaultBackgroundTerminalPath = "./defaultEcoBackgroundTerm.hjson"
	DefaultFatTerminalPath        = "./defaultEcoFatcattleTerm.hjson"
	DefaultSlaughterTerminalPath  = "./defaultEcoSlaughtercattleTerm.hjson"
)

// Endpoint is a possible endpoint a file can have
type Endpoint struct {
	Internal string
	Display  string
}

// Possible endpoints
var (
	Weaning    = Endpoint{Internal: "weaning", Display: "weaning"}
	Background = Endpoint{Internal: "background", Display: "background"}
	Fat        = Endpoint{Internal: "fatcattle", Display: "fed cattle (live)"}
	Slaughter  = Endpoint{Internal: "slaughtercattle", Display: "fed cattle (carcass)"}

	EndpointMap   = map[string]Endpoint{Weaning.Internal: Weaning, Background.Internal: Background, Fat.Internal: Fat, Slaughter.Internal: Slaughter}
	EndpointSlice = []Endpoint{Weaning, Background, Fat, Slaughter}
)

// IndexType is the type of index they want to run
type IndexType string

// Possible IndexTypes
var (
	OwnReplacements = IndexType("Creates own replacements")
	Terminal        = IndexType("Terminal")

	IndexTypes = []IndexType{OwnReplacements, Terminal}
)

// HerdCompositionType holds a slice of BreedCompositions and an identifier for templating reasons
type HerdCompositionType struct {
	ID     string
	Values []HerdComposition
}

// HerdComposition holds a breed compostion
type HerdComposition struct {
	Count, BreedComp interface{}
}

// AgeRange collects values for age distribution field
type AgeRange struct {
	Age     int
	Percent float64
}

// TraitSexPriceField is a single field in TraitSexPricePerCWT element
type TraitSexPriceField struct {
	Trait                 string
	Sex                   string
	WeightLow, WeightHigh int
	Cost                  float64
}

// SexMap maps symbols to animal types
const (
	SteerCode  = "S"
	HeiferCode = "F"
	CowCode    = "C"
)

var SexMap = map[string]string{
	SteerCode:  "Steer",
	HeiferCode: "Heifer",
	CowCode:    "Cow",
}

// TraitMap maps trait symbols to descriptions
var TraitMap = []Component{
	{"USREA,D", "USREA,D", "Ultrasounded rib-eye area", false},
	{"USIMF,D", "USIMF,D", "Ultrasounded intramuscular fat", false},
	{"USFAT,D", "USFAT,D", "Ultrasounded backfat thickness", false},
	{"HCW,D", "HCW,D", "Hot carcass weight", false},
	{"REA,D", "REA,D", "Carcass rib-eye area", false},
	{"FAT,D", "FAT,D", "Carcass backfat thickness", false},
	{"MS,D", "MS,D", "Carcass marbling score", false},
	{"BW,D", "BW,D", "Birth weight", false},
	{"WW,D", "WW,D", "Weaning weight - Direct", false},
	{"WW,M", "WW,M", "Weaning weight - Maternal", false},
	{"YW,D", "YW,D", "Yearling weight", false},
	{"FI,D", "FI,D", "Daily dry matter intake", false},
	{"MW,D", "MW,D", "Mature cow weight", false},
	{"STAY,D", "STAY,D", "Probability of a cow staying in the herd to age six given that she calved as a 2-year-old", false},
	{"HP,D", "HP,D", "Probability of conceiving as a 2-year-old heifer", false},
	{"CE,D", "CD,D", "Calving ease - Direct", false},
	{"CE,M", "CD,M", "Calving ease - Maternal", false},
}

// TraitKeys returns all traits in a sorted slice
func TraitKeys() []string {
	components := make([]string, 0, len(TraitMap))
	for _, t := range TraitMap {
		components = append(components, t.Short)
	}
	return components
}

// BreedComposition holds a breed composition for the breeds page
type BreedComposition struct {
	Name       string
	Encoded    string
	BreedProps []BreedProp `json:"-"`
}

// BreedProp is a single breed-percent pair
type BreedProp struct {
	Breed, Prop string
	Breeds      []string
}

// Component contains a code, description, and boolean value for whether or not this field is selected
// Used for the index components
type Component struct {
	Display     string
	Short, Long string
	Selected    bool
}

// NameVal is a name string to numeric value pair
type NameVal struct {
	Name    string
	Val     float64
	Display bool
}
