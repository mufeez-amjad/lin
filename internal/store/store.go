package store

import (
	"bufio"
	"os"
	"time"

	"github.com/adrg/xdg"
)

type Serializable interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

func getCache(filename string) (string, error) {
	return xdg.CacheFile("/lin/" + filename)
}

func WriteObjectToFile(filename string, objects []Serializable) error {
	path, err := getCache(filename)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
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

func ReadObjectFromFile[T Serializable](filename string, createT func() T) ([]T, time.Time, error) {
	path, err := getCache(filename)
	if err != nil {
		return nil, time.Time{}, err
	}

	var stat os.FileInfo
	stat, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, time.Time{}, nil
		} else {
			return nil, time.Time{}, err
		}
	}

	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, stat.ModTime(), err
	}
	defer file.Close()

	var objects []T
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		obj := createT()
		data := scanner.Bytes()
		if err := obj.Deserialize(data); err != nil {
			return nil, stat.ModTime(), err
		}
		objects = append(objects, obj)
	}

	if err := scanner.Err(); err != nil {
		return nil, stat.ModTime(), err
	}

	return objects, stat.ModTime(), nil
}
