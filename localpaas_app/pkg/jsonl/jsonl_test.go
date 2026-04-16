package jsonl

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestMetadataAndChunk(t *testing.T) {
	tm := time.Now().Round(time.Second)
	m := Metadata{
		Name:      "test",
		Type:      "backup",
		Version:   "1.0",
		Timestamp: tm,
		Note:      "some note",
	}
	assert.Equal(t, "test", m.Name)

	data := testData{ID: 1, Name: "one"}
	chunk := NewChunk("data_type", data)
	assert.Equal(t, "data_type", chunk.Type)
	assert.Equal(t, data, chunk.Data)
}

func TestWriterReader(t *testing.T) {
	buf := new(bytes.Buffer)
	writer := NewWriter(buf)

	metadata := Metadata{Name: "meta"}
	data1 := testData{ID: 1, Name: "one"}
	data2 := testData{ID: 2, Name: "two"}

	err := writer.WriteMetadata(metadata)
	assert.NoError(t, err)
	err = writer.WriteChunk(NewChunk("data", data1))
	assert.NoError(t, err)
	err = writer.Write(data2)
	assert.NoError(t, err)

	reader := NewReader(buf)

	// Read metadata
	var m Metadata
	err = reader.ReadSingleLine(&m)
	assert.NoError(t, err)
	assert.Equal(t, metadata.Name, m.Name)

	// Read data1 via chunk
	var c Chunk[testData]
	err = reader.ReadSingleLine(&c)
	assert.NoError(t, err)
	assert.Equal(t, "data", c.Type)
	assert.Equal(t, data1, c.Data)

	// Read data2
	var d testData
	err = reader.ReadSingleLine(&d)
	assert.NoError(t, err)
	assert.Equal(t, data2, d)

	// EOF
	err = reader.ReadSingleLine(&d)
	assert.ErrorIs(t, err, ErrScannerNotReadable)
}

func TestReader_ReadLines(t *testing.T) {
	buf := bytes.NewBufferString(`{"id":1}
{"id":2}
`)
	reader := NewReader(buf)

	var ids []int
	err := reader.ReadLines(func(data []byte) error {
		var d struct{ ID int }
		if err := json.Unmarshal(data, &d); err != nil {
			return err
		}
		ids = append(ids, d.ID)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2}, ids)
}

type mockCloser struct {
	io.Writer
	closed bool
}

func (m *mockCloser) Close() error {
	m.closed = true
	return nil
}

func TestWriterReader_Close(t *testing.T) {
	t.Run("Writer Close", func(t *testing.T) {
		m := &mockCloser{Writer: new(bytes.Buffer)}
		writer := NewWriter(m)
		err := writer.Close()
		assert.NoError(t, err)
		assert.True(t, m.closed)
	})

	t.Run("Writer Not Closeable", func(t *testing.T) {
		writer := NewWriter(new(bytes.Buffer))
		err := writer.Close()
		assert.ErrorIs(t, err, ErrNotCloseable)
	})
}
