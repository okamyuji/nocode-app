/**
 * データソース一覧コンポーネント
 */

import { useDataSourcesApi } from "@/api";
import { DB_TYPE_LABELS, type DataSource } from "@/types/datasource";
import {
  Badge,
  Box,
  Button,
  Card,
  Dialog,
  Flex,
  Heading,
  HStack,
  IconButton,
  Portal,
  Table,
  Text,
  useDisclosure,
} from "@chakra-ui/react";
import { toaster } from "@/components/ui/toaster";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRef, useState } from "react";
import { FiDatabase, FiEdit2, FiPlus, FiTrash2 } from "react-icons/fi";
import { DataSourceForm } from "./DataSourceForm";

export function DataSourceList() {
  const queryClient = useQueryClient();
  const dataSourcesApi = useDataSourcesApi();
  const cancelRef = useRef<HTMLButtonElement>(null);
  const [selectedDataSource, setSelectedDataSource] =
    useState<DataSource | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<DataSource | null>(null);

  const {
    open: isFormOpen,
    onOpen: onFormOpen,
    onClose: onFormClose,
  } = useDisclosure();
  const {
    open: isDeleteOpen,
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
      toaster.create({
        title: "データソースを削除しました",
        type: "success",
        duration: 3000,
      });
      onDeleteClose();
    },
    onError: () => {
      toaster.create({
        title: "削除に失敗しました",
        type: "error",
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
        <Button colorPalette="blue" onClick={handleCreate}>
          <FiPlus />
          新規データソース
        </Button>
      </Flex>

      {isLoading ? (
        <Text>読み込み中...</Text>
      ) : dataSources.length === 0 ? (
        <Card.Root>
          <Card.Body textAlign="center" py={10}>
            <FiDatabase size={48} style={{ margin: "0 auto", opacity: 0.3 }} />
            <Text mt={4} color="gray.500">
              データソースがありません
            </Text>
            <Button mt={4} colorPalette="blue" onClick={handleCreate}>
              データソースを追加
            </Button>
          </Card.Body>
        </Card.Root>
      ) : (
        <Card.Root>
          <Card.Body p={0}>
            <Table.Root variant="line">
              <Table.Header>
                <Table.Row>
                  <Table.ColumnHeader>名前</Table.ColumnHeader>
                  <Table.ColumnHeader>データベース種類</Table.ColumnHeader>
                  <Table.ColumnHeader>ホスト</Table.ColumnHeader>
                  <Table.ColumnHeader>データベース名</Table.ColumnHeader>
                  <Table.ColumnHeader width="100px">操作</Table.ColumnHeader>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {dataSources.map((ds) => (
                  <Table.Row key={ds.id}>
                    <Table.Cell fontWeight="medium">{ds.name}</Table.Cell>
                    <Table.Cell>
                      <Badge colorPalette={getDBTypeBadgeColor(ds.db_type)}>
                        {DB_TYPE_LABELS[ds.db_type]}
                      </Badge>
                    </Table.Cell>
                    <Table.Cell>
                      {ds.host}:{ds.port}
                    </Table.Cell>
                    <Table.Cell>{ds.database_name}</Table.Cell>
                    <Table.Cell>
                      <HStack gap={1}>
                        <IconButton
                          aria-label="編集"
                          size="sm"
                          variant="ghost"
                          onClick={() => handleEdit(ds)}
                        >
                          <FiEdit2 />
                        </IconButton>
                        <IconButton
                          aria-label="削除"
                          size="sm"
                          variant="ghost"
                          colorPalette="red"
                          onClick={() => handleDelete(ds)}
                        >
                          <FiTrash2 />
                        </IconButton>
                      </HStack>
                    </Table.Cell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table.Root>
          </Card.Body>
        </Card.Root>
      )}

      {/* フォームモーダル */}
      <DataSourceForm
        isOpen={isFormOpen}
        onClose={handleFormClose}
        dataSource={selectedDataSource}
      />

      {/* 削除確認ダイアログ */}
      <Dialog.Root
        role="alertdialog"
        open={isDeleteOpen}
        onOpenChange={(e) => !e.open && onDeleteClose()}
        initialFocusEl={() => cancelRef.current}
      >
        <Portal>
          <Dialog.Backdrop />
          <Dialog.Positioner>
            <Dialog.Content>
              <Dialog.Header fontSize="lg" fontWeight="bold">
                <Dialog.Title>データソースの削除</Dialog.Title>
              </Dialog.Header>
              <Dialog.Body>
                「{deleteTarget?.name}
                」を削除しますか？このデータソースを使用しているアプリにも影響があります。
              </Dialog.Body>
              <Dialog.Footer>
                <Button ref={cancelRef} onClick={onDeleteClose}>
                  キャンセル
                </Button>
                <Button
                  colorPalette="red"
                  onClick={() =>
                    deleteTarget && deleteMutation.mutate(deleteTarget.id)
                  }
                  ml={3}
                  loading={deleteMutation.isPending}
                >
                  削除
                </Button>
              </Dialog.Footer>
            </Dialog.Content>
          </Dialog.Positioner>
        </Portal>
      </Dialog.Root>
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
