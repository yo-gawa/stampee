package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io/ioutil"
	"os"
	"path"

	"github.com/golang/freetype/truetype"
	"github.com/google/martian/log"
	"github.com/jessevdk/go-flags"
	br "github.com/kujtimiihoxha/go-brace-expansion"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var defaultFontColor = color.Black

type config struct {
	position  image.Point
	fontSize  float64
	fontColor color.Gray16
	outputDir string
}

func main() {
	if err := execute(); err != nil {
		log.Errorf("%v", err)
	}
}

func execute() error {
	type options struct {
		Font     string  `short:"f" long:"font" description:"Font file" required:"true"`
		Image    string  `short:"i" long:"image" description:"Template image file" required:"true"`
		String   string  `short:"s" long:"string" description:"String" required:"true"`
		X        int     `short:"x" long:"x" description:"Print position x"`
		Y        int     `short:"y" long:"y" description:"Print position y"`
		FontSize float64 `short:"p" long:"size" description:"Font size (pt)"`
		Dir      string  `short:"d" long:"dir" description:"Output dir"`
	}
	opts := &options{}
	p := flags.NewParser(opts, flags.HelpFlag|flags.PassDoubleDash)
	_, err := p.Parse()
	if err != nil {
		return err
	}

	fnt, err := loadFont(opts.Font)
	if err != nil {
		return fmt.Errorf("loadFont error: %w", err)
	}
	templ, err := loadImage(opts.Image)
	if err != nil {
		return fmt.Errorf("loadImage error: %w", err)
	}

	conf := config{
		position:  image.Point{opts.X, opts.Y},
		fontSize:  opts.FontSize,
		fontColor: defaultFontColor,
		outputDir: opts.Dir,
	}

	for _, s := range br.Expand(opts.String) {
		fmt.Println(s)
		if err := stamp(s, fnt, templ, conf); err != nil {
			return fmt.Errorf("stamp error: %w", err)
		}
	}

	return nil
}

func stamp(str string, fnt *truetype.Font, templ image.Image, config config) error {
	// copy from template image
	dst := image.NewRGBA(templ.Bounds())
	draw.Draw(dst, templ.Bounds(), templ, image.Point{}, draw.Src)

	// write string
	faceOpt := &truetype.Options{
		Size: config.fontSize,
	}
	face := truetype.NewFace(fnt, faceOpt)
	drawer := &font.Drawer{
		Face: face,
		Dst:  dst,
		Src:  image.NewUniform(config.fontColor),
	}

	drawer.Dot = fixed.P(config.position.X, config.position.Y)
	drawer.DrawString(str)

	if err := saveImage(dst, path.Join(config.outputDir, str+".png")); err != nil {
		return err
	}

	return nil
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error font file open: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("error font file open: %w", err)
	}

	return img, nil
}

func saveImage(img *image.RGBA, filename string) error {
	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write file error: %w", err)
	}

	return nil
}

func loadFont(path string) (*truetype.Font, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error font file open: %w", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error font io read: %w", err)
	}

	font, err := truetype.Parse(bytes)
	if err != nil {
		return nil, fmt.Errorf("error font io read: %w", err)
	}

	return font, nil
}
