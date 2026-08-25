/**
 * データソースAPIのテスト
 */

import type { AxiosInstance } from "axios";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type {
  ColumnListResponse,
  CreateDataSourceRequest,
  DataSource,
  DataSourceListResponse,
  TableListResponse,
  TestConnectionResponse,
  UpdateDataSourceRequest,
} from "../types/datasource";
import { createDataSourceApi } from "./datasources";

describe("datasources API", () => {
  let mockClient: {
    get: ReturnType<typeof vi.fn>;
    post: ReturnType<typeof vi.fn>;
    put: ReturnType<typeof vi.fn>;
    delete: ReturnType<typeof vi.fn>;
  };
  let api: ReturnType<typeof createDataSourceApi>;

  beforeEach(() => {
    mockClient = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
    };
    api = createDataSourceApi(mockClient as unknown as AxiosInstance);
  });

  describe("getDataSources", () => {
    it("should fetch data sources with default pagination", async () => {
      const mockResponse: DataSourceListResponse = {
        data_sources: [
          {
            id: 1,
            name: "Test DS",
            db_type: "postgresql",
            host: "localhost",
            port: 5432,
            database_name: "testdb",
            username: "user",
            created_by: 1,
            created_at: "2024-01-01T00:00:00Z",
            updated_at: "2024-01-01T00:00:00Z",
          },
        ],
        pagination: { page: 1, limit: 20, total: 1, total_pages: 1 },
      };

      mockClient.get.mockResolvedValue({ data: mockResponse });

      const result = await api.getDataSources();

      expect(mockClient.get).toHaveBeenCalledWith("/datasources", {
        params: { page: 1, limit: 20 },
      });
      expect(result).toEqual(mockResponse);
    });

    it("should fetch data sources with custom pagination", async () => {
      const mockResponse: DataSourceListResponse = {
        data_sources: [],
        pagination: { page: 2, limit: 10, total: 15, total_pages: 2 },
      };

      mockClient.get.mockResolvedValue({ data: mockResponse });

      const result = await api.getDataSources(2, 10);

      expect(mockClient.get).toHaveBeenCalledWith("/datasources", {
        params: { page: 2, limit: 10 },
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe("getDataSource", () => {
    it("should fetch a single data source by id", async () => {
      const mockDataSource: DataSource = {
        id: 1,
        name: "Test DS",
        db_type: "mysql",
        host: "localhost",
        port: 3306,
        database_name: "testdb",
        username: "user",
        created_by: 1,
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      mockClient.get.mockResolvedValue({ data: mockDataSource });

      const result = await api.getDataSource(1);

      expect(mockClient.get).toHaveBeenCalledWith("/datasources/1");
      expect(result).toEqual(mockDataSource);
    });
  });

  describe("createDataSource", () => {
    it("should create a new data source", async () => {
      const createRequest: CreateDataSourceRequest = {
        name: "New DS",
        db_type: "postgresql",
        host: "localhost",
        port: 5432,
        database_name: "newdb",
        username: "user",
        password: "password",
      };

      const mockResponse: DataSource = {
        id: 2,
        name: "New DS",
        db_type: "postgresql",
        host: "localhost",
        port: 5432,
        database_name: "newdb",
        username: "user",
        created_by: 1,
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      mockClient.post.mockResolvedValue({ data: mockResponse });

      const result = await api.createDataSource(createRequest);

      expect(mockClient.post).toHaveBeenCalledWith(
        "/datasources",
        createRequest
      );
      expect(result).toEqual(mockResponse);
    });
  });

  describe("updateDataSource", () => {
    it("should update an existing data source", async () => {
      const updateRequest: UpdateDataSourceRequest = {
        name: "Updated DS",
        host: "newhost",
      };

      const mockResponse: DataSource = {
        id: 1,
        name: "Updated DS",
        db_type: "postgresql",
        host: "newhost",
        port: 5432,
        database_name: "testdb",
        username: "user",
        created_by: 1,
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-02T00:00:00Z",
      };

      mockClient.put.mockResolvedValue({ data: mockResponse });

      const result = await api.updateDataSource(1, updateRequest);

      expect(mockClient.put).toHaveBeenCalledWith(
        "/datasources/1",
        updateRequest
      );
      expect(result).toEqual(mockResponse);
    });
  });

  describe("deleteDataSource", () => {
    it("should delete a data source", async () => {
      mockClient.delete.mockResolvedValue({});

      await api.deleteDataSource(1);

      expect(mockClient.delete).toHaveBeenCalledWith("/datasources/1");
    });
  });

  describe("testConnection", () => {
    it("should test connection successfully", async () => {
      const testRequest = {
        db_type: "postgresql" as const,
        host: "localhost",
        port: 5432,
        database_name: "testdb",
        username: "user",
        password: "password",
      };

      const mockResponse: TestConnectionResponse = {
        success: true,
        message: "接続に成功しました",
      };

      mockClient.post.mockResolvedValue({ data: mockResponse });

      const result = await api.testConnection(testRequest);

      expect(mockClient.post).toHaveBeenCalledWith(
        "/datasources/test",
        testRequest
      );
      expect(result).toEqual(mockResponse);
    });

    it("should return failure response for failed connection", async () => {
      const testRequest = {
        db_type: "mysql" as const,
        host: "invalid-host",
        port: 3306,
        database_name: "testdb",
        username: "user",
        password: "wrong",
      };

      const mockResponse: TestConnectionResponse = {
        success: false,
        message: "Connection refused",
      };

      mockClient.post.mockResolvedValue({ data: mockResponse });

      const result = await api.testConnection(testRequest);

      expect(result.success).toBe(false);
      expect(result.message).toBe("Connection refused");
    });
  });

  describe("getTables", () => {
    it("should fetch tables for a data source", async () => {
      const mockResponse: TableListResponse = {
        tables: [
          { name: "users", schema: "public", type: "TABLE" },
          { name: "orders", schema: "public", type: "VIEW" },
        ],
      };

      mockClient.get.mockResolvedValue({ data: mockResponse });

      const result = await api.getTables(1);

      expect(mockClient.get).toHaveBeenCalledWith("/datasources/1/tables");
      expect(result).toEqual(mockResponse);
      expect(result.tables).toHaveLength(2);
    });
  });

  describe("getColumns", () => {
    it("should fetch columns for a table", async () => {
      const mockResponse: ColumnListResponse = {
        columns: [
          {
            name: "id",
            data_type: "integer",
            is_nullable: false,
            is_primary_key: true,
          },
          {
            name: "name",
            data_type: "varchar(255)",
            is_nullable: true,
            is_primary_key: false,
          },
        ],
      };

      mockClient.get.mockResolvedValue({ data: mockResponse });

      const result = await api.getColumns(1, "users");

      expect(mockClient.get).toHaveBeenCalledWith(
        "/datasources/1/tables/users/columns"
      );
      expect(result).toEqual(mockResponse);
      expect(result.columns).toHaveLength(2);
    });
  });
});
