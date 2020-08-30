package helmfile

import (
	"encoding/json"
	"log"
	"os"
)

func dump(s string, entries []map[string]interface{}) {
	j, _ := json.Marshal(entries)
	if j == nil {
		j = []byte{}
	}
	log.Printf("DUMP[%s]: %s", s, string(j))
}

func logf(msg string, args ...interface{}) {
	ppid := os.Getppid()
	pid := os.Getpid()
	log.Printf("[DEBUG] helmfile-provider(pid=%d,ppid=%d): "+msg, append([]interface{}{pid, ppid}, args...)...)
}
