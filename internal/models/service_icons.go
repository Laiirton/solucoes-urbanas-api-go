package models

// ServiceIconMapping maps id_service to icon names based on category
var ServiceIconMapping = map[int]string{
	// Limpeza Urbana - brush-cleaning
	1: "brush-cleaning",
	2: "brush-cleaning",
	3: "brush-cleaning",
	4: "brush-cleaning",
	5: "brush-cleaning",
	6: "brush-cleaning",

	// Saúde - hospital
	7:  "hospital",
	8:  "hospital",
	9:  "hospital",
	10: "hospital",
	11: "hospital",
	12: "hospital",
	13: "hospital",

	// Educação - school
	14: "school",
	15: "school",
	16: "school",
	17: "school",

	// Iluminação Pública - lightbulb
	18: "lightbulb",
	19: "lightbulb",

	// Transporte Urbano - bus
	20: "bus",
	21: "bus",

	// Segurança Pública - shield
	22: "shield",
	23: "shield",
	24: "shield",
	25: "shield",
	26: "shield",
	27: "shield",
	28: "shield",

	// Esporte e Lazer - bike
	29: "bike",
	30: "bike",
	31: "bike",
	32: "bike",
	33: "bike",
	34: "bike",

	// Cultura - theater
	35: "theater",
	36: "theater",
	37: "theater",
	38: "theater",
	39: "theater",

	// Tributação - hand-coins
	40: "hand-coins",
	41: "hand-coins",
	42: "hand-coins",
	43: "hand-coins",

	// Assistência Social - hand-helping
	44: "hand-helping",
	45: "hand-helping",
	46: "hand-helping",
	47: "hand-helping",

	// Vias Urbanas - arrow-left-right
	48: "arrow-left-right",
	49: "arrow-left-right",
	50: "arrow-left-right",

	// Arborização e Meio Ambiente - tree
	51: "tree",
	52: "tree",

	// Agricultura - sprout
	53: "sprout",
	54: "sprout",

	// Vigilância Sanitária - shield
	55: "shield",
	56: "shield",

	// Animais - paw
	57: "paw",
	58: "paw",
	59: "paw",
	60: "paw",
	61: "paw",
}

// GetServiceIcon returns the icon name for a given service ID
func GetServiceIcon(serviceID int64) string {
	if icon, ok := ServiceIconMapping[int(serviceID)]; ok {
		return icon
	}
	return ""
}
