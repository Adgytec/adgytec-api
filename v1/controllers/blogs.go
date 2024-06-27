package controllers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/services"
)

func GetUUID(w http.ResponseWriter, r *http.Request) {
	id := generateUUID()

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = struct {
		UUID uuid.UUID `json:"uuid"`
	}{
		UUID: id,
	}

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func PostMedia(w http.ResponseWriter, r *http.Request) {
	// projectId := chi.URLParam(r, "projectId")
	// blogId := chi.URLParam(r, "blogId")
	maxSize := 150 << 20 // 150 mb

	err := helper.ParseMultipartForm(w, r, maxSize)
	if err != nil {
		return
	}

	if _, ok := r.MultipartForm.Value["metadata"]; !ok {
		message := "Missing required field: 'metadata'"
		helper.HandleError(w, &custom.MalformedRequest{
			Status:  http.StatusBadRequest,
			Message: message,
		})
		return
	}

	var bm services.BlogMedia
	err, success := bm.UploadMedia(r)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false

	msg := "uploaded media files"
	if !success {
		msg += ", but with exceptions"
	}
	payload.Message = msg

	helper.EncodeJSON(w, http.StatusCreated, payload)

}

func DeleteMedia(w http.ResponseWriter, r *http.Request) {
	mediaDetails, err := helper.DecodeJSON[services.BlogMedia](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	err = mediaDetails.DeleteMedia()
	if err != nil {
		helper.HandleError(w, err)
		return
	}
	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully deleted the media files"

	helper.EncodeJSON(w, http.StatusCreated, payload)
}

func PostBlog(w http.ResponseWriter, r *http.Request) {
	maxSize := 20 << 20 // 20mb
	err := helper.ParseMultipartForm(w, r, maxSize)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	projectId := chi.URLParam(r, "projectId")
	blogId := chi.URLParam(r, "blogId")
	userId := r.Context().Value(custom.UserID).(string)

	requiredFields := []string{"title", "content", "author"}
	requiredFileFields := "cover"

	fmt.Println("running1")

	for _, field := range requiredFields {
		if _, ok := r.MultipartForm.Value[field]; !ok {
			message := fmt.Sprintf("Missing required field: %s", field)
			helper.HandleError(w, &custom.MalformedRequest{
				Status:  http.StatusBadRequest,
				Message: message,
			})
			return
		}
	}

	fmt.Println("running2")

	if _, ok := r.MultipartForm.File[requiredFileFields]; !ok {
		message := fmt.Sprintf("Missing required file: %s", requiredFileFields)
		helper.HandleError(w, &custom.MalformedRequest{
			Status:  http.StatusBadRequest,
			Message: message,
		})
		return
	}

	fmt.Println("running3")

	title := r.FormValue("title")
	summary := r.FormValue("summary")
	content := r.FormValue("content")
	author := r.FormValue("author")

	var blogItem services.Blog
	blogItem.Title = title
	blogItem.Summary = summary
	blogItem.Id = blogId
	blogItem.Content = content
	blogItem.Author = author

	err = blogItem.CreateBlog(r, projectId, userId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully created the blog"

	helper.EncodeJSON(w, http.StatusCreated, payload)
}

// only title, author, created_at, summary, cover image
func GetAllBlogsByProjectId(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")

	var blogs services.Blog
	all, err := blogs.GetBlogsByProjectId(projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)

}

func GetBlogById(w http.ResponseWriter, r *http.Request) {
	blogId := chi.URLParam(r, "blogId")

	var blogData services.Blog
	blogData.Id = blogId

	blog, err := blogData.GetBlogById()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse

	payload.Error = false
	payload.Data = blog

	helper.EncodeJSON(w, http.StatusOK, payload)
}
