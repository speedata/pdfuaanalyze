package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"

	"rsc.io/pdf"
)

var enc *xml.Encoder

// val must be /Type /StructTreeRoot
func handleStructTreeRoot(val pdf.Value) error {
	var err error
	for _, v := range val.Keys() {
		if v == "K" {
			if err = handleStructElem(val.Key(v)); err != nil {
				return err
			}
		}
	}
	return nil
}

// val must be /Type /StructElem
func handleStructElem(val pdf.Value) error {
	xElt := xml.StartElement{}
	xElt.Name.Local = val.Key("S").Name()
	enc.EncodeToken(xElt)
	var err error
	for _, v := range val.Keys() {
		if v == "K" {
			elt := val.Key(v)
			switch elt.Kind() {
			case pdf.Dict:
				if err = handleStructElem(elt); err != nil {
					return err
				}
			case pdf.Array:
				for i := range elt.Len() {
					v := elt.Index(i)
					switch v.Kind() {
					case pdf.Dict:
						if v.Key("Type").String() == "/OBJR" {
							// ignore for now
						} else {
							if err = handleStructElem(v); err != nil {
								return err
							}
						}
					case pdf.Integer:
						xElt.Attr = append(xElt.Attr, xml.Attr{Name: xml.Name{Local: "mcid"}, Value: v.String()})
					default:
						fmt.Println("~~> unhandled!", v.Kind())
					}
				}
			}
		}
	}
	enc.EncodeToken(xElt.End())
	return nil
}

func dothings() error {
	if len(os.Args) < 2 {
		fmt.Println("main", "dateiname.pdf")
		os.Exit(0)
	}
	rd, err := pdf.Open(os.Args[1])
	if err != nil {
		return err
	}
	var sw strings.Builder
	enc = xml.NewEncoder(&sw)
	enc.Indent("", "  ")

	tr := rd.Trailer()
	root := tr.Key("Root")
	sr := root.Key("StructTreeRoot")
	if !sr.IsNull() {
		handleStructTreeRoot(sr)
	}
	enc.Flush()
	fmt.Println(sw.String())

	return nil
}

func main() {
	if err := dothings(); err != nil {
		log.Fatal(err)
	}
}
