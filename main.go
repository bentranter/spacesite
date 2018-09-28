package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go"
)

var (
	wd     string
	client *minio.Client

	endpoint        = "nyc3.digitaloceanspaces.com" // TODO(ben) don't hardcode.
	accessKeyID     = os.Getenv("SPACES_KEY")
	secretAccessKey = os.Getenv("SPACES_SECRET")
	bucketName      = os.Getenv("SPACES_BUCKET")
)

func main() {
	validateDir()
	createClient()
	uploadDir()
}

// validateDir ensures that there's at least an `index.html` file in the
// current working directory.
func validateDir() {
	// Get the working directory, and ensure that there's a file named
	// "index.html" somewhere in there.
	cwd, err := os.Getwd()
	if err != nil {
		panic("could not determine working directory")
	}

	foundIndexFile := false
	if err := filepath.Walk(
		cwd,
		func(path string, info os.FileInfo, err error) error {
			// Ignore directories.
			if info == nil || info.IsDir() {
				return nil
			}

			// Get the path relative to our root template directory.
			filename, err := filepath.Rel(cwd, path)
			if err != nil {
				return err
			}
			log.Printf("[info] filename: %s", filename)

			if filename == "index.html" {
				foundIndexFile = true
			}
			return nil
		},
	); err != nil {
		log.Fatalln(err)
	}

	if !foundIndexFile {
		log.Fatalln("directory does not contain an index.html file")
	}

	wd = cwd
}

// createClient creates a client that connects to the Digital Ocean Spaces
// API.
func createClient() {
	// Initiate a client using DigitalOcean Spaces.
	mc, err := minio.New(endpoint, accessKeyID, secretAccessKey, true)
	if err != nil {
		log.Fatalln(err)
	}

	// Check to see if your bucket exists or not.
	foundBucket := false
	buckets, err := mc.ListBuckets()
	if err != nil {
		log.Fatalln(err)
	}
	for _, b := range buckets {
		if b.Name == bucketName {
			foundBucket = true
		}
	}

	// Create a new Space for your static site.
	if !foundBucket {
		if err := mc.MakeBucket(bucketName, "us-east-1"); err != nil {
			log.Fatalln(err)
		}
	}

	// Save the client to use it later.
	client = mc
}

// uploadDir uploads all files in the current directory to Digital Ocean
// Spaces.
func uploadDir() {
	// Upload everything in the current directory to that bucket.
	if err := filepath.Walk(
		wd,
		func(path string, info os.FileInfo, _ error) error {
			// Ignore directories.
			if info == nil || info.IsDir() {
				return nil
			}

			// Open the file.
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			// Read the file.
			data, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}

			// Get the actual filename.
			filename, err := filepath.Rel(wd, path)
			if err != nil {
				return err
			}

			// Check the file extension. For CSS files, it will sometimes fail
			// to detect that it's text/css, and will instead use text/plain.
			// Safari won't parse a stylesheet with that Content-Type if we're
			// in strict mode.
			// Determine the file extension.
			ext := ""
			if strings.Index(filename, ".") != -1 {
				ext = filepath.Ext(filename)
			}

			// Detect the content type.
			contentType := http.DetectContentType(data)
			log.Printf("[info] file extension: %s content type: %s", ext, contentType)

			switch ext {
			case ".css":
				if contentType == "text/plain; charset=utf-8" {
					contentType = "text/css; charset=utf-8"
				}

			case ".js":
				if contentType == "text/plain; charset=utf-8" {
					contentType = "application/javascript; charset=utf-8"
				}
			}

			// Upload the object, and make it public.
			_, err = client.PutObject(
				bucketName,
				filename,
				bytes.NewReader(data),
				info.Size(),
				minio.PutObjectOptions{
					UserMetadata: map[string]string{
						"x-amz-acl": "public-read",
					},
					ContentType: contentType,
				},
			)
			return err
		},
	); err != nil {
		log.Fatalf("[error] error uploading objects: %v", err)
	}

	log.Printf("[info] upload successful, check bucket %s", bucketName)
}
