package xml

import (
	"fmt"
	"os"
	"testing"
)

type fs struct {
	Name string
	Data []byte
}

var tests []fs

func Benchmark1(b *testing.B)  { benchmarkXML(b, 0) }
func Benchmark2(b *testing.B)  { benchmarkXML(b, 1) }
func Benchmark3(b *testing.B)  { benchmarkXML(b, 2) }
func Benchmark4(b *testing.B)  { benchmarkXML(b, 3) }
func Benchmark5(b *testing.B)  { benchmarkXML(b, 4) }
func Benchmark6(b *testing.B)  { benchmarkXML(b, 5) }
func Benchmark7(b *testing.B)  { benchmarkXML(b, 6) }
func Benchmark8(b *testing.B)  { benchmarkXML(b, 7) }
func Benchmark9(b *testing.B)  { benchmarkXML(b, 8) }
func Benchmark10(b *testing.B) { benchmarkXML(b, 9) }
func Benchmark11(b *testing.B) { benchmarkXML(b, 10) }
func Benchmark12(b *testing.B) { benchmarkXML(b, 11) }

func benchmarkXML(b *testing.B, idx int) {
	if len(tests) == 0 {
		files, err := os.ReadDir("testdata")
		if err != nil {
			panic(err)
		}

		for i, file := range files {
			data, err := os.ReadFile("testdata/" + file.Name())
			if err != nil {
				panic(err)
			}

			tests = append(tests, fs{
				Name: file.Name(),
				Data: data,
			})
			fmt.Printf("%d\t%s\n", i, file.Name())
		}
		b.ResetTimer()
	}

	f := tests[idx]
	for i := 0; i < b.N; i++ {
		Parse(f.Data)
	}

}
