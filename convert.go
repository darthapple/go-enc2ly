package main

import (
	"fmt"
	"go-enc2ly/lily"
	"log"
	"math/big"
	"sort"

	"go-enc2ly/encore"
)

// TODO - tuplets.
// TODO - clef changes,

type ElemSequence []*encore.MeasElem

func (e ElemSequence) Len() int {
	return len(e)
}

func (e ElemSequence) Less(i, j int) bool {
	return priority(e[i]) < priority(e[j])
}

func (e ElemSequence) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func priority(e *encore.MeasElem) int {
	prio := int(e.AbsTick()) << 10
	switch e.GetType() {
	case 8:
		fallthrough
	case 9:
		prio += 10

	// pref matter:
	case 1:
		fallthrough
	case 2:
		prio += 0

	default:
		prio += 20
	}

	return prio
}

type idKey struct {
	staff int
	voice int
}

func (i *idKey) String() string {
	return fmt.Sprintf("staff%svoice%s", Int2Letter(i.staff), Int2Letter(i.voice))
}

func Convert(data *encore.Data) {
	staves := map[idKey][]*encore.MeasElem{}
	for _, m := range data.Measures {
		for _, e := range m.Elems {
			key := idKey{
				staff: e.GetStaff(),
				voice: e.Voice(),
			}
			staves[key] = append(staves[key], e)
		}
	}
	for _, v := range staves {
		sort.Sort(ElemSequence(v))
	}

	staffVoiceMap := map[int][]idKey{}
	for k, elems := range staves {
		seq := ConvertStaff(elems)
		fmt.Printf("%v = %v\n", k.String(), seq)
		staffVoiceMap[k.staff] = append(staffVoiceMap[k.staff], k)
	}

	fmt.Printf("<<\n")
	for _, voices := range staffVoiceMap {
		fmt.Printf("  \\new Staff << \n")
		for _, voice := range voices {
			fmt.Printf("  \\new Voice \\%s\n", voice.String())
		}
		fmt.Printf(">>\n")
	}
	fmt.Printf(">>\n")
}

func ConvertKey(key byte) *lily.KeySignature {
	names := []string{
		"c", "f", "bes",
		"es", "as", "des", "ges", "ces", "g", "d", "a", "e", "b",
		"fis", "cis", }

	return &lily.KeySignature{
		Name: names[key],
		ScaleType: "major",
	}	
}

func ConvertRest(n *encore.Rest) (dur lily.Duration) {
	dur.DurationLog = int(n.FaceValue) - 1
	if n.DotControl == 25 || n.DotControl == 29 {
		dur.Dots = 1
	}
	return dur
}

func Int2Letter(a int) string {
	return string(byte(a) + 'A')
}

func ConvertNote(n *encore.Note, baseStep lily.Pitch) (pit lily.Pitch, dur lily.Duration) {
	dur.DurationLog = n.DurationLog()
	if n.DotControl == 25 || n.DotControl == 29 {
		dur.Dots = 1
	}

	baseStep.Notename += int(n.Position)
	baseStep.Normalize()
	baseStep.Alteration = int(n.SemitonePitch) - (baseStep.SemitonePitch() + 60)
	return baseStep, dur
}

// Returns the pitch for ledger line below staff.
func BasePitch(clefType byte) lily.Pitch {
	switch clefType {
	case 0:
		return lily.Pitch{
			Notename: 0,
			Octave:   0,
		}
	case 1:
		return lily.Pitch{
			Notename: 2,
			Octave:   -2,
		}
	case 2:
		return lily.Pitch{
			Notename: 1,
			Octave:   -1,
		}
	case 3:
		return lily.Pitch{
			Notename: 6,
			Octave:   -2,
		}
	}
	return lily.Pitch{}
}

func skipTicks(ticks int) *lily.Skip {
	return &lily.Skip{
		Duration: lily.Duration{
			DurationLog: 4,
			Factor:      big.NewRat(int64(ticks), 60),
		},
	}
}

func ConvertStaff(elems []*encore.MeasElem) lily.Elem {
	seq := lily.Seq{}
	lastTick := -1
	var lastNote *lily.Chord
	var articulations []string
	var nextTick int
	for i, e := range elems {
		if e.GetTick() != lastTick && lastNote != nil {
			lastNote.PostEvents = articulations
			articulations = nil
		}

		if e.GetTick() == 0 && lastTick > 0 && e.GetDurationTick() > 0 {
			seq.Elems = append(seq.Elems, &lily.BarCheck{})
		}

		if i == 0 || (e.GetTick() == 0 && elems[i-1].Measure.TimeSignature() != e.Measure.TimeSignature()) {
			seq.Elems = append(seq.Elems, &lily.TimeSignature{
				Num: int(e.Measure.TimeSigNum),
				Den: int(e.Measure.TimeSigDen),
			})
		}
		if i == 0 {
			seq.Elems = append(seq.Elems,
				ConvertKey(e.LineStaffData.Key))
		}
		
		if i > 0 && nextTick < e.AbsTick() {
			seq.Elems = append(seq.Elems, skipTicks(e.AbsTick()-nextTick))
			nextTick = e.AbsTick()
		}

		end := e.AbsTick() + e.GetDurationTick()
		switch t := e.TypeSpecific.(type) {
		case *encore.Tie:
			if lastNote == nil {
				log.Println("no last for tie ", lastTick)
			} else {
				articulations = append(articulations, "~")
			}
		case *encore.Note:
			basePitch := BasePitch(e.LineStaffData.Clef)

			p, d := ConvertNote(t, basePitch)
			if e.GetTick() == lastTick {
				if lastNote == nil {
					log.Println("no last note at ", lastTick)
					continue
				}
				lastNote.Pitch = append(lastNote.Pitch, p)
			} else {
				ch := lily.Chord{Duration: d}
				ch.Pitch = append(ch.Pitch, p)
				lastNote = &ch
				seq.Elems = append(seq.Elems, lastNote)
			}
			lastTick = e.GetTick()
			if end > nextTick {
				nextTick = end
			}
		case *encore.Rest:
			d := ConvertRest(t)
			seq.Elems = append(seq.Elems, &lily.Rest{d})
			if end > nextTick {
				nextTick = end
			}
		case *encore.KeyChange:
			seq.Elems = append(seq.Elems, ConvertKey(t.NewKey))
		}
	}
	return &seq
}
