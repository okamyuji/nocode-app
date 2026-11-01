/**
 * データソース選択コンポーネント
 * 外部アプリ作成時にデータソースとテーブルを選択するためのUI
 */

import { useDataSourcesApi } from "@/api";
import {
  DB_TYPE_LABELS,
  type ColumnInfo,
  type DataSource,
  type TableInfo,
} from "@/types/datasource";
import {
  Alert,
  Badge,
  Box,
  Button,
  Card,
  Field,
  HStack,
  NativeSelect,
  Skeleton,
  Text,
  VStack,
} from "@chakra-ui/react";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { FiCheck, FiDatabase, FiRefreshCw } from "react-icons/fi";

interface DataSourceSelectorProps {
  onSelect: (
    dataSource: DataSource,
    table: TableInfo,
    columns: ColumnInfo[]
  ) => void;
  selectedDataSourceId?: number;
  selectedTableName?: string;
}

export function DataSourceSelector({
  onSelect,
  selectedDataSourceId,
  selectedTableName,
}: DataSourceSelectorProps) {
  const dataSourcesApi = useDataSourcesApi();
  const [dataSourceId, setDataSourceId] = useState<number | null>(
    selectedDataSourceId || null
  );
  const [tableName, setTableName] = useState<string | null>(
    selectedTableName || null
  );

  // データソース一覧を取得
  const { data: dataSourcesData, isLoading: isLoadingDataSources } = useQuery({
    queryKey: ["dataSources"],
    queryFn: () => dataSourcesApi.getDataSources(1, 100),
  });

  // 選択されたデータソースのテーブル一覧を取得
  const {
    data: tablesData,
    isLoading: isLoadingTables,
    refetch: refetchTables,
  } = useQuery({
    queryKey: ["dataSourceTables", dataSourceId],
    queryFn: () =>
      dataSourceId ? dataSourcesApi.getTables(dataSourceId) : null,
    enabled: !!dataSourceId,
  });

  // 選択されたテーブルのカラム一覧を取得
  const { data: columnsData, isLoading: isLoadingColumns } = useQuery({
    queryKey: ["dataSourceColumns", dataSourceId, tableName],
    queryFn: () =>
      dataSourceId && tableName
        ? dataSourcesApi.getColumns(dataSourceId, tableName)
        : null,
    enabled: !!dataSourceId && !!tableName,
  });

  const dataSources = dataSourcesData?.data_sources || [];
  const tables = tablesData?.tables || [];
  const columns = columnsData?.columns || [];

  const selectedDataSource = dataSources.find((ds) => ds.id === dataSourceId);
  const selectedTable = tables.find((t) => t.name === tableName);

  // データソースが変更されたらテーブル選択をリセット
  useEffect(() => {
    if (!selectedDataSourceId) {
      setTableName(null);
    }
  }, [dataSourceId, selectedDataSourceId]);

  const handleConfirm = () => {
    if (selectedDataSource && selectedTable && columns.length > 0) {
      onSelect(selectedDataSource, selectedTable, columns);
    }
  };

  if (isLoadingDataSources) {
    return <Skeleton height="200px" borderRadius="md" />;
  }

  if (dataSources.length === 0) {
    return (
      <Alert.Root status="warning" borderRadius="md">
        <Alert.Indicator />
        <Alert.Title>
          データソースが登録されていません。設定画面からデータソースを追加してください。
        </Alert.Title>
      </Alert.Root>
    );
  }

  return (
    <VStack gap={4} align="stretch">
      {/* データソース選択 */}
      <Field.Root>
        <Field.Label>データソース</Field.Label>
        <NativeSelect.Root>
          <NativeSelect.Field
            placeholder="データソースを選択"
            value={dataSourceId || ""}
            onChange={(e) => setDataSourceId(Number(e.target.value) || null)}
          >
            {dataSources.map((ds) => (
              <option key={ds.id} value={ds.id}>
                {ds.name} ({DB_TYPE_LABELS[ds.db_type]})
              </option>
            ))}
          </NativeSelect.Field>
          <NativeSelect.Indicator />
        </NativeSelect.Root>
      </Field.Root>

      {/* 選択されたデータソースの情報 */}
      {selectedDataSource && (
        <Card.Root variant="outline">
          <Card.Body>
            <HStack justify="space-between">
              <HStack>
                <FiDatabase />
                <Text fontWeight="medium">{selectedDataSource.name}</Text>
                <Badge colorPalette="blue">
                  {DB_TYPE_LABELS[selectedDataSource.db_type]}
                </Badge>
              </HStack>
              <Text fontSize="sm" color="gray.500">
                {selectedDataSource.host}:{selectedDataSource.port}/
                {selectedDataSource.database_name}
              </Text>
            </HStack>
          </Card.Body>
        </Card.Root>
      )}

      {/* テーブル/ビュー選択 */}
      {dataSourceId && (
        <Field.Root>
          <HStack justify="space-between" mb={2}>
            <Field.Label mb={0}>テーブル / ビュー</Field.Label>
            <Button
              size="xs"
              variant="ghost"
              onClick={() => refetchTables()}
              loading={isLoadingTables}
            >
              <FiRefreshCw />
              更新
            </Button>
          </HStack>
          {isLoadingTables ? (
            <Skeleton height="40px" />
          ) : (
            <NativeSelect.Root>
              <NativeSelect.Field
                placeholder="テーブル/ビューを選択"
                value={tableName || ""}
                onChange={(e) => setTableName(e.target.value || null)}
              >
                {tables.map((table) => (
                  <option key={table.name} value={table.name}>
                    [{table.type}]{" "}
                    {table.schema
                      ? `${table.schema}.${table.name}`
                      : table.name}
                  </option>
                ))}
              </NativeSelect.Field>
              <NativeSelect.Indicator />
            </NativeSelect.Root>
          )}
        </Field.Root>
      )}

      {/* カラムプレビュー */}
      {tableName && (
        <Box>
          <Text fontWeight="medium" mb={2}>
            カラム一覧
          </Text>
          {isLoadingColumns ? (
            <Skeleton height="100px" />
          ) : columns.length > 0 ? (
            <Card.Root variant="outline">
              <Card.Body>
                <VStack align="stretch" gap={1}>
                  {columns.map((col) => (
                    <HStack
                      key={col.name}
                      justify="space-between"
                      fontSize="sm"
                      py={1}
                      borderBottomWidth="1px"
                      _last={{ borderBottom: "none" }}
                    >
                      <HStack>
                        <Text fontWeight="medium">{col.name}</Text>
                        {col.is_primary_key && (
                          <Badge colorPalette="yellow" size="sm">
                            PK
                          </Badge>
                        )}
                        {!col.is_nullable && (
                          <Badge colorPalette="red" size="sm">
                            NOT NULL
                          </Badge>
                        )}
                      </HStack>
                      <Text color="gray.500">{col.data_type}</Text>
                    </HStack>
                  ))}
                </VStack>
              </Card.Body>
            </Card.Root>
          ) : (
            <Alert.Root status="info" borderRadius="md">
              <Alert.Indicator />
              <Alert.Title>カラム情報を取得できませんでした</Alert.Title>
            </Alert.Root>
          )}
        </Box>
      )}

      {/* 確定ボタン */}
      {selectedDataSource && selectedTable && columns.length > 0 && (
        <Button colorPalette="blue" onClick={handleConfirm}>
          <FiCheck />
          このテーブルを使用
        </Button>
      )}
    </VStack>
  );
}
