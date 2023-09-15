package store

import (
	"bufio"
	"fmt"
	"os"
)

type Serializable interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

func WriteObjectToFile(filename string, objects []Serializable) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, obj := range objects {
		data, err := obj.Serialize()
		if err != nil {
			return err
		}

		if _, err := file.Write(data); err != nil {
			return err
		}
		if _, err := file.WriteString("\n"); err != nil {
			return err
		}
	}

	return nil
}

func ReadObjectFromFile[T Serializable](filepath string, createT func() T) ([]T, error) {
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
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		obj := createT()
		data := scanner.Bytes()
		if err := obj.Deserialize(data); err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for _, obj := range objects {
		fmt.Printf("read: %v", obj)
	}

	return objects, nil
}
