import { ChartBuilder, ChartRenderer } from "@/components/charts";
import { Loading } from "@/components/common";
import {
  RecordDetail,
  RecordForm,
  RecordPagination,
  RecordTable,
} from "@/components/records";
import { CalendarView, ListView, ViewSelector } from "@/components/views";
import {
  useApp,
  useAuth,
  useBulkDeleteRecords,
  useChartData,
  useCreateRecord,
  useDeleteRecord,
  useFields,
  useRecords,
  useUpdateRecord,
} from "@/hooks";
import { RecordData, RecordItem, RecordQueryOptions, ViewType } from "@/types";
import { ChartDataRequest } from "@/types/chart";
import { AddIcon, ChevronRightIcon, DeleteIcon } from "@chakra-ui/icons";
import {
  AlertDialog,
  AlertDialogBody,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogOverlay,
  Box,
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  Button,
  Card,
  CardBody,
  Heading,
  HStack,
  IconButton,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalHeader,
  ModalOverlay,
  Tooltip,
  useDisclosure,
  useToast,
} from "@chakra-ui/react";
import { useCallback, useMemo, useRef, useState } from "react";
import { Link as RouterLink, useParams } from "react-router-dom";

export function RecordListPage() {
  const { appId } = useParams<{ appId: string }>();
  const numericAppId = Number(appId);
  const { isAdmin } = useAuth();

  const [currentView, setCurrentView] = useState<ViewType>("table");
  const [queryOptions, setQueryOptions] = useState<RecordQueryOptions>({
    page: 1,
    limit: 20,
  });
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [selectedRecord, setSelectedRecord] = useState<RecordItem | null>(null);
  const [editingRecord, setEditingRecord] = useState<RecordItem | null>(null);
  const [chartConfig, setChartConfig] = useState<ChartDataRequest | null>(null);

  const {
    isOpen: isFormOpen,
    onOpen: onFormOpen,
    onClose: onFormClose,
  } = useDisclosure();
  const {
    isOpen: isDetailOpen,
    onOpen: onDetailOpen,
    onClose: onDetailClose,
  } = useDisclosure();
  const {
    isOpen: isDeleteOpen,
    onOpen: onDeleteOpen,
    onClose: onDeleteClose,
  } = useDisclosure();
  const {
    isOpen: isBulkDeleteOpen,
    onOpen: onBulkDeleteOpen,
    onClose: onBulkDeleteClose,
  } = useDisclosure();

  const cancelRef = useRef<HTMLButtonElement>(null);
  const toast = useToast();

  const { data: app, isLoading: isAppLoading } = useApp(numericAppId);
  const { data: fieldsData, isLoading: isFieldsLoading } =
    useFields(numericAppId);
  const { data: recordsData, isLoading: isRecordsLoading } = useRecords(
    numericAppId,
    queryOptions
  );

  const createRecord = useCreateRecord();
  const updateRecord = useUpdateRecord();
  const deleteRecord = useDeleteRecord();
  const bulkDeleteRecords = useBulkDeleteRecords();

  // 外部データソースからのアプリかどうか
  const isExternalApp = app?.is_external === true;

  // Chart data hook (only fetch when in chart view and config is set)
  const { data: chartData } = useChartData(
    numericAppId,
    currentView === "chart" ? chartConfig : null
  );

  const fields = fieldsData?.fields || [];

  // Check if there are date fields for calendar view
  const hasDateField = fields.some(
    (f) => f.field_type === "date" || f.field_type === "datetime"
  );
  const records = useMemo(
    () => recordsData?.records || [],
    [recordsData?.records]
  );
  const pagination = recordsData?.pagination || {
    page: 1,
    limit: 20,
    total: 0,
    total_pages: 0,
  };

  const handleSelectRecord = useCallback((id: number) => {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((i) => i !== id) : [...prev, id]
    );
  }, []);

  const handleSelectAll = useCallback(() => {
    if (selectedIds.length === records.length) {
      setSelectedIds([]);
    } else {
      setSelectedIds(records.map((r) => r.id));
    }
  }, [selectedIds.length, records]);

  const handleView = useCallback(
    (record: RecordItem) => {
      setSelectedRecord(record);
      onDetailOpen();
    },
    [onDetailOpen]
  );

  const handleEdit = useCallback(
    (record: RecordItem) => {
      setEditingRecord(record);
      onFormOpen();
    },
    [onFormOpen]
  );

  const handleDelete = useCallback(
    (record: RecordItem) => {
      setSelectedRecord(record);
      onDeleteOpen();
    },
    [onDeleteOpen]
  );

  const handleCreateNew = useCallback(() => {
    setEditingRecord(null);
    onFormOpen();
  }, [onFormOpen]);

  const handleFormSubmit = async (data: RecordData) => {
    try {
      if (editingRecord) {
        await updateRecord.mutateAsync({
          appId: numericAppId,
          recordId: editingRecord.id,
          data: { data },
        });
        toast({
          title: "レコードを更新しました",
          status: "success",
          duration: 3000,
        });
      } else {
        await createRecord.mutateAsync({
          appId: numericAppId,
          data: { data },
        });
        toast({
          title: "レコードを作成しました",
          status: "success",
          duration: 3000,
        });
      }
      onFormClose();
      setEditingRecord(null);
    } catch {
      toast({
        title: editingRecord ? "更新に失敗しました" : "作成に失敗しました",
        status: "error",
        duration: 5000,
      });
    }
  };

  const handleConfirmDelete = async () => {
    if (!selectedRecord) return;
    try {
      await deleteRecord.mutateAsync({
        appId: numericAppId,
        recordId: selectedRecord.id,
      });
      toast({
        title: "レコードを削除しました",
        status: "success",
        duration: 3000,
      });
      onDeleteClose();
      setSelectedRecord(null);
    } catch {
      toast({
        title: "削除に失敗しました",
        status: "error",
        duration: 5000,
      });
    }
  };

  const handleBulkDelete = async () => {
    try {
      await bulkDeleteRecords.mutateAsync({
        appId: numericAppId,
        data: { ids: selectedIds },
      });
      toast({
        title: `${selectedIds.length}件のレコードを削除しました`,
        status: "success",
        duration: 3000,
      });
      onBulkDeleteClose();
      setSelectedIds([]);
    } catch {
      toast({
        title: "一括削除に失敗しました",
        status: "error",
        duration: 5000,
      });
    }
  };

  if (isAppLoading || isFieldsLoading) {
    return <Loading message="アプリを読み込み中..." />;
  }

  if (!app) {
    return <Box>アプリが見つかりません</Box>;
  }

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
          <BreadcrumbLink>{app.name}</BreadcrumbLink>
        </BreadcrumbItem>
      </Breadcrumb>

      <HStack justify="space-between" mb={6}>
        <Heading size="lg">{app.name}</Heading>
        <HStack spacing={4}>
          <ViewSelector
            currentView={currentView}
            onViewChange={setCurrentView}
            hasDateField={hasDateField}
          />
          {isAdmin &&
            selectedIds.length > 0 &&
            currentView === "table" &&
            !isExternalApp && (
              <Tooltip label={`${selectedIds.length}件を削除`}>
                <IconButton
                  icon={<DeleteIcon />}
                  aria-label="一括削除"
                  colorScheme="red"
                  variant="outline"
                  onClick={onBulkDeleteOpen}
                />
              </Tooltip>
            )}
          {isAdmin && currentView !== "chart" && !isExternalApp && (
            <Button
              leftIcon={<AddIcon />}
              colorScheme="brand"
              onClick={handleCreateNew}
            >
              レコード追加
            </Button>
          )}
        </HStack>
      </HStack>

      <Card>
        <CardBody p={currentView === "chart" ? 4 : 0}>
          {isRecordsLoading && currentView !== "chart" ? (
            <Loading message="レコードを読み込み中..." />
          ) : (
            <>
              {/* Table View */}
              {currentView === "table" && (
                <>
                  <RecordTable
                    records={records}
                    fields={fields}
                    selectedIds={selectedIds}
                    onSelectRecord={handleSelectRecord}
                    onSelectAll={handleSelectAll}
                    onView={handleView}
                    onEdit={isExternalApp ? undefined : handleEdit}
                    onDelete={isExternalApp ? undefined : handleDelete}
                    isAdmin={isAdmin && !isExternalApp}
                  />
                  <Box px={4}>
                    <RecordPagination
                      pagination={pagination}
                      onPageChange={(page) =>
                        setQueryOptions((prev) => ({ ...prev, page }))
                      }
                      onLimitChange={(limit) =>
                        setQueryOptions((prev) => ({ ...prev, limit, page: 1 }))
                      }
                    />
                  </Box>
                </>
              )}

              {/* List View */}
              {currentView === "list" && (
                <>
                  <ListView
                    records={records}
                    fields={fields}
                    onView={handleView}
                    onEdit={isExternalApp ? undefined : handleEdit}
                    onDelete={isExternalApp ? undefined : handleDelete}
                    isAdmin={isAdmin && !isExternalApp}
                  />
                  <Box px={4}>
                    <RecordPagination
                      pagination={pagination}
                      onPageChange={(page) =>
                        setQueryOptions((prev) => ({ ...prev, page }))
                      }
                      onLimitChange={(limit) =>
                        setQueryOptions((prev) => ({ ...prev, limit, page: 1 }))
                      }
                    />
                  </Box>
                </>
              )}

              {/* Calendar View */}
              {currentView === "calendar" && (
                <Box p={4}>
                  <CalendarView
                    records={records}
                    fields={fields}
                    onRecordClick={handleView}
                  />
                </Box>
              )}

              {/* Chart View */}
              {currentView === "chart" && (
                <Box>
                  <ChartBuilder
                    fields={fields}
                    config={chartConfig || undefined}
                    onConfigChange={setChartConfig}
                  />
                  {chartConfig && chartData && (
                    <Box mt={6} h="400px">
                      <ChartRenderer
                        chartType={chartConfig.chart_type}
                        data={chartData}
                      />
                    </Box>
                  )}
                </Box>
              )}
            </>
          )}
        </CardBody>
      </Card>

      {/* Create/Edit Modal */}
      <Modal isOpen={isFormOpen} onClose={onFormClose} size="xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>
            {editingRecord ? "レコードを編集" : "新規レコード"}
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            <RecordForm
              fields={fields}
              record={editingRecord || undefined}
              onSubmit={handleFormSubmit}
              onCancel={onFormClose}
              isSubmitting={createRecord.isPending || updateRecord.isPending}
            />
          </ModalBody>
        </ModalContent>
      </Modal>

      {/* Detail Modal */}
      <Modal isOpen={isDetailOpen} onClose={onDetailClose} size="xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>レコード詳細</ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            {selectedRecord && (
              <RecordDetail record={selectedRecord} fields={fields} />
            )}
          </ModalBody>
        </ModalContent>
      </Modal>

      {/* Delete Confirmation */}
      <AlertDialog
        isOpen={isDeleteOpen}
        leastDestructiveRef={cancelRef}
        onClose={onDeleteClose}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader>レコードの削除</AlertDialogHeader>
            <AlertDialogBody>
              このレコードを削除しますか？この操作は取り消せません。
            </AlertDialogBody>
            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={onDeleteClose}>
                キャンセル
              </Button>
              <Button
                colorScheme="red"
                onClick={handleConfirmDelete}
                ml={3}
                isLoading={deleteRecord.isPending}
              >
                削除
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>

      {/* Bulk Delete Confirmation */}
      <AlertDialog
        isOpen={isBulkDeleteOpen}
        leastDestructiveRef={cancelRef}
        onClose={onBulkDeleteClose}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader>レコードの一括削除</AlertDialogHeader>
            <AlertDialogBody>
              {selectedIds.length}
              件のレコードを削除しますか？この操作は取り消せません。
            </AlertDialogBody>
            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={onBulkDeleteClose}>
                キャンセル
              </Button>
              <Button
                colorScheme="red"
                onClick={handleBulkDelete}
                ml={3}
                isLoading={bulkDeleteRecords.isPending}
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
