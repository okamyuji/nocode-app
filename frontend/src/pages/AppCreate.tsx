import { AppFormBuilder } from "@/components/apps";
import { ExternalAppFormBuilder } from "@/components/apps/ExternalAppFormBuilder";
import { DataSourceSelector } from "@/components/datasources";
import { useAuth } from "@/hooks";
import type { ColumnInfo, DataSource, TableInfo } from "@/types/datasource";
import { ChevronRightIcon } from "@chakra-ui/icons/ChevronRight";
import {
  Alert,
  Box,
  Breadcrumb,
  Button,
  Card,
  Heading,
  HStack,
  RadioGroup,
  Stack,
  Text,
  VStack,
} from "@chakra-ui/react";
import { useState } from "react";
import { FiDatabase, FiPlusCircle } from "react-icons/fi";
import { Navigate, Link as RouterLink } from "react-router-dom";

type CreateMode = "new" | "external";

export function AppCreatePage() {
  const { isAdmin, isAuthenticated } = useAuth();
  const [mode, setMode] = useState<CreateMode | null>(null);
  const [externalSelection, setExternalSelection] = useState<{
    dataSource: DataSource;
    table: TableInfo;
    columns: ColumnInfo[];
  } | null>(null);

  // Redirect non-admin users
  if (isAuthenticated && !isAdmin) {
    return (
      <Box p={6}>
        <Alert.Root status="warning" borderRadius="md">
          <Alert.Indicator />
          <Alert.Content>
            <Alert.Title>アクセス権限がありません</Alert.Title>
            <Alert.Description>
              アプリの作成は管理者のみ可能です。
            </Alert.Description>
          </Alert.Content>
        </Alert.Root>
        <Box mt={4}>
          <Navigate to="/apps" replace />
        </Box>
      </Box>
    );
  }

  const handleExternalSelect = (
    dataSource: DataSource,
    table: TableInfo,
    columns: ColumnInfo[]
  ) => {
    setExternalSelection({ dataSource, table, columns });
  };

  const handleBack = () => {
    if (externalSelection) {
      setExternalSelection(null);
    } else {
      setMode(null);
    }
  };

  return (
    <Box>
      <Breadcrumb.Root mb={4}>
        <Breadcrumb.List gap="8px">
          <Breadcrumb.Item>
            <Breadcrumb.Link asChild>
              <RouterLink to="/apps">アプリ一覧</RouterLink>
            </Breadcrumb.Link>
          </Breadcrumb.Item>
          <Breadcrumb.Separator>
            <ChevronRightIcon color="gray.500" />
          </Breadcrumb.Separator>
          <Breadcrumb.Item>
            <Breadcrumb.CurrentLink>新規作成</Breadcrumb.CurrentLink>
          </Breadcrumb.Item>
        </Breadcrumb.List>
      </Breadcrumb.Root>

      <Heading size="lg" mb={6}>
        アプリを作成
      </Heading>

      {/* モード選択 */}
      {!mode && (
        <VStack gap={4} align="stretch" maxW="600px">
          <Text color="gray.600" mb={2}>
            アプリの作成方法を選択してください
          </Text>
          <RadioGroup.Root
            size="lg"
            onValueChange={(e) => e.value && setMode(e.value as CreateMode)}
          >
            <Stack gap={4}>
              <Card.Root
                cursor="pointer"
                onClick={() => setMode("new")}
                _hover={{ borderColor: "blue.300" }}
                borderWidth="2px"
                borderColor={mode === "new" ? "blue.500" : "gray.200"}
              >
                <Card.Body>
                  <HStack gap={4}>
                    <RadioGroup.Item value="new">
                      <RadioGroup.ItemHiddenInput />
                      <RadioGroup.ItemIndicator />
                    </RadioGroup.Item>
                    <Box flex={1}>
                      <HStack>
                        <FiPlusCircle size={20} />
                        <Text fontWeight="bold">新規テーブル作成</Text>
                      </HStack>
                      <Text fontSize="sm" color="gray.500" mt={1}>
                        新しいテーブルを作成し、フィールドを自由に設計します
                      </Text>
                    </Box>
                  </HStack>
                </Card.Body>
              </Card.Root>

              <Card.Root
                cursor="pointer"
                onClick={() => setMode("external")}
                _hover={{ borderColor: "blue.300" }}
                borderWidth="2px"
                borderColor={mode === "external" ? "blue.500" : "gray.200"}
              >
                <Card.Body>
                  <HStack gap={4}>
                    <RadioGroup.Item value="external">
                      <RadioGroup.ItemHiddenInput />
                      <RadioGroup.ItemIndicator />
                    </RadioGroup.Item>
                    <Box flex={1}>
                      <HStack>
                        <FiDatabase size={20} />
                        <Text fontWeight="bold">外部データソース接続</Text>
                      </HStack>
                      <Text fontSize="sm" color="gray.500" mt={1}>
                        既存のデータベースに接続し、テーブルデータを表示します（読み取り専用）
                      </Text>
                    </Box>
                  </HStack>
                </Card.Body>
              </Card.Root>
            </Stack>
          </RadioGroup.Root>
        </VStack>
      )}

      {/* 新規テーブル作成モード */}
      {mode === "new" && (
        <Box>
          <Button variant="ghost" mb={4} onClick={handleBack}>
            ← 戻る
          </Button>
          <AppFormBuilder />
        </Box>
      )}

      {/* 外部データソース接続モード */}
      {mode === "external" && !externalSelection && (
        <Box maxW="600px">
          <Button variant="ghost" mb={4} onClick={handleBack}>
            ← 戻る
          </Button>
          <Card.Root>
            <Card.Body>
              <Heading size="md" mb={4}>
                データソースとテーブルを選択
              </Heading>
              <DataSourceSelector onSelect={handleExternalSelect} />
            </Card.Body>
          </Card.Root>
        </Box>
      )}

      {/* 外部アプリのフィールド設定 */}
      {mode === "external" && externalSelection && (
        <Box>
          <Button variant="ghost" mb={4} onClick={handleBack}>
            ← 戻る
          </Button>
          <ExternalAppFormBuilder
            dataSource={externalSelection.dataSource}
            table={externalSelection.table}
            columns={externalSelection.columns}
          />
        </Box>
      )}
    </Box>
  );
}
