/**
 * 外部データソース用アプリフォームビルダー
 */

import type {
  CreateExternalAppRequest,
  CreateExternalFieldRequest,
} from "@/types/datasource";
import {
  mapDataTypeToFieldType,
  type ColumnInfo,
  type DataSource,
  type TableInfo,
} from "@/types/datasource";
import { FIELD_TYPE_LABELS, type FieldType } from "@/types/field";
import {
  Box,
  Button,
  Card,
  CardBody,
  Checkbox,
  FormControl,
  FormErrorMessage,
  FormLabel,
  Heading,
  HStack,
  Input,
  Select,
  Table,
  Tbody,
  Td,
  Text,
  Th,
  Thead,
  Tr,
  useToast,
  VStack,
} from "@chakra-ui/react";
import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

interface ExternalAppFormBuilderProps {
  dataSource: DataSource;
  table: TableInfo;
  columns: ColumnInfo[];
}

interface FieldMapping {
  source_column_name: string;
  field_code: string;
  field_name: string;
  field_type: FieldType;
  required: boolean;
  selected: boolean;
}

export function ExternalAppFormBuilder({
  dataSource,
  table,
  columns,
}: ExternalAppFormBuilderProps) {
  const toast = useToast();
  const navigate = useNavigate();

  const [appName, setAppName] = useState(table.name);
  const [appDescription, setAppDescription] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  // カラムからフィールドマッピングを初期化
  const [fieldMappings, setFieldMappings] = useState<FieldMapping[]>(() =>
    columns.map((col, index) => ({
      source_column_name: col.name,
      field_code: col.name.toLowerCase().replace(/[^a-z0-9_]/g, "_"),
      field_name: col.name,
      field_type: mapDataTypeToFieldType(col.data_type) as FieldType,
      required: !col.is_nullable,
      selected: true,
      display_order: index,
    }))
  );

  const createMutation = useMutation({
    mutationFn: (data: CreateExternalAppRequest) => {
      // appsApiにexternal用のメソッドを追加する必要がありますが、
      // 既存のcreateメソッドを拡張するか、新しいエンドポイントを使用
      // ここでは簡略化のためにfetchを直接使用
      const token = localStorage.getItem("token");
      return fetch(`${import.meta.env.VITE_API_URL}/apps/external`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(data),
      }).then((res) => {
        if (!res.ok) throw new Error("アプリの作成に失敗しました");
        return res.json();
      });
    },
    onSuccess: () => {
      toast({
        title: "アプリを作成しました",
        status: "success",
        duration: 3000,
      });
      navigate("/apps");
    },
    onError: (error: Error) => {
      toast({
        title: "作成に失敗しました",
        description: error.message,
        status: "error",
        duration: 5000,
      });
    },
  });

  const handleFieldMappingChange = (
    index: number,
    field: keyof FieldMapping,
    value: string | boolean
  ) => {
    setFieldMappings((prev) => {
      const newMappings = [...prev];
      newMappings[index] = { ...newMappings[index], [field]: value };
      return newMappings;
    });
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!appName.trim()) {
      newErrors.appName = "アプリ名は必須です";
    }

    const selectedFields = fieldMappings.filter((f) => f.selected);
    if (selectedFields.length === 0) {
      newErrors.fields = "少なくとも1つのフィールドを選択してください";
    }

    // フィールドコードの重複チェック
    const codes = selectedFields.map((f) => f.field_code);
    const duplicates = codes.filter((code, i) => codes.indexOf(code) !== i);
    if (duplicates.length > 0) {
      newErrors.fields = `フィールドコードが重複しています: ${duplicates.join(", ")}`;
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = () => {
    if (!validate()) return;

    const selectedFields = fieldMappings
      .filter((f) => f.selected)
      .map(
        (f, index): CreateExternalFieldRequest => ({
          source_column_name: f.source_column_name,
          field_code: f.field_code,
          field_name: f.field_name,
          field_type: f.field_type,
          required: f.required,
          display_order: index,
        })
      );

    const request: CreateExternalAppRequest = {
      name: appName,
      description: appDescription,
      data_source_id: dataSource.id,
      source_table_name: table.name,
      fields: selectedFields,
    };

    createMutation.mutate(request);
  };

  const selectedCount = fieldMappings.filter((f) => f.selected).length;

  return (
    <VStack spacing={6} align="stretch">
      {/* 基本情報 */}
      <Card>
        <CardBody>
          <Heading size="md" mb={4}>
            アプリ基本情報
          </Heading>
          <VStack spacing={4}>
            <FormControl isInvalid={!!errors.appName}>
              <FormLabel>アプリ名</FormLabel>
              <Input
                value={appName}
                onChange={(e) => setAppName(e.target.value)}
                placeholder="アプリ名を入力"
              />
              <FormErrorMessage>{errors.appName}</FormErrorMessage>
            </FormControl>

            <FormControl>
              <FormLabel>説明</FormLabel>
              <Input
                value={appDescription}
                onChange={(e) => setAppDescription(e.target.value)}
                placeholder="アプリの説明（任意）"
              />
            </FormControl>

            <Box w="100%">
              <Text fontSize="sm" color="gray.500">
                データソース: {dataSource.name}
              </Text>
              <Text fontSize="sm" color="gray.500">
                テーブル: {table.schema ? `${table.schema}.` : ""}
                {table.name}
              </Text>
            </Box>
          </VStack>
        </CardBody>
      </Card>

      {/* フィールドマッピング */}
      <Card>
        <CardBody>
          <HStack justify="space-between" mb={4}>
            <Heading size="md">フィールド設定</Heading>
            <Text fontSize="sm" color="gray.500">
              {selectedCount} / {fieldMappings.length} フィールド選択中
            </Text>
          </HStack>

          {errors.fields && (
            <Text color="red.500" mb={4}>
              {errors.fields}
            </Text>
          )}

          <Box overflowX="auto">
            <Table size="sm">
              <Thead>
                <Tr>
                  <Th width="50px">選択</Th>
                  <Th>元カラム名</Th>
                  <Th>フィールドコード</Th>
                  <Th>表示名</Th>
                  <Th>フィールドタイプ</Th>
                  <Th width="80px">必須</Th>
                </Tr>
              </Thead>
              <Tbody>
                {fieldMappings.map((mapping, index) => (
                  <Tr
                    key={mapping.source_column_name}
                    opacity={mapping.selected ? 1 : 0.5}
                  >
                    <Td>
                      <Checkbox
                        isChecked={mapping.selected}
                        onChange={(e) =>
                          handleFieldMappingChange(
                            index,
                            "selected",
                            e.target.checked
                          )
                        }
                      />
                    </Td>
                    <Td>
                      <Text fontFamily="mono" fontSize="sm">
                        {mapping.source_column_name}
                      </Text>
                    </Td>
                    <Td>
                      <Input
                        size="sm"
                        value={mapping.field_code}
                        onChange={(e) =>
                          handleFieldMappingChange(
                            index,
                            "field_code",
                            e.target.value
                          )
                        }
                        isDisabled={!mapping.selected}
                      />
                    </Td>
                    <Td>
                      <Input
                        size="sm"
                        value={mapping.field_name}
                        onChange={(e) =>
                          handleFieldMappingChange(
                            index,
                            "field_name",
                            e.target.value
                          )
                        }
                        isDisabled={!mapping.selected}
                      />
                    </Td>
                    <Td>
                      <Select
                        size="sm"
                        value={mapping.field_type}
                        onChange={(e) =>
                          handleFieldMappingChange(
                            index,
                            "field_type",
                            e.target.value
                          )
                        }
                        isDisabled={!mapping.selected}
                      >
                        {Object.entries(FIELD_TYPE_LABELS).map(
                          ([type, label]) => (
                            <option key={type} value={type}>
                              {label}
                            </option>
                          )
                        )}
                      </Select>
                    </Td>
                    <Td>
                      <Checkbox
                        isChecked={mapping.required}
                        onChange={(e) =>
                          handleFieldMappingChange(
                            index,
                            "required",
                            e.target.checked
                          )
                        }
                        isDisabled={!mapping.selected}
                      />
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          </Box>
        </CardBody>
      </Card>

      {/* 注意事項 */}
      <Card bg="orange.50">
        <CardBody>
          <Text fontWeight="bold" color="orange.700" mb={2}>
            ⚠️ 読み取り専用
          </Text>
          <Text fontSize="sm" color="orange.600">
            外部データソースから作成されたアプリは読み取り専用です。
            レコードの追加・編集・削除はできません。
          </Text>
        </CardBody>
      </Card>

      {/* 送信ボタン */}
      <HStack justify="flex-end">
        <Button
          colorScheme="blue"
          size="lg"
          onClick={handleSubmit}
          isLoading={createMutation.isPending}
        >
          アプリを作成
        </Button>
      </HStack>
    </VStack>
  );
}
