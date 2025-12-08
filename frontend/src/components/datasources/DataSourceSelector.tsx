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
  AlertIcon,
  Badge,
  Box,
  Button,
  Card,
  CardBody,
  FormControl,
  FormLabel,
  HStack,
  Select,
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
      <Alert status="warning" borderRadius="md">
        <AlertIcon />
        データソースが登録されていません。設定画面からデータソースを追加してください。
      </Alert>
    );
  }

  return (
    <VStack spacing={4} align="stretch">
      {/* データソース選択 */}
      <FormControl>
        <FormLabel>データソース</FormLabel>
        <Select
          placeholder="データソースを選択"
          value={dataSourceId || ""}
          onChange={(e) => setDataSourceId(Number(e.target.value) || null)}
        >
          {dataSources.map((ds) => (
            <option key={ds.id} value={ds.id}>
              {ds.name} ({DB_TYPE_LABELS[ds.db_type]})
            </option>
          ))}
        </Select>
      </FormControl>

      {/* 選択されたデータソースの情報 */}
      {selectedDataSource && (
        <Card variant="outline">
          <CardBody>
            <HStack justify="space-between">
              <HStack>
                <FiDatabase />
                <Text fontWeight="medium">{selectedDataSource.name}</Text>
                <Badge colorScheme="blue">
                  {DB_TYPE_LABELS[selectedDataSource.db_type]}
                </Badge>
              </HStack>
              <Text fontSize="sm" color="gray.500">
                {selectedDataSource.host}:{selectedDataSource.port}/
                {selectedDataSource.database_name}
              </Text>
            </HStack>
          </CardBody>
        </Card>
      )}

      {/* テーブル/ビュー選択 */}
      {dataSourceId && (
        <FormControl>
          <HStack justify="space-between" mb={2}>
            <FormLabel mb={0}>テーブル / ビュー</FormLabel>
            <Button
              size="xs"
              variant="ghost"
              leftIcon={<FiRefreshCw />}
              onClick={() => refetchTables()}
              isLoading={isLoadingTables}
            >
              更新
            </Button>
          </HStack>
          {isLoadingTables ? (
            <Skeleton height="40px" />
          ) : (
            <Select
              placeholder="テーブル/ビューを選択"
              value={tableName || ""}
              onChange={(e) => setTableName(e.target.value || null)}
            >
              {tables.map((table) => (
                <option key={table.name} value={table.name}>
                  [{table.type}]{" "}
                  {table.schema ? `${table.schema}.${table.name}` : table.name}
                </option>
              ))}
            </Select>
          )}
        </FormControl>
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
            <Card variant="outline">
              <CardBody>
                <VStack align="stretch" spacing={1}>
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
                          <Badge colorScheme="yellow" size="sm">
                            PK
                          </Badge>
                        )}
                        {!col.is_nullable && (
                          <Badge colorScheme="red" size="sm">
                            NOT NULL
                          </Badge>
                        )}
                      </HStack>
                      <Text color="gray.500">{col.data_type}</Text>
                    </HStack>
                  ))}
                </VStack>
              </CardBody>
            </Card>
          ) : (
            <Alert status="info" borderRadius="md">
              <AlertIcon />
              カラム情報を取得できませんでした
            </Alert>
          )}
        </Box>
      )}

      {/* 確定ボタン */}
      {selectedDataSource && selectedTable && columns.length > 0 && (
        <Button
          colorScheme="blue"
          leftIcon={<FiCheck />}
          onClick={handleConfirm}
        >
          このテーブルを使用
        </Button>
      )}
    </VStack>
  );
}
