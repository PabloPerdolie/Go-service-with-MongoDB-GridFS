package db

import (
	"PR9/internal/domain/file"
	"PR9/pkg/client/mongodb"
	"bytes"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"mime/multipart"
	"os"
	"time"
)

type db struct {
	collection *mongo.Collection
	fs         *gridfs.Bucket
}

func NewStorage() file.Storage {
	client, bucket, err := mongodb.NewClient(context.Background(), os.Getenv("DATABASE"))
	if err != nil {
		panic(err)
		return nil
	}
	return &db{
		collection: client.Collection(os.Getenv("COLLECTION")),
		fs:         bucket,
	}
}

func (d *db) Insert(ctx context.Context, fileData multipart.File, header *multipart.FileHeader) (fileID string, err error) {
	newFile, err := os.Create(os.Getenv("PATH_INPUT") + header.Filename)
	if err != nil {
		return "", fmt.Errorf("failed to create a new file: %v", err)
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, fileData)
	if err != nil {
		return "", fmt.Errorf("failed to save file")
	}

	fileTemp, err := os.Open(os.Getenv("PATH_INPUT") + header.Filename)
	uploadOpts := options.GridFSUpload().SetMetadata(bson.D{{"size", header.Size}})
	id, err := d.fs.UploadFromStream(header.Filename, io.Reader(fileTemp), uploadOpts)
	if err != nil {
		return "", fmt.Errorf("failed to save file in GridFS: %v", err)
	}

	resFile := file.File{
		ID:   id,
		Name: header.Filename,
		Size: fmt.Sprintf("%d", header.Size),
		Date: fmt.Sprintf("%v", time.Now()),
	}

	result, err := d.collection.InsertOne(ctx, resFile)
	if err != nil {
		return "", fmt.Errorf("failed to insert file: %v", err)
	}
	oid, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		log.Println(fmt.Sprintf("INSERT file=%v", resFile))
		return oid.Hex(), nil
	}

	return "", fmt.Errorf("failed to convert objectid to hex")
}

func (d *db) FindOne(ctx context.Context, id string) (f file.File, err error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return f, fmt.Errorf("failed to convert to objectid: %v", err)
	}
	filter := bson.M{"_id": oid}
	result := d.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		// TODO 404
		return f, fmt.Errorf("failed to find by id: %v", result.Err())
	}
	if err = result.Decode(&f); err != nil {
		return f, fmt.Errorf("failed to decode file data from db: %v", err)
	}

	var buf bytes.Buffer
	_, err = d.fs.DownloadToStream(oid, &buf)
	if err != nil {
		return f, err
	}

	newFile, err := os.Create(os.Getenv("PATH_OUT") + f.Name)
	if err != nil {
		return f, fmt.Errorf("failed to create a new file: %v", err)
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, &buf)
	if err != nil {
		return f, fmt.Errorf("failed to save file")
	}

	log.Println(fmt.Sprintf("FIND AND DOWNLOAD file=%v", f))

	return f, nil
}

func (d *db) FindAll(ctx context.Context) (files []file.File, err error) {
	result, err := d.collection.Find(ctx, bson.D{{}})
	if result.Err() != nil {
		return files, fmt.Errorf("failed to find all files: %v", result.Err())
	}

	if err = result.All(ctx, &files); err != nil {
		return files, fmt.Errorf("failed to read file data from cursor: %v", err)
	}

	log.Println(fmt.Sprintf("FIND files=%v", files))

	return files, nil
}

func (d *db) Update(ctx context.Context, name string, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("failed to convert from hex: %v", err)
	}

	if err = d.fs.Rename(oid, name); err != nil {
		return fmt.Errorf("failed to rename file id=%s, %v", id, err)
	}

	resFile := file.File{
		Name: name,
	}

	fileBytes, err := bson.Marshal(resFile)
	if err != nil {
		return fmt.Errorf("failed to marshal file data %v", err)
	}

	filter := bson.M{"_id": oid}

	var updateFileObj bson.M

	if err = bson.Unmarshal(fileBytes, &updateFileObj); err != nil {
		return fmt.Errorf("failed to unmarshal filebytes: %v", err)
	}

	delete(updateFileObj, "_id")
	delete(updateFileObj, "size")
	delete(updateFileObj, "date")

	update := bson.M{
		"$set": updateFileObj,
	}

	result, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update file obj: %v", err)
	}

	if result.MatchedCount == 0 {
		// todo 404
		return fmt.Errorf("file not found")
	}

	log.Println(fmt.Sprintf("UPDATE matched %d docs and modified %d documents", result.MatchedCount, result.ModifiedCount))

	return nil
}

func (d *db) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("failed to convert from hex: %v", err)
	}

	filter := bson.M{"_id": oid}
	result, err := d.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete file from MongoDB by id: %v", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("file not found")
	}
	fmt.Printf("Deleted %d documents", result.DeletedCount)

	if err := d.fs.Delete(oid); err != nil {
		return fmt.Errorf("failed to delete file from GridFS by id: %v", err)
	}
	log.Println(fmt.Sprintf("DELETE with id=%s", id))
	return nil
}
