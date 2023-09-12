package store

import (
	"fmt"
	"io"
	"os"
)

type Serializable interface {
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
}

func WriteObjectToFile(filename string, objects []Serializable) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, obj := range objects {
		err := obj.Serialize(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadObjectFromFile[T Serializable](filepath string) ([]T, error) {
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	file, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var objects []T
	for {
		var obj T
		fmt.Println("reading row")
		if err := obj.Deserialize(file); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}
	return objects, nil
}
