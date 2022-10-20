package models

type UserTypes string

const (
	DIRECTOR          = "f"
	DIRECTIVE         = "e"
	TEACHER           = "d"
	ATTORNEY          = "c"
	STUDENT_DIRECTIVE = "b"
	STUDENT           = "a"
)

type User struct {
	Name           string `json:"name,omitempty" bson:"name"`
	FirstLastname  string `json:"first_lastname,omitempty" bson:"first_lastname"`
	SecondLastname string `json:"second_lastname,omitempty" bson:"second_lastname"`
	ID             string `json:"_id,omitempty" bson:"_id"`
}
