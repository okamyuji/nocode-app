import { AppFormBuilder } from "@/components/apps";
import { ExternalAppFormBuilder } from "@/components/apps/ExternalAppFormBuilder";
import { DataSourceSelector } from "@/components/datasources";
import { useAuth } from "@/hooks";
import type { ColumnInfo, DataSource, TableInfo } from "@/types/datasource";
import { ChevronRightIcon } from "@chakra-ui/icons";
import {
  Alert,
  AlertDescription,
  AlertIcon,
  AlertTitle,
  Box,
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  Button,
  Card,
  CardBody,
  Heading,
  HStack,
  Radio,
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
        <Alert status="warning" borderRadius="md">
          <AlertIcon />
          <Box>
            <AlertTitle>アクセス権限がありません</AlertTitle>
            <AlertDescription>
              アプリの作成は管理者のみ可能です。
            </AlertDescription>
          </Box>
        </Alert>
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
      <Breadcrumb
        spacing="8px"
        separator={<ChevronRightIcon color="gray.500" />}
        mb={4}
      >
        <BreadcrumbItem>
          <BreadcrumbLink as={RouterLink} to="/apps">
            アプリ一覧
          </BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbItem isCurrentPage>
          <BreadcrumbLink>新規作成</BreadcrumbLink>
        </BreadcrumbItem>
      </Breadcrumb>

      <Heading size="lg" mb={6}>
        アプリを作成
      </Heading>

      {/* モード選択 */}
      {!mode && (
        <VStack spacing={4} align="stretch" maxW="600px">
          <Text color="gray.600" mb={2}>
            アプリの作成方法を選択してください
          </Text>
          <RadioGroup onChange={(value) => setMode(value as CreateMode)}>
            <Stack spacing={4}>
              <Card
                cursor="pointer"
                onClick={() => setMode("new")}
                _hover={{ borderColor: "blue.300" }}
                borderWidth="2px"
                borderColor={mode === "new" ? "blue.500" : "gray.200"}
              >
                <CardBody>
                  <HStack spacing={4}>
                    <Radio value="new" size="lg" />
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
                </CardBody>
              </Card>

              <Card
                cursor="pointer"
                onClick={() => setMode("external")}
                _hover={{ borderColor: "blue.300" }}
                borderWidth="2px"
                borderColor={mode === "external" ? "blue.500" : "gray.200"}
              >
                <CardBody>
                  <HStack spacing={4}>
                    <Radio value="external" size="lg" />
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
                </CardBody>
              </Card>
            </Stack>
          </RadioGroup>
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
          <Card>
            <CardBody>
              <Heading size="md" mb={4}>
                データソースとテーブルを選択
              </Heading>
              <DataSourceSelector onSelect={handleExternalSelect} />
            </CardBody>
          </Card>
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
