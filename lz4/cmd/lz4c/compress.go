package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"code.cloudfoundry.org/bytefmt"
	"github.com/schollz/progressbar/v3"

	"github.com/pierrec/cmdflag"
	"github.com/pierrec/lz4/v4"
)

// Compress compresses a set of files or from stdin to stdout.
func Compress(fs *flag.FlagSet) cmdflag.Handler {
	var blockMaxSize string
	fs.StringVar(&blockMaxSize, "size", "4M", "block max size [64K,256K,1M,4M]")
	var blockChecksum bool
	fs.BoolVar(&blockChecksum, "bc", false, "enable block checksum")
	var streamChecksum bool
	fs.BoolVar(&streamChecksum, "sc", false, "disable stream checksum")
	var level int
	fs.IntVar(&level, "l", 0, "compression level (0=fastest)")
	var concurrency int
	fs.IntVar(&concurrency, "c", -1, "concurrency (default=all CPUs")

	return func(args ...string) (int, error) {
		sz, err := bytefmt.ToBytes(blockMaxSize)
		if err != nil {
			return 0, err
		}

		zw := lz4.NewWriter(nil)
		options := []lz4.Option{
			lz4.BlockChecksumOption(blockChecksum),
			lz4.BlockSizeOption(lz4.BlockSize(sz)),
			lz4.ChecksumOption(streamChecksum),
			lz4.CompressionLevelOption(lz4.CompressionLevel(level)),
			lz4.ConcurrencyOption(concurrency),
		}
		if err := zw.Apply(options...); err != nil {
			return 0, err
		}

		// Use stdin/stdout if no file provided.
		if len(args) == 0 {
			zw.Reset(os.Stdout)
			_, err := io.Copy(zw, os.Stdin)
			if err != nil {
				return 0, err
			}
			return 0, zw.Close()
		}

		for fidx, filename := range args {
			// Input file.
			file, err := os.Open(filename)
			if err != nil {
				return fidx, err
			}
			finfo, err := file.Stat()
			if err != nil {
				return fidx, err
			}
			mode := finfo.Mode() // use the same mode for the output file

			// Accumulate compressed bytes num.
			var (
				zsize int64
				size  = finfo.Size()
			)
			if size > 0 {
				// Progress bar setup.
				numBlocks := int(size) / int(sz)
				bar := progressbar.NewOptions(numBlocks,
					// File transfers are usually slow, make sure we display the bar at 0%.
					progressbar.OptionSetRenderBlankState(true),
					// Display the filename.
					progressbar.OptionSetDescription(filename),
					progressbar.OptionClearOnFinish(),
				)
				err = zw.Apply(
					lz4.OnBlockDoneOption(func(n int) {
						_ = bar.Add(1)
						atomic.AddInt64(&zsize, int64(n))
					}),
				)
				if err != nil {
					return 0, err
				}
			}

			// Output file.
			zfilename := fmt.Sprintf("%s%s", filename, lz4Extension)
			zfile, err := os.OpenFile(zfilename, os.O_CREATE|os.O_WRONLY, mode)
			if err != nil {
				return fidx, err
			}
			zw.Reset(zfile)

			// Compress.
			_, err = io.Copy(zw, file)
			if err != nil {
				return fidx, err
			}
			for _, c := range []io.Closer{zw, zfile} {
				err := c.Close()
				if err != nil {
					return fidx, err
				}
			}

			if size > 0 {
				fmt.Printf("%s %.02f%%\n", zfilename, float64(zsize)*100/float64(size))
			}
		}

		return len(args), nil
	}
}
