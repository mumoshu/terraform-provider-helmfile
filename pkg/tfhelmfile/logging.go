package tfhelmfile

import (
	"encoding/json"
	"log"
)

func dump(s string, entries []map[string]interface{}) {
	j, _ := json.Marshal(entries)
	if j == nil {
		j = []byte{}
	}
	log.Printf("DUMP[%s]: %s", s, string(j))
}
