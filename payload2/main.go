package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/deanrock/blake/payload2/gob"
)

func hashData(data interface{}) string {
	var buffer bytes.Buffer

	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(data); err != nil {
		log.Fatal(err)
	}

	x := ""
	for _, i := range buffer.Bytes() {
		x = fmt.Sprintf("%s 0x%x", x, i)
	}

	return x
}

func main() {
	data := []string{"a", "b"}
	fmt.Printf("%s", hashData(data))
}
