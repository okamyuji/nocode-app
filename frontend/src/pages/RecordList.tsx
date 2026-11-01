import { ChartBuilder, ChartRenderer } from "@/components/charts";
import { Loading } from "@/components/common";
import { toaster } from "@/components/ui/toaster";
import { Tooltip } from "@/components/ui/tooltip";
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
import { AddIcon } from "@chakra-ui/icons/Add";
import { ChevronRightIcon } from "@chakra-ui/icons/ChevronRight";
import { DeleteIcon } from "@chakra-ui/icons/Delete";
import {
  Box,
  Breadcrumb,
  Button,
  Card,
  CloseButton,
  Dialog,
  Heading,
  HStack,
  IconButton,
  Portal,
  useDisclosure,
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
    open: isFormOpen,
    onOpen: onFormOpen,
    onClose: onFormClose,
  } = useDisclosure();
  const {
    open: isDetailOpen,
    onOpen: onDetailOpen,
    onClose: onDetailClose,
  } = useDisclosure();
  const {
    open: isDeleteOpen,
    onOpen: onDeleteOpen,
    onClose: onDeleteClose,
  } = useDisclosure();
  const {
    open: isBulkDeleteOpen,
    onOpen: onBulkDeleteOpen,
    onClose: onBulkDeleteClose,
  } = useDisclosure();

  const cancelRef = useRef<HTMLButtonElement>(null);

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
        toaster.create({
          title: "レコードを更新しました",
          type: "success",
          duration: 3000,
        });
      } else {
        await createRecord.mutateAsync({
          appId: numericAppId,
          data: { data },
        });
        toaster.create({
          title: "レコードを作成しました",
          type: "success",
          duration: 3000,
        });
      }
      onFormClose();
      setEditingRecord(null);
    } catch {
      toaster.create({
        title: editingRecord ? "更新に失敗しました" : "作成に失敗しました",
        type: "error",
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
      toaster.create({
        title: "レコードを削除しました",
        type: "success",
        duration: 3000,
      });
      onDeleteClose();
      setSelectedRecord(null);
    } catch {
      toaster.create({
        title: "削除に失敗しました",
        type: "error",
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
      toaster.create({
        title: `${selectedIds.length}件のレコードを削除しました`,
        type: "success",
        duration: 3000,
      });
      onBulkDeleteClose();
      setSelectedIds([]);
    } catch {
      toaster.create({
        title: "一括削除に失敗しました",
        type: "error",
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
            <Breadcrumb.CurrentLink>{app.name}</Breadcrumb.CurrentLink>
          </Breadcrumb.Item>
        </Breadcrumb.List>
      </Breadcrumb.Root>

      <HStack justify="space-between" mb={6}>
        <Heading size="lg">{app.name}</Heading>
        <HStack gap={4}>
          <ViewSelector
            currentView={currentView}
            onViewChange={setCurrentView}
            hasDateField={hasDateField}
          />
          {isAdmin &&
            selectedIds.length > 0 &&
            currentView === "table" &&
            !isExternalApp && (
              <Tooltip content={`${selectedIds.length}件を削除`}>
                <IconButton
                  aria-label="一括削除"
                  colorPalette="red"
                  variant="outline"
                  onClick={onBulkDeleteOpen}
                >
                  <DeleteIcon />
                </IconButton>
              </Tooltip>
            )}
          {isAdmin && currentView !== "chart" && !isExternalApp && (
            <Button colorPalette="brand" onClick={handleCreateNew}>
              <AddIcon />
              レコード追加
            </Button>
          )}
        </HStack>
      </HStack>

      <Card.Root overflow="hidden">
        <Card.Body p={currentView === "chart" ? 4 : 0} minW={0}>
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
        </Card.Body>
      </Card.Root>

      {/* Create/Edit Modal */}
      <Dialog.Root
        open={isFormOpen}
        onOpenChange={(e) => !e.open && onFormClose()}
        size="xl"
      >
        <Portal>
          <Dialog.Backdrop />
          <Dialog.Positioner>
            <Dialog.Content>
              <Dialog.Header>
                <Dialog.Title>
                  {editingRecord ? "レコードを編集" : "新規レコード"}
                </Dialog.Title>
              </Dialog.Header>
              <Dialog.CloseTrigger asChild>
                <CloseButton size="sm" />
              </Dialog.CloseTrigger>
              <Dialog.Body pb={6}>
                <RecordForm
                  fields={fields}
                  record={editingRecord || undefined}
                  onSubmit={handleFormSubmit}
                  onCancel={onFormClose}
                  isSubmitting={
                    createRecord.isPending || updateRecord.isPending
                  }
                />
              </Dialog.Body>
            </Dialog.Content>
          </Dialog.Positioner>
        </Portal>
      </Dialog.Root>

      {/* Detail Modal */}
      <Dialog.Root
        open={isDetailOpen}
        onOpenChange={(e) => !e.open && onDetailClose()}
        size="xl"
      >
        <Portal>
          <Dialog.Backdrop />
          <Dialog.Positioner>
            <Dialog.Content>
              <Dialog.Header>
                <Dialog.Title>レコード詳細</Dialog.Title>
              </Dialog.Header>
              <Dialog.CloseTrigger asChild>
                <CloseButton size="sm" />
              </Dialog.CloseTrigger>
              <Dialog.Body pb={6}>
                {selectedRecord && (
                  <RecordDetail record={selectedRecord} fields={fields} />
                )}
              </Dialog.Body>
            </Dialog.Content>
          </Dialog.Positioner>
        </Portal>
      </Dialog.Root>

      {/* Delete Confirmation */}
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
                <Dialog.Title>レコードの削除</Dialog.Title>
              </Dialog.Header>
              <Dialog.Body>
                このレコードを削除しますか？この操作は取り消せません。
              </Dialog.Body>
              <Dialog.Footer>
                <Button ref={cancelRef} onClick={onDeleteClose}>
                  キャンセル
                </Button>
                <Button
                  colorPalette="red"
                  onClick={handleConfirmDelete}
                  ml={3}
                  loading={deleteRecord.isPending}
                >
                  削除
                </Button>
              </Dialog.Footer>
            </Dialog.Content>
          </Dialog.Positioner>
        </Portal>
      </Dialog.Root>

      {/* Bulk Delete Confirmation */}
      <Dialog.Root
        role="alertdialog"
        open={isBulkDeleteOpen}
        onOpenChange={(e) => !e.open && onBulkDeleteClose()}
        initialFocusEl={() => cancelRef.current}
      >
        <Portal>
          <Dialog.Backdrop />
          <Dialog.Positioner>
            <Dialog.Content>
              <Dialog.Header fontSize="lg" fontWeight="bold">
                <Dialog.Title>レコードの一括削除</Dialog.Title>
              </Dialog.Header>
              <Dialog.Body>
                {selectedIds.length}
                件のレコードを削除しますか？この操作は取り消せません。
              </Dialog.Body>
              <Dialog.Footer>
                <Button ref={cancelRef} onClick={onBulkDeleteClose}>
                  キャンセル
                </Button>
                <Button
                  colorPalette="red"
                  onClick={handleBulkDelete}
                  ml={3}
                  loading={bulkDeleteRecords.isPending}
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
