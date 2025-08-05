package utils

import (
	"io"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

type ExifDecoderInterface interface {
	Decode(file io.Reader) (*exif.Exif, error)
}
type ExifDecoder struct{}

func (e *ExifDecoder) Decode(file io.Reader) (*exif.Exif, error) {
	return exif.Decode(file)
}

func ExtractTime(x *exif.Exif) (time.Time, error) {
	return x.DateTime()
}

func IsToday(t time.Time) bool {
	now := time.Now()
	return now.Year() == t.Year() && now.YearDay() == t.YearDay()
}
