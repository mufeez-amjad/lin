package store

import (
	"encoding/binary"
	"io"
	"os"

	linproto "lin_cli/internal/proto"

	"google.golang.org/protobuf/proto"
)

func WriteProtobufToFile(filename string, messages []*linproto.Issue) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, msg := range messages {
		data, err := proto.Marshal(msg)
		if err != nil {
			return err
		}

		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(len(data)))

		if _, err := file.Write(buf); err != nil {
			return err
		}

		if _, err := file.Write(data); err != nil {
			return err
		}
	}

	return nil
}

func ReadProtobufFromFile(filepath string) ([]*linproto.Issue, error) {
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

	var offset int64
	content := make([][]byte, 0)
	for {
		buf := make([]byte, 4)
		if _, err := file.ReadAt(buf, offset); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		itemSize := binary.LittleEndian.Uint32(buf)
		offset += 4

		item := make([]byte, itemSize)
		if _, err := file.ReadAt(item, offset); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		content = append(content, item)
		offset += int64(itemSize)
	}

	var messages []*linproto.Issue

	for _, item := range content {
		t := new(linproto.Issue)
		err = proto.Unmarshal(item, t)
		if err != nil {
			return nil, err
		}
		messages = append(messages, t)
	}

	return messages, nil
}
