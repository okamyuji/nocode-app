/**
 * データソース作成・編集フォームコンポーネント
 */

import { useDataSourcesApi } from "@/api";
import {
  DB_TYPE_LABELS,
  DEFAULT_PORTS,
  type CreateDataSourceRequest,
  type DataSource,
  type DBType,
} from "@/types/datasource";
import {
  Alert,
  Button,
  CloseButton,
  Dialog,
  Field,
  HStack,
  Input,
  NativeSelect,
  Portal,
  VStack,
} from "@chakra-ui/react";
import { toaster } from "@/components/ui/toaster";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";

interface DataSourceFormProps {
  isOpen: boolean;
  onClose: () => void;
  dataSource?: DataSource | null;
}

const DB_TYPES: DBType[] = ["postgresql", "mysql", "oracle", "sqlserver"];

export function DataSourceForm({
  isOpen,
  onClose,
  dataSource,
}: DataSourceFormProps) {
  const queryClient = useQueryClient();
  const dataSourcesApi = useDataSourcesApi();
  const isEdit = !!dataSource;

  const [formData, setFormData] = useState<CreateDataSourceRequest>({
    name: "",
    db_type: "postgresql",
    host: "",
    port: 5432,
    database_name: "",
    username: "",
    password: "",
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [testResult, setTestResult] = useState<{
    success: boolean;
    message: string;
  } | null>(null);

  useEffect(() => {
    if (dataSource) {
      setFormData({
        name: dataSource.name,
        db_type: dataSource.db_type,
        host: dataSource.host,
        port: dataSource.port,
        database_name: dataSource.database_name,
        username: dataSource.username,
        password: "", // パスワードは表示しない
      });
    } else {
      setFormData({
        name: "",
        db_type: "postgresql",
        host: "",
        port: 5432,
        database_name: "",
        username: "",
        password: "",
      });
    }
    setErrors({});
    setTestResult(null);
  }, [dataSource, isOpen]);

  const handleDBTypeChange = (dbType: DBType) => {
    setFormData((prev) => ({
      ...prev,
      db_type: dbType,
      port: DEFAULT_PORTS[dbType],
    }));
  };

  const testMutation = useMutation({
    mutationFn: () =>
      dataSourcesApi.testConnection({
        db_type: formData.db_type,
        host: formData.host,
        port: formData.port,
        database_name: formData.database_name,
        username: formData.username,
        password: formData.password,
      }),
    onSuccess: (data) => {
      setTestResult(data);
    },
    onError: () => {
      setTestResult({
        success: false,
        message: "接続テストに失敗しました",
      });
    },
  });

  const createMutation = useMutation({
    mutationFn: () => dataSourcesApi.createDataSource(formData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dataSources"] });
      toaster.create({
        title: "データソースを作成しました",
        type: "success",
        duration: 3000,
      });
      onClose();
    },
    onError: (error: Error) => {
      toaster.create({
        title: "作成に失敗しました",
        description: error.message,
        type: "error",
        duration: 5000,
      });
    },
  });

  const updateMutation = useMutation({
    mutationFn: () =>
      dataSourcesApi.updateDataSource(dataSource!.id, {
        name: formData.name !== dataSource!.name ? formData.name : undefined,
        host: formData.host !== dataSource!.host ? formData.host : undefined,
        port: formData.port !== dataSource!.port ? formData.port : undefined,
        database_name:
          formData.database_name !== dataSource!.database_name
            ? formData.database_name
            : undefined,
        username:
          formData.username !== dataSource!.username
            ? formData.username
            : undefined,
        password: formData.password || undefined,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dataSources"] });
      toaster.create({
        title: "データソースを更新しました",
        type: "success",
        duration: 3000,
      });
      onClose();
    },
    onError: (error: Error) => {
      toaster.create({
        title: "更新に失敗しました",
        description: error.message,
        type: "error",
        duration: 5000,
      });
    },
  });

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = "名前は必須です";
    }
    if (!formData.host.trim()) {
      newErrors.host = "ホストは必須です";
    }
    if (!formData.port || formData.port < 1 || formData.port > 65535) {
      newErrors.port = "有効なポート番号を入力してください";
    }
    if (!formData.database_name.trim()) {
      newErrors.database_name = "データベース名は必須です";
    }
    if (!formData.username.trim()) {
      newErrors.username = "ユーザー名は必須です";
    }
    if (!isEdit && !formData.password.trim()) {
      newErrors.password = "パスワードは必須です";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = () => {
    if (!validate()) return;

    if (isEdit) {
      updateMutation.mutate();
    } else {
      createMutation.mutate();
    }
  };

  const handleTestConnection = () => {
    if (!validate()) return;
    testMutation.mutate();
  };

  const isLoading =
    createMutation.isPending ||
    updateMutation.isPending ||
    testMutation.isPending;

  return (
    <Dialog.Root
      open={isOpen}
      onOpenChange={(e) => !e.open && onClose()}
      size="lg"
    >
      <Portal>
        <Dialog.Backdrop />
        <Dialog.Positioner>
          <Dialog.Content>
            <Dialog.Header>
              <Dialog.Title>
                {isEdit ? "データソース編集" : "データソース作成"}
              </Dialog.Title>
            </Dialog.Header>
            <Dialog.CloseTrigger asChild>
              <CloseButton size="sm" />
            </Dialog.CloseTrigger>
            <Dialog.Body>
              <VStack gap={4} align="stretch">
                <Field.Root invalid={!!errors.name}>
                  <Field.Label>名前</Field.Label>
                  <Input
                    value={formData.name}
                    onChange={(e) =>
                      setFormData((prev) => ({ ...prev, name: e.target.value }))
                    }
                    placeholder="本番DB、開発DBなど"
                  />
                  <Field.ErrorText>{errors.name}</Field.ErrorText>
                </Field.Root>

                <Field.Root>
                  <Field.Label>データベース種類</Field.Label>
                  <NativeSelect.Root disabled={isEdit}>
                    <NativeSelect.Field
                      value={formData.db_type}
                      onChange={(e) =>
                        handleDBTypeChange(e.target.value as DBType)
                      }
                    >
                      {DB_TYPES.map((type) => (
                        <option key={type} value={type}>
                          {DB_TYPE_LABELS[type]}
                        </option>
                      ))}
                    </NativeSelect.Field>
                    <NativeSelect.Indicator />
                  </NativeSelect.Root>
                </Field.Root>

                <HStack width="100%">
                  <Field.Root invalid={!!errors.host} flex={3}>
                    <Field.Label>ホスト</Field.Label>
                    <Input
                      value={formData.host}
                      onChange={(e) =>
                        setFormData((prev) => ({
                          ...prev,
                          host: e.target.value,
                        }))
                      }
                      placeholder="localhost または IPアドレス"
                    />
                    <Field.ErrorText>{errors.host}</Field.ErrorText>
                  </Field.Root>

                  <Field.Root invalid={!!errors.port} flex={1}>
                    <Field.Label>ポート</Field.Label>
                    <Input
                      type="number"
                      value={formData.port}
                      onChange={(e) =>
                        setFormData((prev) => ({
                          ...prev,
                          port: parseInt(e.target.value) || 0,
                        }))
                      }
                    />
                    <Field.ErrorText>{errors.port}</Field.ErrorText>
                  </Field.Root>
                </HStack>

                <Field.Root invalid={!!errors.database_name}>
                  <Field.Label>データベース名</Field.Label>
                  <Input
                    value={formData.database_name}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        database_name: e.target.value,
                      }))
                    }
                    placeholder="データベース名"
                  />
                  <Field.ErrorText>{errors.database_name}</Field.ErrorText>
                </Field.Root>

                <Field.Root invalid={!!errors.username}>
                  <Field.Label>ユーザー名</Field.Label>
                  <Input
                    value={formData.username}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        username: e.target.value,
                      }))
                    }
                    placeholder="データベースユーザー名"
                  />
                  <Field.ErrorText>{errors.username}</Field.ErrorText>
                </Field.Root>

                <Field.Root invalid={!!errors.password}>
                  <Field.Label>
                    パスワード{isEdit && "（変更する場合のみ入力）"}
                  </Field.Label>
                  <Input
                    type="password"
                    value={formData.password}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        password: e.target.value,
                      }))
                    }
                    placeholder="パスワード"
                  />
                  <Field.ErrorText>{errors.password}</Field.ErrorText>
                </Field.Root>

                {testResult && (
                  <Alert.Root
                    status={testResult.success ? "success" : "error"}
                    borderRadius="md"
                  >
                    <Alert.Indicator />
                    <Alert.Title>{testResult.message}</Alert.Title>
                  </Alert.Root>
                )}
              </VStack>
            </Dialog.Body>

            <Dialog.Footer>
              <HStack gap={3}>
                <Button variant="ghost" onClick={onClose} disabled={isLoading}>
                  キャンセル
                </Button>
                <Button
                  variant="outline"
                  onClick={handleTestConnection}
                  loading={testMutation.isPending}
                >
                  テスト接続
                </Button>
                <Button
                  colorPalette="blue"
                  onClick={handleSubmit}
                  loading={createMutation.isPending || updateMutation.isPending}
                >
                  {isEdit ? "更新" : "作成"}
                </Button>
              </HStack>
            </Dialog.Footer>
          </Dialog.Content>
        </Dialog.Positioner>
      </Portal>
    </Dialog.Root>
  );
}
