package pkg

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"time"
)

func Compress(files []string) io.Reader {
	// Create a pipe.
	tarReader, tarWriter := io.Pipe()
	gzipReader, gzipWriter := io.Pipe()

	// Use a goroutine to write the tar.gz file to the pipe.
	go func() {
		// Create a new tar archive.
		tw := tar.NewWriter(tarWriter)

		// Add some files to the archive.
		var files = []struct {
			Name, Body string
		}{
			{"readme.txt", "This archive contains some text files."},
			{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
			{"todo.txt", "Get animal handling licence.\nWrite more Go."},
		}
		for _, file := range files {
			hdr := &tar.Header{
				Name:    file.Name,
				Size:    int64(len(file.Body)),
				Mode:    0600,
				ModTime: time.Now(),
			}
			if err := tw.WriteHeader(hdr); err != nil {
				log.Fatalln(err)
			}
			if _, err := tw.Write([]byte(file.Body)); err != nil {
				log.Fatalln(err)
			}
		}
		// Make sure to check the error on Close.
		if err := tw.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	go func() {
		// Gzip the tar archive.
		gw := gzip.NewWriter(gzipWriter)
		if _, err := io.Copy(gw, tarReader); err != nil {
			log.Fatalln(err)
		}
		if err := gw.Close(); err != nil {
			log.Fatalln(err)
		}

		// Close the pipe writer to signal the end of the tar.gz file.
		if err := gzipWriter.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	return gzipReader
	//
	//// Read the tar.gz file from the pipe and print its contents.
	//buf := new(strings.Builder)
	//if _, err := io.Copy(buf, gzipReader); err != nil {
	//	log.Fatalln(err)
	//}
	//fmt.Println(buf.String())
}

type DecompressedFile struct {
	Name string
	Body io.Reader
}

// Decompress is broken this is all just WIP
// TODO(manuel, 2023-01-05): fix this, not sure if this is even really that needed
func Decompress(ctx *context.Context, r io.Reader, out chan (*DecompressedFile)) error {
	defer close(out)

	// Create a new gzip reader.
	gr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gr.Close()

	tarReader, tarWriter := io.Pipe()

	// Create a new tar reader.
	tr := tar.NewReader(tarReader)

	tarDecompress := func() error {
		// Iterate through the files in the archive.
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				// end of tar archive
				break
			}
			if err != nil {
				return err
			}
			out <- &DecompressedFile{
				Name: hdr.Name,
				Body: tr,
			}
		}
		return nil
	}

	gzipDecompress := func() error {
		if _, err := io.Copy(tarWriter, gr); err != nil {
			return err
		}

		return nil
	}

	errGroup := errgroup.Group{}
	errGroup.Go(gzipDecompress)
	errGroup.Go(tarDecompress)

	return errGroup.Wait()

	// Iterate through the files in the archive.
}
