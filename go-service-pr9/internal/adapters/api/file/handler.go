package file

import (
	"PR9/internal/domain/file"
	db "PR9/internal/domain/file/mongodb"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
)

type handler struct {
	storage file.Storage
}

type Handler interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetOne(w http.ResponseWriter, r *http.Request)
	GetOneInfo(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

func NewHandler(ctx context.Context) Handler {
	return &handler{
		storage: db.NewStorage(),
	}
}

type data struct {
	FileInfo   file.File
	FileObject multipart.File
}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {

	// 10 MB - максимальный размер файла
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	formFile, header, err := r.FormFile("file") // "file" - имя поля формы для файла
	if err != nil {
		http.Error(w, "Unable to retrieve file", http.StatusBadRequest)
		return
	}

	f := data{
		FileObject: formFile,
	}
	defer f.FileObject.Close()
	id, err := h.storage.Insert(context.Background(), f.FileObject, header)
	if err != nil {
		http.Error(w, "Failed to insert", http.StatusNotFound)
		log.Println(fmt.Sprintf("%s %v", "Failed to insert:", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Data has been succecfully saved in MongoDB with ID=%s", id)))
}

func (h *handler) GetAll(w http.ResponseWriter, r *http.Request) {
	files, err := h.storage.FindAll(context.Background())
	if err != nil {
		http.Error(w, "Failed to find all files", http.StatusNotFound)
		log.Println(fmt.Sprintf("%s %v", "Failed to find all files:", err))
		return
	}
	encoded, err := json.Marshal(files)
	if err != nil {
		http.Error(w, "Failed to decode data", http.StatusNotFound)
		log.Println(fmt.Sprintf("%s %v", "Failed to decode data:", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(encoded)
}

func (h *handler) GetOne(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id")

	fileTemp, err := h.storage.FindOne(context.Background(), id)
	if err != nil {
		http.Error(w, "Failed to find file", http.StatusNotFound)
		log.Println(fmt.Sprintf("%s %v", "Failed to find all files:", err))
		return
	}
	encoded, err := json.Marshal(fileTemp)
	if err != nil {
		http.Error(w, "Failed to decode data", http.StatusNotFound)
		log.Println(fmt.Sprintf("%s %v", "Failed to decode data:", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(encoded)
}

func (h *handler) GetOneInfo(w http.ResponseWriter, r *http.Request) {
	//w.WriteHeader()
	//w.Write()
}

func (h *handler) Update(w http.ResponseWriter, r *http.Request) {
	fileInfo := file.File{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&fileInfo); err != nil {
		http.Error(w, "Failed to decode data", http.StatusBadRequest)
		log.Println(fmt.Sprintf("%s %v", "Failed to update:", err))
		return
	}

	query := r.URL.Query()
	id := query.Get("id")

	err := h.storage.Update(context.Background(), fileInfo.Name, id)
	if err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		log.Println(fmt.Sprintf("%s %v", "Failed to update:", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Data has been succecfully updated in MongoDB with ID=%s", id)))
}

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id")

	if err := h.storage.Delete(context.Background(), id); err != nil {
		http.Error(w, "Failed to delete", http.StatusNotFound)
		log.Println(fmt.Sprintf("Failed to delete id=%s :%v", id, err))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Data has been succecfully deleted from MongoDB with ID=%s", id)))
}
