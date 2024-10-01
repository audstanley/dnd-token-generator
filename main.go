package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
	"github.com/spf13/cobra"
)

func cropToCircle(img image.Image, borderWidth int) image.Image {
	originalSize := img.Bounds().Size()

	// Resize the image to 81%
	resizedImg := resize.Resize(uint(int(float64(originalSize.X)*0.81)), 0, img, resize.Lanczos3)

	// Create a new RGBA image with transparent background and adjusted size for border
	circleMask := image.NewRGBA(image.Rect(-borderWidth, -borderWidth, originalSize.X+borderWidth, originalSize.Y+borderWidth))
	draw.Draw(circleMask, circleMask.Rect, image.Transparent, image.Point{0, 0}, draw.Src)

	// Calculate inner and outer circle radii
	innerRadius := originalSize.X/2 - borderWidth
	outerRadius := originalSize.X / 2

	// Get the bounds of the resized image
	resizedBounds := resizedImg.Bounds()

	// Iterate over the original image size, including the border
	for y := -borderWidth; y < originalSize.Y+borderWidth; y++ {
		for x := -borderWidth; x < originalSize.X+borderWidth; x++ {
			// Distance from center
			distance := math.Sqrt(math.Pow(float64(x+borderWidth-originalSize.X/2), 2) + math.Pow(float64(y+borderWidth-originalSize.Y/2), 2))

			// Check for ring (between inner and outer radius)
			if distance > float64(innerRadius) && distance <= float64(outerRadius) {
				circleMask.Set(x+borderWidth, y+borderWidth, color.Black) // Set black for border
			} else if distance <= float64(innerRadius) {
				// Use resized image data
				resizedX := int(float64(x+borderWidth) * float64(resizedBounds.Dx()) / float64(originalSize.X))
				resizedY := int(float64(y+borderWidth) * float64(resizedBounds.Dy()) / float64(originalSize.Y))
				circleMask.Set(x+borderWidth, y+borderWidth, resizedImg.At(resizedX, resizedY))
			}
		}
	}

	// Removed: Logic for drawing the resized image onto the circle mask

	return circleMask
}

func distance(p1, p2 image.Point) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(float64(dx*dx + dy*dy))
}

func main() {
	var folderPath string
	var pixels int

	rootCmd := &cobra.Command{
		Use:   "crop-images",
		Short: "Crops images into circles",
		Long:  "Crops images into circles with a specified border width",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Please provide the folder path")
				os.Exit(1)
			}

			folderPath = args[0]

			err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				if filepath.Ext(path) == ".png" {
					fmt.Println("Processing:", path)

					file, err := os.Open(path)
					if err != nil {
						return err
					}
					defer file.Close()

					img, err := png.Decode(file)
					if err != nil {
						return err
					}

					croppedImg := cropToCircle(img, pixels)

					outFile, err := os.Create(filepath.Join("cropped_images", filepath.Base(path)))
					if err != nil {
						return err
					}
					defer outFile.Close()

					err = png.Encode(outFile, croppedImg)
					if err != nil {
						return err
					}
				}

				return nil
			})

			if err != nil {
				fmt.Println("Error:", err)
			}
		},
	}

	rootCmd.Flags().StringVarP(&folderPath, "folder", "f", "", "Folder path")
	rootCmd.Flags().IntVarP(&pixels, "pixels", "p", 20, "Border width in pixels")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
