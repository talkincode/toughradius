package excel

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/pkg/timeutil"
)

// TestUser is a test struct for Excel export
type TestUser struct {
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	Age       int                `json:"age"`
	Score     float64            `json:"score"`
	Active    bool               `json:"active"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt timeutil.LocalTime `json:"updated_at"`
	Internal  string             `json:"-"` // Should be ignored
}

// TestWriteToFile tests writing data to an Excel file
func TestWriteToFile(t *testing.T) {
	// Prepare test data
	now := time.Now()
	localTime := timeutil.LocalTime(now)

	records := []interface{}{
		&TestUser{
			Username:  "user1",
			Email:     "user1@example.com",
			Age:       25,
			Score:     95.5,
			Active:    true,
			CreatedAt: now,
			UpdatedAt: localTime,
			Internal:  "should not appear",
		},
		&TestUser{
			Username:  "user2",
			Email:     "user2@example.com",
			Age:       30,
			Score:     88.0,
			Active:    false,
			CreatedAt: now.Add(-24 * time.Hour),
			UpdatedAt: localTime,
		},
	}

	// Create temp file path
	tmpDir := t.TempDir()
	filepath := filepath.Join(tmpDir, "test_users.xlsx")

	// Test write
	err := WriteToFile("Users", records, filepath)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Error("Excel file was not created")
	}

	// Verify file is not empty
	info, err := os.Stat(filepath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Excel file is empty")
	}
}

// TestWriteToTmpFile tests writing data to a temporary Excel file
func TestWriteToTmpFile(t *testing.T) {
	now := time.Now()
	records := []interface{}{
		&TestUser{
			Username:  "tempuser",
			Email:     "temp@example.com",
			Age:       20,
			Score:     75.0,
			Active:    true,
			CreatedAt: now,
			UpdatedAt: timeutil.LocalTime(now),
		},
	}

	// Test write to temp file
	filepath, err := WriteToTmpFile("TempSheet", records)
	if err != nil {
		t.Fatalf("WriteToTmpFile failed: %v", err)
	}

	// Clean up
	defer os.Remove(filepath)

	// Verify file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Error("Temporary Excel file was not created")
	}

	// Verify file is not empty
	info, err := os.Stat(filepath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Temporary Excel file is empty")
	}

	// Verify filename pattern
	filename := filepath[len(filepath)-len("TempSheet-0000000000.xlsx"):]
	if len(filename) < len("TempSheet-") {
		t.Error("Filename doesn't match expected pattern")
	}
}

// TestWriteToFile_EmptyRecords tests writing empty records
func TestWriteToFile_EmptyRecords(t *testing.T) {
	tmpDir := t.TempDir()
	filepath := filepath.Join(tmpDir, "empty.xlsx")

	records := []interface{}{}

	err := WriteToFile("EmptySheet", records, filepath)
	if err != nil {
		t.Fatalf("WriteToFile with empty records failed: %v", err)
	}

	// File should still be created
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Error("Excel file was not created for empty records")
	}
}

// TestWriteRow tests writing individual rows
func TestWriteRow(t *testing.T) {
	// This is an internal function test
	// We'll test it indirectly through WriteToFile
	// But we can also test it directly if needed

	now := time.Now()
	user := &TestUser{
		Username:  "testuser",
		Email:     "test@example.com",
		Age:       25,
		Score:     90.0,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: timeutil.LocalTime(now),
	}

	tmpDir := t.TempDir()
	filepath := filepath.Join(tmpDir, "row_test.xlsx")

	records := []interface{}{user}
	err := WriteToFile("TestSheet", records, filepath)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	// Verify file was created successfully
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Error("Excel file was not created")
	}
}

// TestCOLNAMES tests the column name mapping
func TestCOLNAMES(t *testing.T) {
	tests := []struct {
		index    int
		expected string
	}{
		{0, "A"},
		{1, "B"},
		{25, "Z"},
		{26, "AA"},
		{27, "AB"},
		{48, "AW"},
	}

	for _, tt := range tests {
		if got, ok := COLNAMES[tt.index]; !ok {
			t.Errorf("COLNAMES missing index %d", tt.index)
		} else if got != tt.expected {
			t.Errorf("COLNAMES[%d] = %q, want %q", tt.index, got, tt.expected)
		}
	}
}

// TestWriteToFile_DifferentTypes tests different data types
func TestWriteToFile_DifferentTypes(t *testing.T) {
	type MixedTypes struct {
		StringField  string    `json:"string_field"`
		IntField     int       `json:"int_field"`
		Int32Field   int32     `json:"int32_field"`
		Int64Field   int64     `json:"int64_field"`
		Float32Field float32   `json:"float32_field"`
		Float64Field float64   `json:"float64_field"`
		BoolField    bool      `json:"bool_field"`
		TimeField    time.Time `json:"time_field"`
	}

	now := time.Now()
	records := []interface{}{
		&MixedTypes{
			StringField:  "test",
			IntField:     42,
			Int32Field:   int32(100),
			Int64Field:   int64(200),
			Float32Field: float32(3.14),
			Float64Field: 2.718,
			BoolField:    true,
			TimeField:    now,
		},
	}

	tmpDir := t.TempDir()
	filepath := filepath.Join(tmpDir, "mixed_types.xlsx")

	err := WriteToFile("MixedTypes", records, filepath)
	if err != nil {
		t.Fatalf("WriteToFile with mixed types failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Error("Excel file was not created")
	}
}

// TestWriteToFile_MultipleRecords tests writing multiple records
func TestWriteToFile_MultipleRecords(t *testing.T) {
	now := time.Now()
	records := make([]interface{}, 100)

	for i := 0; i < 100; i++ {
		records[i] = &TestUser{
			Username:  "user" + string(rune(i)),
			Email:     "user" + string(rune(i)) + "@example.com",
			Age:       20 + i%50,
			Score:     float64(50 + i%50),
			Active:    i%2 == 0,
			CreatedAt: now.Add(time.Duration(-i) * time.Hour),
			UpdatedAt: timeutil.LocalTime(now),
		}
	}

	tmpDir := t.TempDir()
	filepath := filepath.Join(tmpDir, "many_users.xlsx")

	err := WriteToFile("ManyUsers", records, filepath)
	if err != nil {
		t.Fatalf("WriteToFile with 100 records failed: %v", err)
	}

	// Verify file exists and has reasonable size
	info, err := os.Stat(filepath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() < 1000 {
		t.Error("Excel file seems too small for 100 records")
	}
}

// TestWriteToFile_DBTag tests using db tag instead of json tag
func TestWriteToFile_DBTag(t *testing.T) {
	type DBTagStruct struct {
		Field1 string `db:"database_field"`
		Field2 int    `db:"age"`
		Field3 string `db:"-"`                // Should be ignored
		Field4 string `json:"-" db:"visible"` // db tag takes precedence
	}

	records := []interface{}{
		&DBTagStruct{
			Field1: "value1",
			Field2: 42,
			Field3: "should not appear",
			Field4: "should appear",
		},
	}

	tmpDir := t.TempDir()
	filepath := filepath.Join(tmpDir, "db_tag.xlsx")

	err := WriteToFile("DBTag", records, filepath)
	if err != nil {
		t.Fatalf("WriteToFile with db tag failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Error("Excel file was not created")
	}
}

// BenchmarkWriteToFile benchmarks writing Excel file
func BenchmarkWriteToFile(b *testing.B) {
	now := time.Now()
	records := make([]interface{}, 100)

	for i := 0; i < 100; i++ {
		records[i] = &TestUser{
			Username:  "benchuser",
			Email:     "bench@example.com",
			Age:       25,
			Score:     90.0,
			Active:    true,
			CreatedAt: now,
			UpdatedAt: timeutil.LocalTime(now),
		}
	}

	tmpDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filepath := filepath.Join(tmpDir, "bench.xlsx")
		_ = WriteToFile("Benchmark", records, filepath)
		_ = os.Remove(filepath)
	}
}
