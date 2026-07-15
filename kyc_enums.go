package gaian

// Gender constrains the SubmitKYCRequest.Gender field.
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

// DocumentType constrains the SubmitKYCRequest.Type field — the kind of
// identity document being submitted.
type DocumentType string

const (
	DocumentTypeIDCard   DocumentType = "ID_CARD"
	DocumentTypePassport DocumentType = "PASSPORT"
)

// OccupationCode constrains the SubmitKYCRequest.Occupation field.
type OccupationCode string

const (
	OccupationCSuiteExecutive       OccupationCode = "OCC1"
	OccupationEmployee              OccupationCode = "OCC2"
	OccupationEntrepreneur          OccupationCode = "OCC3"
	OccupationForeignWorker         OccupationCode = "OCC4"
	OccupationForeignDomesticWorker OccupationCode = "OCC5"
	OccupationHomeMaker             OccupationCode = "OCC6"
	OccupationMiddleManagement      OccupationCode = "OCC7"
	OccupationPilot                 OccupationCode = "OCC8"
	OccupationProfessional          OccupationCode = "OCC9"
	OccupationSeniorManagement      OccupationCode = "OCC10"
	OccupationStudent               OccupationCode = "OCC11"
	OccupationOthers                OccupationCode = "OCC12"
)

// NationalityCode constrains the SubmitKYCRequest.Nationality field to an
// ISO 3166-1 alpha-2 country code (e.g. "VN", "US"). Use Valid to check a
// value before sending a request.
type NationalityCode string

// Valid reports whether n is a recognized ISO 3166-1 alpha-2 country code.
func (n NationalityCode) Valid() bool {
	_, ok := iso3166Alpha2Codes[n]
	return ok
}

// iso3166Alpha2Codes lists every officially assigned ISO 3166-1 alpha-2
// country code. See: https://www.iso.org/iso-3166-country-codes.html
var iso3166Alpha2Codes = map[NationalityCode]struct{}{
	"AD": {}, "AE": {}, "AF": {}, "AG": {}, "AI": {}, "AL": {}, "AM": {}, "AO": {}, "AQ": {}, "AR": {},
	"AS": {}, "AT": {}, "AU": {}, "AW": {}, "AX": {}, "AZ": {}, "BA": {}, "BB": {}, "BD": {}, "BE": {},
	"BF": {}, "BG": {}, "BH": {}, "BI": {}, "BJ": {}, "BL": {}, "BM": {}, "BN": {}, "BO": {}, "BQ": {},
	"BR": {}, "BS": {}, "BT": {}, "BV": {}, "BW": {}, "BY": {}, "BZ": {}, "CA": {}, "CC": {}, "CD": {},
	"CF": {}, "CG": {}, "CH": {}, "CI": {}, "CK": {}, "CL": {}, "CM": {}, "CN": {}, "CO": {}, "CR": {},
	"CU": {}, "CV": {}, "CW": {}, "CX": {}, "CY": {}, "CZ": {}, "DE": {}, "DJ": {}, "DK": {}, "DM": {},
	"DO": {}, "DZ": {}, "EC": {}, "EE": {}, "EG": {}, "EH": {}, "ER": {}, "ES": {}, "ET": {}, "FI": {},
	"FJ": {}, "FK": {}, "FM": {}, "FO": {}, "FR": {}, "GA": {}, "GB": {}, "GD": {}, "GE": {}, "GF": {},
	"GG": {}, "GH": {}, "GI": {}, "GL": {}, "GM": {}, "GN": {}, "GP": {}, "GQ": {}, "GR": {}, "GS": {},
	"GT": {}, "GU": {}, "GW": {}, "GY": {}, "HK": {}, "HM": {}, "HN": {}, "HR": {}, "HT": {}, "HU": {},
	"ID": {}, "IE": {}, "IL": {}, "IM": {}, "IN": {}, "IO": {}, "IQ": {}, "IR": {}, "IS": {}, "IT": {},
	"JE": {}, "JM": {}, "JO": {}, "JP": {}, "KE": {}, "KG": {}, "KH": {}, "KI": {}, "KM": {}, "KN": {},
	"KP": {}, "KR": {}, "KW": {}, "KY": {}, "KZ": {}, "LA": {}, "LB": {}, "LC": {}, "LI": {}, "LK": {},
	"LR": {}, "LS": {}, "LT": {}, "LU": {}, "LV": {}, "LY": {}, "MA": {}, "MC": {}, "MD": {}, "ME": {},
	"MF": {}, "MG": {}, "MH": {}, "MK": {}, "ML": {}, "MM": {}, "MN": {}, "MO": {}, "MP": {}, "MQ": {},
	"MR": {}, "MS": {}, "MT": {}, "MU": {}, "MV": {}, "MW": {}, "MX": {}, "MY": {}, "MZ": {}, "NA": {},
	"NC": {}, "NE": {}, "NF": {}, "NG": {}, "NI": {}, "NL": {}, "NO": {}, "NP": {}, "NR": {}, "NU": {},
	"NZ": {}, "OM": {}, "PA": {}, "PE": {}, "PF": {}, "PG": {}, "PH": {}, "PK": {}, "PL": {}, "PM": {},
	"PN": {}, "PR": {}, "PS": {}, "PT": {}, "PW": {}, "PY": {}, "QA": {}, "RE": {}, "RO": {}, "RS": {},
	"RU": {}, "RW": {}, "SA": {}, "SB": {}, "SC": {}, "SD": {}, "SE": {}, "SG": {}, "SH": {}, "SI": {},
	"SJ": {}, "SK": {}, "SL": {}, "SM": {}, "SN": {}, "SO": {}, "SR": {}, "SS": {}, "ST": {}, "SV": {},
	"SX": {}, "SY": {}, "SZ": {}, "TC": {}, "TD": {}, "TF": {}, "TG": {}, "TH": {}, "TJ": {}, "TK": {},
	"TL": {}, "TM": {}, "TN": {}, "TO": {}, "TR": {}, "TT": {}, "TV": {}, "TW": {}, "TZ": {}, "UA": {},
	"UG": {}, "UM": {}, "US": {}, "UY": {}, "UZ": {}, "VA": {}, "VC": {}, "VE": {}, "VG": {}, "VI": {},
	"VN": {}, "VU": {}, "WF": {}, "WS": {}, "XK": {}, "YE": {}, "YT": {}, "ZA": {}, "ZM": {}, "ZW": {},
}
