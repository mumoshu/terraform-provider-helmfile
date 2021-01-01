package helmfile

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"hash/fnv"
)

func HashObject(obj interface{}) (string, error) {
	hash := fnv.New32a()

	hash.Reset()

	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hash, "%#v", obj)

	sum := fmt.Sprint(hash.Sum32())

	return SafeEncodeString(sum), nil
}
