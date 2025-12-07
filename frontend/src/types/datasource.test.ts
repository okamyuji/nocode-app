/**
 * データソース型定義のテスト
 */

import { describe, expect, it } from "vitest";
import {
  DB_TYPE_LABELS,
  DEFAULT_PORTS,
  mapDataTypeToFieldType,
} from "./datasource";

describe("datasource types", () => {
  describe("DB_TYPE_LABELS", () => {
    it("should have labels for all supported database types", () => {
      expect(DB_TYPE_LABELS.postgresql).toBe("PostgreSQL");
      expect(DB_TYPE_LABELS.mysql).toBe("MySQL");
      expect(DB_TYPE_LABELS.oracle).toBe("Oracle");
      expect(DB_TYPE_LABELS.sqlserver).toBe("SQL Server");
    });

    it("should have exactly 4 database types", () => {
      expect(Object.keys(DB_TYPE_LABELS)).toHaveLength(4);
    });
  });

  describe("DEFAULT_PORTS", () => {
    it("should have correct default ports for each database type", () => {
      expect(DEFAULT_PORTS.postgresql).toBe(5432);
      expect(DEFAULT_PORTS.mysql).toBe(3306);
      expect(DEFAULT_PORTS.oracle).toBe(1521);
      expect(DEFAULT_PORTS.sqlserver).toBe(1433);
    });

    it("should have exactly 4 database types", () => {
      expect(Object.keys(DEFAULT_PORTS)).toHaveLength(4);
    });
  });

  describe("mapDataTypeToFieldType", () => {
    describe("number types", () => {
      it("should map integer types to number", () => {
        expect(mapDataTypeToFieldType("int")).toBe("number");
        expect(mapDataTypeToFieldType("INT")).toBe("number");
        expect(mapDataTypeToFieldType("bigint")).toBe("number");
        expect(mapDataTypeToFieldType("smallint")).toBe("number");
        expect(mapDataTypeToFieldType("tinyint")).toBe("number");
      });

      it("should map decimal types to number", () => {
        expect(mapDataTypeToFieldType("decimal")).toBe("number");
        expect(mapDataTypeToFieldType("decimal(10,2)")).toBe("number");
        expect(mapDataTypeToFieldType("numeric")).toBe("number");
        expect(mapDataTypeToFieldType("numeric(18,4)")).toBe("number");
      });

      it("should map float types to number", () => {
        expect(mapDataTypeToFieldType("float")).toBe("number");
        expect(mapDataTypeToFieldType("double")).toBe("number");
        expect(mapDataTypeToFieldType("real")).toBe("number");
      });

      it("should map Oracle number type to number", () => {
        expect(mapDataTypeToFieldType("number")).toBe("number");
        expect(mapDataTypeToFieldType("NUMBER(10)")).toBe("number");
      });
    });

    describe("datetime types", () => {
      it("should map datetime types to datetime", () => {
        expect(mapDataTypeToFieldType("datetime")).toBe("datetime");
        expect(mapDataTypeToFieldType("datetime2")).toBe("datetime");
        expect(mapDataTypeToFieldType("timestamp")).toBe("datetime");
        expect(mapDataTypeToFieldType("timestamp with time zone")).toBe(
          "datetime"
        );
      });

      it("should map date type to date", () => {
        expect(mapDataTypeToFieldType("date")).toBe("date");
        expect(mapDataTypeToFieldType("DATE")).toBe("date");
      });
    });

    describe("text types", () => {
      it("should map long text types to textarea", () => {
        expect(mapDataTypeToFieldType("text")).toBe("textarea");
        expect(mapDataTypeToFieldType("TEXT")).toBe("textarea");
        expect(mapDataTypeToFieldType("clob")).toBe("textarea");
        expect(mapDataTypeToFieldType("longtext")).toBe("textarea");
        expect(mapDataTypeToFieldType("mediumtext")).toBe("textarea");
        expect(mapDataTypeToFieldType("ntext")).toBe("textarea");
      });

      it("should map varchar types to text", () => {
        expect(mapDataTypeToFieldType("varchar")).toBe("text");
        expect(mapDataTypeToFieldType("varchar(255)")).toBe("text");
        expect(mapDataTypeToFieldType("nvarchar")).toBe("text");
        expect(mapDataTypeToFieldType("char")).toBe("text");
      });
    });

    describe("boolean types", () => {
      it("should map boolean types to checkbox", () => {
        expect(mapDataTypeToFieldType("bool")).toBe("checkbox");
        expect(mapDataTypeToFieldType("boolean")).toBe("checkbox");
        expect(mapDataTypeToFieldType("bit")).toBe("checkbox");
      });
    });

    describe("json types", () => {
      it("should map json types to textarea", () => {
        expect(mapDataTypeToFieldType("json")).toBe("textarea");
        expect(mapDataTypeToFieldType("jsonb")).toBe("textarea");
      });
    });

    describe("default fallback", () => {
      it("should return text for unknown types", () => {
        expect(mapDataTypeToFieldType("unknown")).toBe("text");
        expect(mapDataTypeToFieldType("blob")).toBe("text");
        expect(mapDataTypeToFieldType("binary")).toBe("text");
        expect(mapDataTypeToFieldType("image")).toBe("text");
      });
    });
  });
});
