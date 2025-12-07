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
  AlertIcon,
  Button,
  FormControl,
  FormErrorMessage,
  FormLabel,
  HStack,
  Input,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Select,
  useToast,
  VStack,
} from "@chakra-ui/react";
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
  const toast = useToast();
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
      toast({
        title: "データソースを作成しました",
        status: "success",
        duration: 3000,
      });
      onClose();
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
      toast({
        title: "データソースを更新しました",
        status: "success",
        duration: 3000,
      });
      onClose();
    },
    onError: (error: Error) => {
      toast({
        title: "更新に失敗しました",
        description: error.message,
        status: "error",
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
    <Modal isOpen={isOpen} onClose={onClose} size="lg">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>
          {isEdit ? "データソース編集" : "データソース作成"}
        </ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <VStack spacing={4}>
            <FormControl isInvalid={!!errors.name}>
              <FormLabel>名前</FormLabel>
              <Input
                value={formData.name}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, name: e.target.value }))
                }
                placeholder="本番DB、開発DBなど"
              />
              <FormErrorMessage>{errors.name}</FormErrorMessage>
            </FormControl>

            <FormControl>
              <FormLabel>データベース種類</FormLabel>
              <Select
                value={formData.db_type}
                onChange={(e) => handleDBTypeChange(e.target.value as DBType)}
                isDisabled={isEdit}
              >
                {DB_TYPES.map((type) => (
                  <option key={type} value={type}>
                    {DB_TYPE_LABELS[type]}
                  </option>
                ))}
              </Select>
            </FormControl>

            <HStack width="100%">
              <FormControl isInvalid={!!errors.host} flex={3}>
                <FormLabel>ホスト</FormLabel>
                <Input
                  value={formData.host}
                  onChange={(e) =>
                    setFormData((prev) => ({ ...prev, host: e.target.value }))
                  }
                  placeholder="localhost または IPアドレス"
                />
                <FormErrorMessage>{errors.host}</FormErrorMessage>
              </FormControl>

              <FormControl isInvalid={!!errors.port} flex={1}>
                <FormLabel>ポート</FormLabel>
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
                <FormErrorMessage>{errors.port}</FormErrorMessage>
              </FormControl>
            </HStack>

            <FormControl isInvalid={!!errors.database_name}>
              <FormLabel>データベース名</FormLabel>
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
              <FormErrorMessage>{errors.database_name}</FormErrorMessage>
            </FormControl>

            <FormControl isInvalid={!!errors.username}>
              <FormLabel>ユーザー名</FormLabel>
              <Input
                value={formData.username}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, username: e.target.value }))
                }
                placeholder="データベースユーザー名"
              />
              <FormErrorMessage>{errors.username}</FormErrorMessage>
            </FormControl>

            <FormControl isInvalid={!!errors.password}>
              <FormLabel>
                パスワード{isEdit && "（変更する場合のみ入力）"}
              </FormLabel>
              <Input
                type="password"
                value={formData.password}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, password: e.target.value }))
                }
                placeholder="パスワード"
              />
              <FormErrorMessage>{errors.password}</FormErrorMessage>
            </FormControl>

            {testResult && (
              <Alert
                status={testResult.success ? "success" : "error"}
                borderRadius="md"
              >
                <AlertIcon />
                {testResult.message}
              </Alert>
            )}
          </VStack>
        </ModalBody>

        <ModalFooter>
          <HStack spacing={3}>
            <Button variant="ghost" onClick={onClose} isDisabled={isLoading}>
              キャンセル
            </Button>
            <Button
              variant="outline"
              onClick={handleTestConnection}
              isLoading={testMutation.isPending}
            >
              テスト接続
            </Button>
            <Button
              colorScheme="blue"
              onClick={handleSubmit}
              isLoading={createMutation.isPending || updateMutation.isPending}
            >
              {isEdit ? "更新" : "作成"}
            </Button>
          </HStack>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
}
