package file

import "go.mongodb.org/mongo-driver/bson/primitive"

type File struct {
	ID   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name"`
	Size string             `json:"size" bson:"size"`
	Date string             `json:"date" bson:"date"`
}

//type CreateFileDTO struct {
//	Name string `json:"name"`
//	Size string `json:"size"`
//	Date string `json:"date"`
//}
