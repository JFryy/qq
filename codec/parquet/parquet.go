package parquet

import (
	"bytes"
	"context"
	"fmt"
	"github.com/apache/arrow/go/v16/arrow"
	"github.com/apache/arrow/go/v16/arrow/array"
	"github.com/apache/arrow/go/v16/arrow/memory"
	"github.com/apache/arrow/go/v16/parquet"
	"github.com/apache/arrow/go/v16/parquet/compress"
	"github.com/apache/arrow/go/v16/parquet/file"
	"github.com/apache/arrow/go/v16/parquet/pqarrow"
	"github.com/goccy/go-json"
	"reflect"
)

type Codec struct{}

func (c *Codec) Marshal(v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil, fmt.Errorf("input data must be a slice")
	}

	if rv.Len() == 0 {
		return nil, fmt.Errorf("no data to write")
	}

	firstElem := rv.Index(0).Interface()
	firstElemValue, ok := firstElem.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("slice elements must be of type map[string]any")
	}

	mem := memory.NewGoAllocator()
	var fields []arrow.Field

	for key := range firstElemValue {
		fields = append(fields, arrow.Field{Name: key, Type: arrow.BinaryTypes.String, Nullable: true})
	}

	schema := arrow.NewSchema(fields, nil)

	var buf bytes.Buffer
	props := parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Snappy))
	arrowProps := pqarrow.NewArrowWriterProperties(pqarrow.WithStoreSchema())

	writer, err := pqarrow.NewFileWriter(schema, &buf, props, arrowProps)
	if err != nil {
		return nil, fmt.Errorf("error creating parquet writer: %v", err)
	}
	defer writer.Close()

	builders := make([]array.Builder, len(fields))
	for i := range fields {
		builders[i] = array.NewStringBuilder(mem)
	}

	for i := 0; i < rv.Len(); i++ {
		recordMap := rv.Index(i).Interface().(map[string]any)
		for j, field := range fields {
			if value, ok := recordMap[field.Name]; ok {
				builders[j].(*array.StringBuilder).Append(fmt.Sprintf("%v", value))
			} else {
				builders[j].(*array.StringBuilder).AppendNull()
			}
		}
	}

	columns := make([]arrow.Array, len(builders))
	for i, builder := range builders {
		columns[i] = builder.NewArray()
	}

	record := array.NewRecord(schema, columns, int64(rv.Len()))
	defer record.Release()

	if err := writer.Write(record); err != nil {
		return nil, fmt.Errorf("error writing record to parquet: %v", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error closing parquet writer: %v", err)
	}

	return buf.Bytes(), nil
}

func (c *Codec) Unmarshal(input []byte, v any) error {
	reader := bytes.NewReader(input)

	parquetFile, err := file.NewParquetReader(reader)
	if err != nil {
		return fmt.Errorf("error creating parquet reader: %v", err)
	}
	defer parquetFile.Close()

	fileReader, err := pqarrow.NewFileReader(parquetFile, pqarrow.ArrowReadProperties{}, memory.NewGoAllocator())
	if err != nil {
		return fmt.Errorf("error creating arrow file reader: %v", err)
	}

	// Read the whole table
	table, err := fileReader.ReadTable(context.Background())
	if err != nil {
		return fmt.Errorf("error reading table: %v", err)
	}
	defer table.Release()

	// Convert table to records
	tableReader := array.NewTableReader(table, 1000) // batch size
	defer tableReader.Release()

	var records []map[string]any

	for tableReader.Next() {
		record := tableReader.Record()
		schema := record.Schema()
		numRows := record.NumRows()
		numCols := record.NumCols()

		for i := int64(0); i < numRows; i++ {
			rowMap := make(map[string]any)
			for j := 0; j < int(numCols); j++ {
				field := schema.Field(j)
				column := record.Column(j)

				if column.IsNull(int(i)) {
					rowMap[field.Name] = nil
				} else {
					switch arr := column.(type) {
					case *array.String:
						rowMap[field.Name] = arr.Value(int(i))
					case *array.Int64:
						rowMap[field.Name] = arr.Value(int(i))
					case *array.Float64:
						rowMap[field.Name] = arr.Value(int(i))
					case *array.Boolean:
						rowMap[field.Name] = arr.Value(int(i))
					case *array.Int32:
						rowMap[field.Name] = arr.Value(int(i))
					case *array.Float32:
						rowMap[field.Name] = arr.Value(int(i))
					default:
						rowMap[field.Name] = fmt.Sprintf("%v", column.GetOneForMarshal(int(i)))
					}
				}
			}
			records = append(records, rowMap)
		}
	}

	// Always use JSON marshaling for consistent type handling
	jsonData, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	if err := json.Unmarshal(jsonData, v); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return nil
}
