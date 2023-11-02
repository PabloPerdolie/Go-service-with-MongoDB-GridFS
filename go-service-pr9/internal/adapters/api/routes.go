package api

import (
	"PR9/internal/adapters/api/file"
	"context"
	"github.com/gorilla/mux"
)

func SetupRoutes(r *mux.Router) {

	handler := file.NewHandler(context.Background())

	r.HandleFunc("/files", handler.Create).Methods("POST")
	r.HandleFunc("/files", handler.GetAll).Methods("GET")
	r.HandleFunc("/files/down", handler.GetOne).Methods("GET")
	r.HandleFunc("/files/info", handler.GetOneInfo).Methods("GET")
	r.HandleFunc("/files/upd", handler.Update).Methods("PUT")
	r.HandleFunc("/files/del", handler.Delete).Methods("DELETE")
}

//GET /files - получение списка файлов
//GET /files/{id} - файла по id
//GET /files/{id}/info - получение информации о файле по id
//POST /files - загрузка файла
//UPDATE /files/{id} - обновление файла по id
//DELETE /files/{id} - удаление файла по id
