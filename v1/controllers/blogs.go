package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/services"
)

func GetUUID(w http.ResponseWriter, r *http.Request) {
	id := services.GenerateUUID()

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
	maxSize := 25 << 20 // 150 mb

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
	maxSize := 15 << 20 // 20mb
	err := helper.ParseMultipartForm(w, r, maxSize)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	projectId := chi.URLParam(r, "projectId")
	blogId := chi.URLParam(r, "blogId")
	userId := r.Context().Value(custom.UserID).(string)

	requiredFields := []string{"title", "content", "category"}
	requiredFileFields := "cover"

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

	title := r.FormValue("title")
	summary := r.FormValue("summary")
	content := r.FormValue("content")
	author := r.FormValue("author")
	category := r.FormValue("category")

	var blogItem services.Blog
	blogItem.Title = title
	blogItem.Summary = summary
	blogItem.Id = blogId
	blogItem.Content = content
	blogItem.Author = author
	blogItem.Category = category

	if _, ok := r.MultipartForm.File[requiredFileFields]; !ok {
		// message := fmt.Sprintf("Missing required file: %s", requiredFileFields)
		// helper.HandleError(w, &custom.MalformedRequest{
		// 	Status:  http.StatusBadRequest,
		// 	Message: message,
		// })
		// return
		err = blogItem.CreateBlogWithoutCover(projectId, userId)
	} else {
		err = blogItem.CreateBlog(r, projectId, userId)
	}

	// err = blogItem.CreateBlog(r, projectId, userId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully added new blog"

	helper.EncodeJSON(w, http.StatusCreated, payload)
}

// only title, author, created_at, summary, cover image
func GetAllBlogsByProjectId(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")
	cursor := r.URL.Query().Get("cursor")

	if len(cursor) == 0 {
		cursor = getNow()
	}

	var blogs services.Blog
	all, err := blogs.GetBlogsByProjectId(projectId, cursor)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)

}

func GetAllBlogsByProjectIdClient(w http.ResponseWriter, r *http.Request) {
	projectId := r.Context().Value(custom.ProjectId).(string)
	cursor := r.URL.Query().Get("cursor")

	log.Printf("cursor value: %v\n", cursor)
	if len(cursor) == 0 {
		cursor = getNow()
	}

	var blogs services.Blog
	all, err := blogs.GetBlogsByProjectId(projectId, cursor)
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

func PatchBlogMetadataById(w http.ResponseWriter, r *http.Request) {
	blogId := chi.URLParam(r, "blogId")

	blogDetails, err := helper.DecodeJSON[services.BlogMetadata](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	blogDetails.Id = blogId
	err = blogDetails.PatchBlogMetadataById()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully updated blog data"

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func DeleteBlogById(w http.ResponseWriter, r *http.Request) {
	blogId := chi.URLParam(r, "blogId")
	projectId := chi.URLParam(r, "projectId")

	var blog services.Blog
	blog.Id = blogId

	err := blog.DeleteBlogById(projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully deleted the blog item"

	helper.EncodeJSON(w, http.StatusOK, payload)

}

func PatchBlogCover(w http.ResponseWriter, r *http.Request) {
	maxSize := 10 << 20 // 10mb
	err := helper.ParseMultipartForm(w, r, maxSize)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	requiredFileFields := "cover"
	if _, ok := r.MultipartForm.File[requiredFileFields]; !ok {
		message := fmt.Sprintf("Missing required file: %s", requiredFileFields)
		helper.HandleError(w, &custom.MalformedRequest{
			Status:  http.StatusBadRequest,
			Message: message,
		})
		return
	}

	projectId := chi.URLParam(r, "projectId")
	blogId := chi.URLParam(r, "blogId")

	var blog services.Blog

	blog.Id = blogId
	err = blog.PatchBlogCover(r, projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully updated blog cover image"

	helper.EncodeJSON(w, http.StatusOK, payload)

}

func PatchBlogContent(w http.ResponseWriter, r *http.Request) {
	blogId := chi.URLParam(r, "blogId")

	blogContent, err := helper.DecodeJSON[services.Blog](w, r, mb*10)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	blogContent.Id = blogId
	err = blogContent.PatchBlogContent()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully updated blog content"

	helper.EncodeJSON(w, http.StatusOK, payload)
}
