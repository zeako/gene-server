package genefinder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/zeako/gene-server/pkg/bufferpool"
	"golang.org/x/sync/errgroup"
)

// GenePrefix is expected for each gene string.
const GenePrefix = "AAAAAAAAAAA"

// LegalChars are gene's sequence legal characters.
const LegalChars = "AGCT"

// DefaultBufferSize is the default buffer size used by GeneFinder.
// Defaults to 16MB.
const DefaultBufferSize = 1024 * 1024 * 16

// ValidationError used for any gene input related errors
type ValidationError struct {
	s string
}

func (e *ValidationError) Error() string {
	return e.s
}

// GeneFinder enables finding gene sequences in DNA files,
// it exposes Find function for searching arbitrary length sequences.
//
// GeneFinder searches for the sequence concurrently using buffer pool to reduce heap allocation overhead,
// its buffer pool is shared by concurrent calls.
type GeneFinder struct {
	// DNA file object
	f *os.File

	// File size
	fsize int

	// Buffers pool
	bufpool *bufferpool.BufferPool
}

// New returns a newly initialized GeneFinder object.
func New(f *os.File) (*GeneFinder, error) {
	finfo, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &GeneFinder{
		f:       f,
		fsize:   int(finfo.Size()),
		bufpool: bufferpool.New(DefaultBufferSize),
	}, nil
}

func validateTemplate(s string) error {
	if !strings.HasPrefix(s, GenePrefix) {
		return &ValidationError{fmt.Sprintf("missing gene prefix: %s", GenePrefix)}
	}

	for _, c := range s[len(GenePrefix):] {
		if !strings.ContainsRune(LegalChars, c) {
			return &ValidationError{"invalid gene template"}
		}
	}

	return nil
}

// Find searches for a gene sequence concurrently.
// it returns whether the sequence was found and any errors if found.
//
// If requested gene sequence size is greater than DefaultBufferSize
// a new buffer pool is allocated for the specific call,
// it is discarded when the routine finishes.
func (gf *GeneFinder) Find(gene string) (bool, error) {
	if len(gene) > gf.fsize {
		return false, &ValidationError{"gene sequence larger than file"}
	}
	if err := validateTemplate(gene); err != nil {
		return false, err
	}

	bufpool := gf.bufpool
	if len(gene) > bufpool.BufferSize() {
		bufpool = bufferpool.New(len(gene) * 2)
	}
	bufsize := bufpool.BufferSize()

	overlapped := bufsize - len(gene)
	chunks := int(math.Ceil(float64(gf.fsize) / float64(overlapped)))

	found := make(chan bool, chunks)
	genebytes := []byte(gene)

	// used for cooperative goroutines cancellation
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	defer cancel()

	for offset := 0; offset < gf.fsize; offset += overlapped {
		offset := int64(offset)
		g.Go(func() error {
			res, e := make(chan bool, 1), make(chan error, 1)
			go func() {
				buf := bufpool.Get()
				defer bufpool.Put(buf)

				_buf := buf.Bytes()
				n, err := gf.f.ReadAt(_buf, offset)
				if err != nil && err != io.EOF {
					e <- err
					return
				}

				res <- bytes.Contains(_buf[:n], genebytes)
			}()

			select {
			case <-ctx.Done():
				return nil
			case f := <-res:
				if f {
					found <- true
					cancel()
				}
			case e := <-e:
				defer cancel()
				return e
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return false, err
	}
	close(found)
	return <-found, nil
}
