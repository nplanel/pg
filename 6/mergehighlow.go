package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		return
	}

	fh, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()
	fl, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	defer fl.Close()

	fhs, err := fh.Stat()
	if err != nil {
		log.Fatal("stat fh", err)
	}
	fhl, err := fl.Stat()
	if err != nil {
		log.Fatal("stat fl", err)
	}

	if fhs.Size() != fhl.Size() {
		log.Fatal("high and low file are not the same size")
	}

	datah := make([]byte, fhs.Size())
	_, err = fh.Read(datah)
	if err != nil {
		log.Fatal("read datah ", err)
	}
	datal := make([]byte, fhl.Size())
	_, err = fl.Read(datal)
	if err != nil {
		log.Fatal("read datal ", err)
	}

	fout, err := os.Create("fout")
	if err != nil {
		log.Fatal("open fout", err)
	}
	for i, dl := range datal {
		dh := datah[i]
		fout.Write([]byte{dh})
		fout.Write([]byte{dl})
	}
	fout.Close()

	csumh := 0
	for _, dh := range datah {
		csumh = (csumh + int(dh)) & 0xffff
	}
	csuml := 0
	for _, dl := range datal {
		csuml = (csuml + int(dl)) & 0xffff
	}

	fmt.Printf("0x%04x 0x%04x\n", csumh, csuml)
}
