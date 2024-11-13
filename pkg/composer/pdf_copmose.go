package composer

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jung-kurt/gofpdf"
)

func ComposeFromFiles(files []io.ReadCloser) (*bytes.Buffer, error) {
	var opt gofpdf.ImageOptions
	pdf := gofpdf.New("P", "mm", "A4", "")

	dir, err := ioutil.TempDir("", "pdfcomp")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for i, f := range files {
		if f == nil {
			continue
		}

		tmpfn := filepath.Join(dir, fmt.Sprintf("tmpfile_%d", i))
		data, _ := ioutil.ReadAll(f)
		if err := ioutil.WriteFile(tmpfn, data, 0666); err != nil {
			log.Fatal(err)
		}
		tmpf, _ := os.Open(tmpfn)

		if tmpf == nil {
			continue
		}
		m, format, err := image.DecodeConfig(tmpf)
		if err != nil {
			continue
		}
		tmpf.Seek(0, 0)
		weigh, height := calculateRotation(m.Width, m.Height)
		if weigh > 0 {
			pdf.AddPage()
		}
		_ = pdf.RegisterImageReader("file"+strconv.Itoa(i), format, tmpf)
		pdf.ImageOptions("file"+strconv.Itoa(i), 0, 0, weigh, height, true, opt, 0, "")
		tmpf.Close()
	}

	fb := []byte{}
	rw := bytes.NewBuffer(fb)

	err = pdf.Output(rw)
	if err != nil {
		return nil, err
	}

	//o := ioutil.NopCloser(rw)
	return rw, nil
}

func calculateRotation(w, h int) (width, height float64) {
	x := float32(w)
	y := float32(h)
	if x >= y {
		return 210, 0
	}
	f := x / 100
	percent := y / f
	diff := percent - 100
	if diff > 41 {
		return 0, 280
	}

	return 210, 0
}
