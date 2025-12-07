package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsValidDBType(t *testing.T) {
	tests := []struct {
		name     string
		dbType   string
		expected bool
	}{
		{"postgresql is valid", "postgresql", true},
		{"mysql is valid", "mysql", true},
		{"oracle is valid", "oracle", true},
		{"sqlserver is valid", "sqlserver", true},
		{"invalid type", "mongodb", false},
		{"empty string", "", false},
		{"uppercase is invalid", "POSTGRESQL", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidDBType(tt.dbType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDefaultPort(t *testing.T) {
	tests := []struct {
		name     string
		dbType   DBType
		expected int
	}{
		{"postgresql default port", DBTypePostgreSQL, 5432},
		{"mysql default port", DBTypeMySQL, 3306},
		{"oracle default port", DBTypeOracle, 1521},
		{"sqlserver default port", DBTypeSQLServer, 1433},
		{"unknown type", DBType("unknown"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDefaultPort(tt.dbType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataSource_ToResponse(t *testing.T) {
	now := time.Now()
	ds := &DataSource{
		ID:                1,
		Name:              "test-ds",
		DBType:            DBTypePostgreSQL,
		Host:              "localhost",
		Port:              5432,
		DatabaseName:      "testdb",
		Username:          "testuser",
		EncryptedPassword: "encrypted-password",
		CreatedBy:         100,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	resp := ds.ToResponse()

	assert.Equal(t, uint64(1), resp.ID)
	assert.Equal(t, "test-ds", resp.Name)
	assert.Equal(t, DBTypePostgreSQL, resp.DBType)
	assert.Equal(t, "localhost", resp.Host)
	assert.Equal(t, 5432, resp.Port)
	assert.Equal(t, "testdb", resp.DatabaseName)
	assert.Equal(t, "testuser", resp.Username)
	assert.Equal(t, uint64(100), resp.CreatedBy)
	assert.Equal(t, now, resp.CreatedAt)
	assert.Equal(t, now, resp.UpdatedAt)
}

func TestValidDBTypes(t *testing.T) {
	// ValidDBTypesが正しい値を含んでいることを確認
	assert.Len(t, ValidDBTypes, 4)
	assert.Contains(t, ValidDBTypes, DBTypePostgreSQL)
	assert.Contains(t, ValidDBTypes, DBTypeMySQL)
	assert.Contains(t, ValidDBTypes, DBTypeOracle)
	assert.Contains(t, ValidDBTypes, DBTypeSQLServer)
}
