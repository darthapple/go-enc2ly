package main

import (
	"bytes"
	"fmt"
	"log"
	"io/ioutil"
)

func analyze(d *Data) {
//	analyzeTags(content)
//	Convert(d)
	//	analyzeStaff(d)
	//	messM(d)
//mess(d)
	//	analyzeKeyCh(d)
//	analyzeAll(d)
	//	analyzeStaff(d)
	analyzeMeasStaff(d)
//		analyzeStaffdata(d)
//		analyzeStaffHeader(d)	
//		analyzeLine(d)	
}

func analyzeTags(content []byte) {
	tags := map[string]int{}
	lastHI := 0
	lastHName := ""
	for i, _ := range content {
		if isH(content[i:]) && i-lastHI > 4 {
			log.Printf("Header %q, delta %d", lastHName, i-lastHI)
			sectionContent := content[lastHI:i]
			want := []byte("Flaut")
			if idx:= bytes.Index(sectionContent, want); idx > 0 {
				log.Println("found first staff", idx)

				log.Printf("content %q", content[200:432])
			}
			lastHI = i
			lastHName = string(content[i : i+4])
			tags[lastHName]++
		}
	}

	if false {
		// find size counter in header.
		log.Println(tags)
		head := content[:341]
		for t, cnt := range tags {
			offsets := []int{}
			for i, c := range head {
				if cnt == int(c) {
					offsets = append(offsets, i)
				}
			}

			log.Printf("tag %q can be at %v", t, offsets)
		}
	}
}

func analyzeLine(d *Data) {
	for i, l  := range d.Lines {
		fmt.Printf("linesize %d %v\n", i, l.VarSize)
		fmt.Printf(" %+v, %+v\n", l.LineData, l.Staffs)
	}
}

func analyzeAll(d *Data) {
	for i, m := range d.Measures[:2] {
		fmt.Printf("meas %d\n", i)
		for _, e  := range m.Elems {
			fmt.Printf("%+v\n", e)
		}
	}
}

func analyzeStaff(d *Data) {
	for _, m := range  d.Measures {
		for _, e  := range m.Elems {
			if e.GetStaff() == 0 && e.GetTypeName() == "Note"{
				fmt.Printf("%+v\n", e)
			}
		}
	}
}

func analyzeMeasStaff(d *Data) {
	for _, e  := range d.Measures[3].Elems {
		if e.GetStaff() == 0 {
			fmt.Printf("%+v\n", e)
		}
	}
}

func analyzeKeyCh(d *Data) {
	for i, m  := range d.Measures {
		for j, e  := range m.Elems {
			if e.GetType() == 32 {
				log.Printf("meas %d elt %d staff %d", i, j,
					e.GetStaff())
			}
		}
	}
}

func analyzeStaffdata(d *Data) {
	for i, s := range d.Staff {
		fmt.Printf("%d %+v\n", i, s)
	}
}

func analyzeStaffHeader(d *Data) {
	occs := make([]map[int]int, len(d.Staff[0].VarData))
	for i := range occs {
		occs[i] = make(map[int]int)
	}
	
	for _, c := range d.Staff {
		for i := range c.VarData {
			m := occs[i]
			m[int(c.VarData[i])]++
		}
	}
	log.Printf("looking for key")
	for j, o := range occs {
		if len(o) == 1 {
			continue
		}
		log.Println("values", j, len(o))
		for _, c := range d.Staff {
			fmt.Printf("%d ", c.VarData[j])
		}
		fmt.Printf("\n")
	}
	
	for i, o := range occs {
		if len(o) == 3 {
			fmt.Printf("%d: %d diff %v\n", i, len(o), o)
		}
	}
}

func messM(d *Data) {
	raw := make([]byte, len(d.Raw))
	copy(raw, d.Raw)

	d2 := Data{}
	readData(raw, &d2)
	
	err := ioutil.WriteFile("mess.enc", raw, 0644)
	if err != nil {
		log.Fatalf("WriteFile:", err)
	}
	
}

func mess(d *Data) {
	fmt.Printf("mess\n")
	for i := 0; i < 13; i ++ {
		raw := make([]byte, len(d.Raw))
		copy(raw, d.Raw)

		for _, m := range d.Measures[5:6] {
			for _, e  := range m.Elems {
				if e.GetStaff() == 6 {
					raw[e.GetOffset() + 5 + i] += 3
				}
			}
		}

		d2 := Data{}
		readData(raw, &d2)
		fmt.Printf("messed\n")
		
		err := ioutil.WriteFile(fmt.Sprintf("mess%d.enc", i), raw, 0644)
		if err != nil {
			log.Fatalf("WriteFile:", err)
		}
	}
}

func isH(x []byte) bool {
	for i := 0; i < 4; i++ {
		if !(('0' <= x[i] && x[i] <= '9') ||
			('A' <= x[i] && x[i] <= 'Z')) {
			return false
		}
	}
	return true
}

func dumpBytes(d []byte) {
	for i, c := range d {
		fmt.Printf("%5d: %3d", i, c)
		if i % 4 == 3 && i > 0 {
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n")
}