package pgatour

var bdlToPGATourID = map[int]string{
	7:  "R2026006", // Sony Open in Hawaii
	8:  "R2026002", // The American Express
	9:  "R2026004", // Farmers Insurance Open
	10: "R2026003", // WM Phoenix Open
	11: "R2026005", // AT&T Pebble Beach Pro-Am
	12: "R2026007", // The Genesis Invitational
	13: "R2026010", // Cognizant Classic in The Palm Beaches
	14: "R2026009", // Arnold Palmer Invitational presented by Mastercard
	15: "R2026483", // Puerto Rico Open
	16: "R2026011", // THE PLAYERS Championship
	17: "R2026475", // Valspar Championship
	18: "R2026020", // Texas Children's Houston Open
	19: "R2026041", // Valero Texas Open
	20: "R2026014", // Masters Tournament
	21: "R2026012", // RBC Heritage
	22: "R2026018", // Zurich Classic of New Orleans
	23: "R2026556", // Cadillac Championship
	24: "R2026480", // Truist Championship
	25: "R2026553", // ONEflight Myrtle Beach Classic
	26: "R2026033", // PGA Championship
	27: "R2026019", // THE CJ CUP Byron Nelson
	28: "R2026021", // Charles Schwab Challenge
	29: "R2026023", // the Memorial Tournament presented by Workday
	30: "R2026032", // RBC Canadian Open
	31: "R2026026", // U.S. Open
	32: "R2026034", // Travelers Championship
	33: "R2026030", // John Deere Classic
	34: "R2026541", // Genesis Scottish Open
	35: "R2026518", // ISCO Championship
	36: "R2026100", // The Open Championship
	37: "R2026522", // Corales Puntacana Championship
	38: "R2026525", // 3M Open
	39: "R2026524", // Rocket Classic
	40: "R2026013", // Wyndham Championship
	41: "R2026027", // FedEx St. Jude Championship
	42: "R2026028", // BMW Championship
}

func GetPGATourID(bdlID int) string {
	return bdlToPGATourID[bdlID]
}
