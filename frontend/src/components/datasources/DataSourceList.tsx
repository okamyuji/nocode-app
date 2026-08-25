/**
 * データソース一覧コンポーネント
 */

import { useDataSourcesApi } from "@/api";
import { DB_TYPE_LABELS, type DataSource } from "@/types/datasource";
import {
  AlertDialog,
  AlertDialogBody,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogOverlay,
  Badge,
  Box,
  Button,
  Card,
  CardBody,
  Flex,
  Heading,
  HStack,
  IconButton,
  Table,
  Tbody,
  Td,
  Text,
  Th,
  Thead,
  Tr,
  useDisclosure,
  useToast,
} from "@chakra-ui/react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRef, useState } from "react";
import { FiDatabase, FiEdit2, FiPlus, FiTrash2 } from "react-icons/fi";
import { DataSourceForm } from "./DataSourceForm";

export function DataSourceList() {
  const toast = useToast();
  const queryClient = useQueryClient();
  const dataSourcesApi = useDataSourcesApi();
  const cancelRef = useRef<HTMLButtonElement>(null);
  const [selectedDataSource, setSelectedDataSource] =
    useState<DataSource | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<DataSource | null>(null);

  const {
    isOpen: isFormOpen,
    onOpen: onFormOpen,
    onClose: onFormClose,
  } = useDisclosure();
  const {
    isOpen: isDeleteOpen,
    onOpen: onDeleteOpen,
    onClose: onDeleteClose,
  } = useDisclosure();

  const { data, isLoading } = useQuery({
    queryKey: ["dataSources"],
    queryFn: () => dataSourcesApi.getDataSources(1, 100),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => dataSourcesApi.deleteDataSource(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dataSources"] });
      toast({
        title: "データソースを削除しました",
        status: "success",
        duration: 3000,
      });
      onDeleteClose();
    },
    onError: () => {
      toast({
        title: "削除に失敗しました",
        status: "error",
        duration: 3000,
      });
    },
  });

  const handleEdit = (ds: DataSource) => {
    setSelectedDataSource(ds);
    onFormOpen();
  };

  const handleDelete = (ds: DataSource) => {
    setDeleteTarget(ds);
    onDeleteOpen();
  };

  const handleFormClose = () => {
    setSelectedDataSource(null);
    onFormClose();
  };

  const handleCreate = () => {
    setSelectedDataSource(null);
    onFormOpen();
  };

  const dataSources = data?.data_sources || [];

  return (
    <Box>
      <Flex justify="space-between" align="center" mb={6}>
        <HStack>
          <FiDatabase size={24} />
          <Heading size="md">データソース管理</Heading>
        </HStack>
        <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={handleCreate}>
          新規データソース
        </Button>
      </Flex>

      {isLoading ? (
        <Text>読み込み中...</Text>
      ) : dataSources.length === 0 ? (
        <Card>
          <CardBody textAlign="center" py={10}>
            <FiDatabase size={48} style={{ margin: "0 auto", opacity: 0.3 }} />
            <Text mt={4} color="gray.500">
              データソースがありません
            </Text>
            <Button mt={4} colorScheme="blue" onClick={handleCreate}>
              データソースを追加
            </Button>
          </CardBody>
        </Card>
      ) : (
        <Card>
          <CardBody p={0}>
            <Table variant="simple">
              <Thead>
                <Tr>
                  <Th>名前</Th>
                  <Th>データベース種類</Th>
                  <Th>ホスト</Th>
                  <Th>データベース名</Th>
                  <Th width="100px">操作</Th>
                </Tr>
              </Thead>
              <Tbody>
                {dataSources.map((ds) => (
                  <Tr key={ds.id}>
                    <Td fontWeight="medium">{ds.name}</Td>
                    <Td>
                      <Badge colorScheme={getDBTypeBadgeColor(ds.db_type)}>
                        {DB_TYPE_LABELS[ds.db_type]}
                      </Badge>
                    </Td>
                    <Td>
                      {ds.host}:{ds.port}
                    </Td>
                    <Td>{ds.database_name}</Td>
                    <Td>
                      <HStack spacing={1}>
                        <IconButton
                          aria-label="編集"
                          icon={<FiEdit2 />}
                          size="sm"
                          variant="ghost"
                          onClick={() => handleEdit(ds)}
                        />
                        <IconButton
                          aria-label="削除"
                          icon={<FiTrash2 />}
                          size="sm"
                          variant="ghost"
                          colorScheme="red"
                          onClick={() => handleDelete(ds)}
                        />
                      </HStack>
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          </CardBody>
        </Card>
      )}

      {/* フォームモーダル */}
      <DataSourceForm
        isOpen={isFormOpen}
        onClose={handleFormClose}
        dataSource={selectedDataSource}
      />

      {/* 削除確認ダイアログ */}
      <AlertDialog
        isOpen={isDeleteOpen}
        leastDestructiveRef={cancelRef}
        onClose={onDeleteClose}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader fontSize="lg" fontWeight="bold">
              データソースの削除
            </AlertDialogHeader>
            <AlertDialogBody>
              「{deleteTarget?.name}
              」を削除しますか？このデータソースを使用しているアプリにも影響があります。
            </AlertDialogBody>
            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={onDeleteClose}>
                キャンセル
              </Button>
              <Button
                colorScheme="red"
                onClick={() =>
                  deleteTarget && deleteMutation.mutate(deleteTarget.id)
                }
                ml={3}
                isLoading={deleteMutation.isPending}
              >
                削除
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
    </Box>
  );
}

function getDBTypeBadgeColor(dbType: string): string {
  switch (dbType) {
    case "postgresql":
      return "blue";
    case "mysql":
      return "orange";
    case "oracle":
      return "red";
    case "sqlserver":
      return "purple";
    default:
      return "gray";
  }
}
