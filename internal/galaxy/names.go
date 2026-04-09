package galaxy

var questCriticalNames = []string{
	"Acamar",
	"Adahn",
	"Aldebaran",
	"Arouan",
	"Baratas",
	"Daled",
	"Deneb",
	"Devidia",
	"Eridani",
	"Gemulon",
	"Halley",
	"Japori",
	"Kravat",
	"Linnet",
	"Melina",
	"Nix",
	"Regulus",
	"Utopia",
	"Zalkon",
}

var additionalNames = []string{
	"Aldea", "Andevian", "Antedi",
	"Balosnee", "Brax", "Bretel", "Calondia",
	"Campor", "Capelle", "Carzon", "Castor", "Cestus",
	"Cheron", "Courteney", "Damast", "Davlos",
	"Deneva", "Draylon", "Drema",
	"Endor", "Esmee", "Exo", "Ferris", "Festen",
	"Fourmi", "Frolix", "Guinifer", "Hades",
	"Hamlet", "Helena", "Hulst", "Iodine", "Iralius",
	"Janus", "Jarada", "Jason", "Kaylon",
	"Khefka", "Kira", "Klaatu", "Klaestron", "Korma",
	"Krios", "Laertes", "Largo", "Lave",
	"Ligon", "Lowry", "Magrat", "Malcoria",
	"Mentar", "Merik", "Mintaka", "Montor", "Mordan",
	"Myrthe", "Nelvana", "Nyle", "Odet",
	"Og", "Omega", "Omphalos", "Orias", "Othello",
	"Parade", "Penthara", "Picard", "Pollux", "Quator",
	"Rakhar", "Ran", "Relva", "Rhymus",
	"Rochani", "Rubicum", "Rutia", "Sarpeidon", "Sefalla",
	"Seltrice", "Sigma", "Sol", "Somari", "Stakoron",
	"Styris", "Talani", "Tamus", "Tantalos", "Tanuga",
	"Tarchannen", "Terosa", "Thera", "Titan", "Torin",
	"Triacus", "Turkana", "Tyrus", "Umberlee",
	"Vadera", "Vagra", "Vandor", "Ventax", "Xenon",
	"Xerxes", "Yew", "Yojimbo", "Zuul",
}

func allSystemNames() []string {
	names := make([]string, 0, len(questCriticalNames)+len(additionalNames))
	names = append(names, questCriticalNames...)
	names = append(names, additionalNames...)
	return names
}
