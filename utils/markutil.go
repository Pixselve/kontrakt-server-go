package utils

import "kontrakt-server/prisma/db"

type MarkData struct {
	Text  string
	Style string
}

func GetMarkData(mark db.Mark) MarkData {
	switch mark {
	default:
		return MarkData{
			Text:  "À faire",
			Style: `{"fill":{"type":"pattern","color":["#a3a3a3"],"pattern":1}}`,
		}
	case db.MarkTOFINISH:
		return MarkData{
			Text:  "À Terminer",
			Style: `{"fill":{"type":"pattern","color":["#a3a3a3"],"pattern":1}}`,
		}
	case db.MarkTOCORRECT:
		return MarkData{
			Text:  "À corriger",
			Style: `{"fill":{"type":"pattern","color":["#a3a3a3"],"pattern":1}}`,
		}
	case db.MarkGOOD:
		return MarkData{
			Text:  "Acquis avec quelques erreurs",
			Style: `{"fill":{"type":"pattern","color":["#0040ff"],"pattern":1}}`,
		}
	case db.MarkVERYGOOD:
		return MarkData{
			Text:  "Acquis",
			Style: `{"fill":{"type":"pattern","color":["#15ff00"],"pattern":1}}`,
		}
	case db.MarkBAD:
		return MarkData{
			Text:  "En voie d'acquisition",
			Style: `{"fill":{"type":"pattern","color":["#ff8c00"],"pattern":1}}`,
		}
	case db.MarkVERYBAD:
		return MarkData{
			Text:  "Non acquis",
			Style: `{"fill":{"type":"pattern","color":["#ff0000"],"pattern":1}}`,
		}
	}
}
