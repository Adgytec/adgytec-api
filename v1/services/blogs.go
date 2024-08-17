package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/minio/minio-go/v7"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
	"golang.org/x/net/html"
)

type FileMetaData struct {
	Path string `json:"path"`
}

type BlogMedia struct {
	Paths []string `json:"paths,omitempty"`
}

type Blog struct {
	Title     string    `json:"title" db:"title"`
	Summary   string    `json:"summary,omitempty" db:"short_text"`
	Content   string    `json:"content,omitempty" db:"content"`
	Author    string    `json:"author" db:"author"`
	Id        string    `json:"blogId" db:"blog_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Cover     string    `json:"cover" db:"cover_image"`
	Category  string    `json:"category" db:"category"`
}

type BlogSummary struct {
	Title     string          `json:"title" db:"title"`
	Summary   string          `json:"summary,omitempty" db:"short_text"`
	Author    string          `json:"author" db:"author"`
	Id        string          `json:"blogId" db:"blog_id"`
	CreatedAt time.Time       `json:"createdAt" db:"created_at"`
	Cover     string          `json:"cover" db:"cover_image"`
	Category  json.RawMessage `json:"category" db:"category"`
}

type BlogMetadata struct {
	Id       string
	Title    string
	Summary  string
	Category string
}

func (bm *BlogMedia) UploadMedia(r *http.Request) (error, bool) {
	metadataJSON := r.FormValue("metadata")
	var metadata []FileMetaData
	err := json.Unmarshal([]byte(metadataJSON), &metadata)
	if err != nil {
		return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: "Invalid file metadata."}, false
	}

	isSuccess := true
	for i, meta := range metadata {
		go func(index int, metadata FileMetaData) {
			file, header, err := r.FormFile(fmt.Sprintf("media_%d", index))
			if err != nil {
				log.Printf("error reteriving file: %v\n", err)
				isSuccess = false
				return
			}
			defer file.Close()

			// contentType := header.Header.Get("Content-type")
			// if !strings.HasPrefix(contentType, "image/") {
			// 	isSuccess = false
			// 	return
			// }

			contentType, err := isImageFile(header)
			if err != nil {
				isSuccess = false
				return
			}

			img, format, err := image.Decode(file)
			if err != nil {
				log.Printf("Error decoding image: %v\n", err)
				isSuccess = false
				return
			}

			buf := new(bytes.Buffer)
			err = handleImage(img, buf, format)
			if err != nil {
				isSuccess = false
				return
			}

			_, err = spaceStorage.PutObject(ctx,
				os.Getenv("SPACE_STORAGE_BUCKET_NAME"),
				metadata.Path,
				buf,
				int64(buf.Len()), minio.PutObjectOptions{ContentType: contentType})
			if err != nil {
				isSuccess = false
				return
			}

		}(i, meta)
	}

	return nil, isSuccess
}

func (bm *BlogMedia) DeleteMedia() error {
	if len(bm.Paths) == 0 {
		return nil
	}

	objectChan := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectChan)
		for _, path := range bm.Paths {
			objectChan <- minio.ObjectInfo{Key: path}
		}
	}()

	e := spaceStorage.RemoveObjects(ctx,
		os.Getenv("SPACE_STORAGE_BUCKET_NAME"),
		objectChan,
		minio.RemoveObjectsOptions{},
	)

	isErr := false
	for err := range e {
		log.Printf("Error deleting objects in space storage, %v\n", err)
		isErr = true
	}

	if isErr {
		return errors.New("error deleting image from space storage")
	}

	return nil
}

func addBlogToDatabase(b *Blog, projectId, userId string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	args := dbqueries.CreateBlogItemArgs(b.Id, userId, projectId, b.Title,
		b.Cover, b.Summary, b.Content, b.Author, b.Category)

	_, err := db.Exec(ctx, dbqueries.CreateBlogItem, args)
	if err != nil {
		log.Printf("Error adding blog item in database: %v\n", err)
	}
	errChan <- err
}

func (b *Blog) CreateBlog(r *http.Request, projectId, userId string) error {
	file, header, err := r.FormFile("cover")
	if err != nil {
		log.Printf("Error retriving file: %v\n ", err)
		return err
	}
	defer file.Close()

	// contentType := header.Header.Get("Content-type")
	// if !strings.HasPrefix(contentType, "image/") {
	// 	return (&custom.MalformedRequest{
	// 		Status:  http.StatusUnsupportedMediaType,
	// 		Message: http.StatusText(http.StatusUnsupportedMediaType),
	// 	})
	// }

	contentType, err := isImageFile(header)
	if err != nil {
		return err
	}

	img, format, err := image.Decode(file)
	if err != nil {
		log.Printf("Error decoding image: %v\n", err)
		return err
	}

	buf := new(bytes.Buffer)
	err = handleImage(img, buf, format)
	if err != nil {
		return err
	}

	objectName := fmt.Sprintf("services/blogs/%v/%v/%v.%v", projectId, b.Id, generateRandomString(), format)

	if val := os.Getenv("ENV"); val == "dev" {
		objectName = "dev/" + objectName
	}
	b.Cover = objectName

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 2)

	wg.Add(2)

	go uploadImageToCloudStorage(objectName, buf, contentType, wg, errChan)
	go addBlogToDatabase(b, projectId, userId, wg, errChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Blog) GetBlogsByProjectId(projectId string) (*[]BlogSummary, error) {
	args := dbqueries.GetBlogsByProjectIdArgs(projectId)
	rows, err := db.Query(ctx, dbqueries.GetBlogsByProjectId, args)

	if err != nil {
		log.Printf("Error fetching blogs from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	blogs, err := pgx.CollectRows(rows, pgx.RowToStructByName[BlogSummary])
	if err != nil {
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	wg := new(sync.WaitGroup)
	urlChan := make(chan IndexedValue, len(blogs))

	for ind, item := range blogs {
		wg.Add(1)

		img := item.Cover
		go generatePresignedUrl(img, ind, expires, wg, urlChan)
	}

	wg.Wait()
	close(urlChan)

	for url := range urlChan {
		ind := url.Index
		blogs[ind].Cover = url.Url
	}

	return &blogs, nil
}

func (b *Blog) GetBlogById() (*Blog, error) {
	args := dbqueries.GetBlogsByIdArgs(b.Id)
	rows, err := db.Query(ctx, dbqueries.GetBlogById, args)
	if err != nil {
		log.Printf("Error fetching blog from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	blog, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Blog])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message := "Blog with the provided ID does not exist."
			return nil, &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
		}
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	cover, err := spaceStorage.PresignedGetObject(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), blog.Cover, week, nil)
	if err != nil {
		log.Printf("error generating presigned url for cover image: %v\n", err)
	} else {
		blog.Cover = cover.String()
	}

	// copied will reread it
	doc, err := html.Parse(bytes.NewReader([]byte(blog.Content)))
	if err != nil {
		log.Printf("error parsing html: %v\n", err)
		return &blog, err
	}

	var updateImgTags func(*html.Node)
	updateImgTags = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			var dataKey string
			for i := 0; i < len(n.Attr); i++ {
				if n.Attr[i].Key == "data-path" {
					dataKey = n.Attr[i].Val
					// n.Attr = append(n.Attr[:i], n.Attr[i+1:]...) // Remove data-key attribute
					break
				}
			}
			if dataKey != "" {
				// Generate presigned URL
				isPresigned := true
				presignedURL, err := spaceStorage.PresignedGetObject(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), dataKey, week, nil)
				if err != nil {
					log.Printf("Can't genrate url for image: %v\n", err)
					isPresigned = false
				}
				// Add or update src attribute
				hasSrc := false
				for i := 0; i < len(n.Attr); i++ {
					if n.Attr[i].Key == "src" {
						if isPresigned {
							n.Attr[i].Val = presignedURL.String()
						} else {
							n.Attr[i].Val = "https://images.unsplash.com/photo-1713171158509-f2a6582581a0?q=80&w=2070&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"
						}
						hasSrc = true
						break
					}
				}
				if !hasSrc {
					if isPresigned {
						n.Attr = append(n.Attr, html.Attribute{Key: "src", Val: presignedURL.String()})
					} else {
						n.Attr = append(n.Attr, html.Attribute{Key: "src", Val: "https://images.unsplash.com/photo-1713171158509-f2a6582581a0?q=80&w=2070&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"})
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			updateImgTags(c)
		}
	}
	updateImgTags(doc)

	var buf bytes.Buffer
	err = html.Render(&buf, doc)
	if err != nil {
		log.Printf("error getting html from buffer: %v\n", err)
		return &blog, nil
	}
	updatedHTMLContent := buf.String()
	blog.Content = updatedHTMLContent

	return &blog, nil
}

func (bm *BlogMetadata) PatchBlogMetadataById() error {
	args := dbqueries.PatchBlogMetadataByIdArgs(bm.Title, bm.Summary, bm.Id, bm.Category)
	_, err := db.Exec(ctx, dbqueries.PatchBlogMetadataById, args)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid blog id to update."
				return &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
			}
		}

		log.Printf("Error updating blog data: %v\n", err)
		return err
	}
	return nil
}

func deleteBlogFromDatabase(b *Blog) error {
	args := dbqueries.DeleteBlogByIdArgs(b.Id)
	_, err := db.Exec(ctx, dbqueries.DeleteBlogById, args)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid blog id to delete."
				return &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
			}
		}

		log.Printf("Error deleting blog data: %v\n", err)
	}

	return err
}

func deleteBlogMedia(projectId, blogId string) {
	mediaPrefix := fmt.Sprintf("services/blogs/%v/%v", projectId, blogId)
	if os.Getenv("ENV") == "dev" {
		mediaPrefix = "dev/" + mediaPrefix
	}
	objectsCh := make(chan minio.ObjectInfo)
	log.Println(mediaPrefix)

	go func() {
		defer close(objectsCh)

		opts := minio.ListObjectsOptions{
			Recursive: true,
			Prefix:    mediaPrefix,
		}
		// List all objects from a bucket-name with a matching prefix.
		for object := range spaceStorage.ListObjects(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), opts) {
			if object.Err != nil {
				log.Printf("error listing object: %v\n", object.Err)
			} else {
				objectsCh <- object
			}
		}
	}()

	opts := minio.RemoveObjectsOptions{}

	for rErr := range spaceStorage.RemoveObjects(context.Background(), os.Getenv("SPACE_STORAGE_BUCKET_NAME"), objectsCh, opts) {
		log.Println("Error detected during deletion: ", rErr)
	}
}

func (b *Blog) DeleteBlogById(projectId string) error {

	err := deleteBlogFromDatabase(b)
	if err == nil {
		go deleteBlogMedia(projectId, b.Id)
	}

	return err
}

func handleBlogCoverDatabase(cover, blogid string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	args := dbqueries.PatchBlogCoverArgs(blogid, cover)
	rows, err := db.Query(ctx, dbqueries.PatchBlogCover, args)
	if err != nil {
		log.Printf("error updating cover image in db: %v\n", err)
		errChan <- err
		return
	}
	defer rows.Close()

	prevPath, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[struct {
		Image string `db:"image"`
	}])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message := "blog with the following id doesn't exist"
			errChan <- &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid blog id."
				errChan <- &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
				return
			}
		}

		log.Printf("Error reading rows: %v\n", err)
		errChan <- nil
		return
	}

	go func() {
		err = spaceStorage.RemoveObject(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), prevPath.Image, minio.RemoveObjectOptions{})
		if err != nil {
			log.Printf("Error deleting image from space storage: %v\n", err)
		}
	}()

	errChan <- nil
}

func (b *Blog) PatchBlogCover(r *http.Request, projectId string) error {
	file, header, err := r.FormFile("cover")
	if err != nil {
		log.Printf("Error retriving file: %v\n ", err)
		return err
	}

	defer file.Close()

	// contentType := header.Header.Get("Content-type")
	// if !strings.HasPrefix(contentType, "image/") {
	// 	return (&custom.MalformedRequest{
	// 		Status:  http.StatusUnsupportedMediaType,
	// 		Message: http.StatusText(http.StatusUnsupportedMediaType),
	// 	})
	// }

	contentType, err := isImageFile(header)
	if err != nil {
		return err
	}

	img, format, err := image.Decode(file)
	if err != nil {
		log.Printf("Error decoding image: %v\n", err)
		return err
	}

	buf := new(bytes.Buffer)
	err = handleImage(img, buf, format)
	if err != nil {
		return err
	}

	objectName := fmt.Sprintf("services/blogs/%v/%v/%v.%v", projectId, b.Id, generateRandomString(), format)

	if val := os.Getenv("ENV"); val == "dev" {
		objectName = "dev/" + objectName
	}
	b.Cover = objectName

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 2)

	wg.Add(2)

	go uploadImageToCloudStorage(objectName, buf, contentType, wg, errChan)
	go handleBlogCoverDatabase(objectName, b.Id, wg, errChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Blog) PatchBlogContent() error {
	args := dbqueries.PatchBlogContentArgs(b.Id, b.Content)
	_, err := db.Exec(ctx, dbqueries.PatchBlogContent, args)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid blog id to update."
				return &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
			}
		}

		log.Printf("error updating blog contnet: %v\n", err)
	}
	return err
}
