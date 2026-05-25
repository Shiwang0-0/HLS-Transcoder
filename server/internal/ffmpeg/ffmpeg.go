package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Pair struct {
	Resolution string
	Bitrate    string
}

func GenerateTranscoding(inputPath string) (string, string, error) {

	videoID := filepath.Base(filepath.Dir(inputPath))

	outputRoot := filepath.Join("media/output", videoID)

	variants := []Pair{
		{"1080", "5000k"},
		{"720", "2800k"},
		{"480", "1400k"},
	}

	for _, variant := range variants {
		err := GenerateVariant(inputPath, outputRoot, variant.Resolution, variant.Bitrate)

		if err != nil {
			return "", "", err
		}
	}
	return outputRoot, videoID, nil
}

func GenerateVariant(inputPath string, outputRoot string, resolution string, bitrate string) error {

	// media/uploads/uuid/video.mp4

	outputDir := filepath.Join(outputRoot, resolution)

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return err
	}

	hlsPath := filepath.Join(outputDir, "index.m3u8")
	segmentPath := filepath.Join(outputDir, "segment%03d.ts")

	scale := fmt.Sprintf("scale=-2:%s", resolution)

	cmd := exec.Command(
		"ffmpeg",

		"-i", inputPath,

		// video codec
		"-codec:v", "libx264",

		// audio codec
		"-codec:a", "aac",

		// scaling
		"-vf", scale,

		// bitrate
		"-b:v", bitrate,

		// hls configs
		"-hls_time", "10",
		"-hls_playlist_type", "vod",

		"-hls_segment_filename", segmentPath,

		"-start_number", "0", hlsPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Generating %sp transcoding...\n", resolution)

	err = cmd.Run()
	if err != nil {
		return err
	}

	fmt.Printf("%sp transcoding completed\n", resolution)

	return nil
}
