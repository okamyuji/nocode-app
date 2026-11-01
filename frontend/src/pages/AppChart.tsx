import { ChartConfigForm, ChartRenderer } from "@/components/charts";
import { Loading } from "@/components/common";
import { toaster } from "@/components/ui/toaster";
import {
  useApp,
  useChartConfigs,
  useChartData,
  useDeleteChartConfig,
  useFields,
  useSaveChartConfig,
} from "@/hooks";
import {
  ChartConfig,
  ChartDataRequest,
  ChartType,
  CHART_TYPE_LABELS,
} from "@/types";
import { ChevronRightIcon } from "@chakra-ui/icons/ChevronRight";
import {
  Box,
  Breadcrumb,
  Button,
  Card,
  Dialog,
  Field,
  Flex,
  Heading,
  HStack,
  IconButton,
  Input,
  Portal,
  SimpleGrid,
  Tag,
  Text,
  useDisclosure,
  VStack,
} from "@chakra-ui/react";
import { useRef, useState } from "react";
import { FiEdit2, FiPlus, FiTrash2 } from "react-icons/fi";
import { Link as RouterLink, useParams } from "react-router-dom";

export function AppChartPage() {
  const { appId } = useParams<{ appId: string }>();
  const numericAppId = Number(appId);
  const cancelRef = useRef<HTMLButtonElement>(null);

  // 編集モード管理
  const [isCreating, setIsCreating] = useState(false);
  const [editingConfig, setEditingConfig] = useState<ChartConfig | null>(null);
  const [chartName, setChartName] = useState("");
  const [chartConfig, setChartConfig] = useState<ChartDataRequest | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ChartConfig | null>(null);

  // 削除確認ダイアログ
  const {
    open: isDeleteOpen,
    onOpen: onDeleteOpen,
    onClose: onDeleteClose,
  } = useDisclosure();

  const { data: app, isLoading: isAppLoading } = useApp(numericAppId);
  const { data: fieldsData, isLoading: isFieldsLoading } =
    useFields(numericAppId);
  const { data: chartConfigsData, isLoading: isConfigsLoading } =
    useChartConfigs(numericAppId);
  const { data: chartData, isLoading: isChartLoading } = useChartData(
    numericAppId,
    chartConfig
  );

  const saveChartConfigMutation = useSaveChartConfig();
  const deleteChartConfigMutation = useDeleteChartConfig();

  const fields = fieldsData?.fields || [];
  const savedConfigs = chartConfigsData?.configs || [];

  // 新規作成モード開始
  const handleStartCreate = () => {
    setIsCreating(true);
    setEditingConfig(null);
    setChartName("");
    setChartConfig(null);
  };

  // 編集モード開始
  const handleStartEdit = (config: ChartConfig) => {
    setIsCreating(false);
    setEditingConfig(config);
    setChartName(config.name);
    setChartConfig(config.config);
  };

  // キャンセル
  const handleCancel = () => {
    setIsCreating(false);
    setEditingConfig(null);
    setChartName("");
    setChartConfig(null);
  };

  // 保存
  const handleSave = async () => {
    if (!chartName.trim()) {
      toaster.create({
        title: "グラフ名を入力してください",
        type: "warning",
        duration: 3000,
      });
      return;
    }
    if (!chartConfig?.x_axis?.field) {
      toaster.create({
        title: "X軸フィールドを選択してください",
        type: "warning",
        duration: 3000,
      });
      return;
    }

    try {
      await saveChartConfigMutation.mutateAsync({
        appId: numericAppId,
        data: {
          id: editingConfig?.id,
          name: chartName.trim(),
          chart_type: chartConfig.chart_type,
          config: chartConfig,
        },
      });
      toaster.create({
        title: editingConfig
          ? "グラフ設定を更新しました"
          : "グラフ設定を保存しました",
        type: "success",
        duration: 3000,
      });
      handleCancel();
    } catch {
      toaster.create({
        title: "保存に失敗しました",
        type: "error",
        duration: 3000,
      });
    }
  };

  // 削除確認
  const handleDeleteClick = (config: ChartConfig) => {
    setDeleteTarget(config);
    onDeleteOpen();
  };

  // 削除実行
  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await deleteChartConfigMutation.mutateAsync({
        appId: numericAppId,
        configId: deleteTarget.id,
      });
      toaster.create({
        title: "グラフ設定を削除しました",
        type: "success",
        duration: 3000,
      });
      onDeleteClose();
      setDeleteTarget(null);
    } catch {
      toaster.create({
        title: "削除に失敗しました",
        type: "error",
        duration: 3000,
      });
    }
  };

  if (isAppLoading || isFieldsLoading || isConfigsLoading) {
    return <Loading message="アプリを読み込み中..." />;
  }

  if (!app) {
    return <Box>アプリが見つかりません</Box>;
  }

  const isEditing = isCreating || editingConfig !== null;

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
            <Breadcrumb.Link asChild>
              <RouterLink to={`/apps/${appId}/records`}>{app.name}</RouterLink>
            </Breadcrumb.Link>
          </Breadcrumb.Item>
          <Breadcrumb.Separator>
            <ChevronRightIcon color="gray.500" />
          </Breadcrumb.Separator>
          <Breadcrumb.Item>
            <Breadcrumb.CurrentLink>グラフ設定</Breadcrumb.CurrentLink>
          </Breadcrumb.Item>
        </Breadcrumb.List>
      </Breadcrumb.Root>

      <Flex justify="space-between" align="center" mb={6}>
        <Heading size="lg">{app.name} - グラフ設定</Heading>
        {!isEditing && (
          <Button colorPalette="brand" onClick={handleStartCreate}>
            <FiPlus />
            新規グラフ
          </Button>
        )}
      </Flex>

      {isEditing ? (
        // 編集フォーム
        <SimpleGrid columns={{ base: 1, lg: 2 }} gap={6}>
          <Card.Root>
            <Card.Header>
              <Heading size="md">
                {editingConfig ? "グラフ設定を編集" : "新規グラフ設定"}
              </Heading>
            </Card.Header>
            <Card.Body>
              <VStack gap={4} align="stretch">
                <Field.Root required>
                  <Field.Label>
                    グラフ名
                    <Field.RequiredIndicator />
                  </Field.Label>
                  <Input
                    value={chartName}
                    onChange={(e) => setChartName(e.target.value)}
                    placeholder="例: 月別売上推移"
                  />
                </Field.Root>
                <ChartConfigForm
                  fields={fields}
                  onSubmit={setChartConfig}
                  initialConfig={
                    chartConfig || editingConfig?.config || undefined
                  }
                  isLoading={isChartLoading}
                />
              </VStack>
            </Card.Body>
          </Card.Root>

          <Card.Root>
            <Card.Header>
              <Heading size="md">プレビュー</Heading>
            </Card.Header>
            <Card.Body>
              {isChartLoading ? (
                <Loading message="グラフを生成中..." />
              ) : chartData ? (
                <Box>
                  <ChartRenderer
                    data={chartData}
                    chartType={chartConfig?.chart_type || "bar"}
                    height={350}
                  />
                  <HStack mt={4} justify="flex-end" gap={3}>
                    <Button variant="ghost" onClick={handleCancel}>
                      キャンセル
                    </Button>
                    <Button
                      colorPalette="brand"
                      onClick={handleSave}
                      loading={saveChartConfigMutation.isPending}
                    >
                      {editingConfig ? "更新" : "保存"}
                    </Button>
                  </HStack>
                </Box>
              ) : (
                <Box
                  h="400px"
                  display="flex"
                  alignItems="center"
                  justifyContent="center"
                  border="2px dashed"
                  borderColor="gray.200"
                  borderRadius="md"
                >
                  <Text color="gray.400">
                    グラフ設定を入力して「グラフを表示」をクリックしてください
                  </Text>
                </Box>
              )}
            </Card.Body>
          </Card.Root>
        </SimpleGrid>
      ) : (
        // 保存済みグラフ一覧
        <Box>
          {savedConfigs.length === 0 ? (
            <Card.Root>
              <Card.Body>
                <VStack py={10} gap={4}>
                  <Text color="gray.500">保存済みのグラフ設定がありません</Text>
                  <Button colorPalette="brand" onClick={handleStartCreate}>
                    <FiPlus />
                    最初のグラフを作成
                  </Button>
                </VStack>
              </Card.Body>
            </Card.Root>
          ) : (
            <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} gap={4}>
              {savedConfigs.map((config) => (
                <SavedChartCard
                  key={config.id}
                  config={config}
                  appId={numericAppId}
                  onEdit={handleStartEdit}
                  onDelete={handleDeleteClick}
                />
              ))}
            </SimpleGrid>
          )}
        </Box>
      )}

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
                <Dialog.Title>グラフ設定の削除</Dialog.Title>
              </Dialog.Header>
              <Dialog.Body>
                「{deleteTarget?.name}
                」を削除しますか？この操作は取り消せません。
              </Dialog.Body>
              <Dialog.Footer>
                <Button ref={cancelRef} onClick={onDeleteClose}>
                  キャンセル
                </Button>
                <Button
                  colorPalette="red"
                  onClick={handleDelete}
                  ml={3}
                  loading={deleteChartConfigMutation.isPending}
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

// 保存済みグラフカードコンポーネント
interface SavedChartCardProps {
  config: ChartConfig;
  appId: number;
  onEdit: (config: ChartConfig) => void;
  onDelete: (config: ChartConfig) => void;
}

function SavedChartCard({
  config,
  appId,
  onEdit,
  onDelete,
}: SavedChartCardProps) {
  const { data: chartData, isLoading } = useChartData(appId, config.config);

  return (
    <Card.Root>
      <Card.Header pb={2}>
        <Flex justify="space-between" align="center">
          <Heading size="sm" lineClamp={1}>
            {config.name}
          </Heading>
          <HStack gap={1}>
            <IconButton
              aria-label="編集"
              size="sm"
              variant="ghost"
              onClick={() => onEdit(config)}
            >
              <FiEdit2 />
            </IconButton>
            <IconButton
              aria-label="削除"
              size="sm"
              variant="ghost"
              colorPalette="red"
              onClick={() => onDelete(config)}
            >
              <FiTrash2 />
            </IconButton>
          </HStack>
        </Flex>
        <Tag.Root size="sm" mt={2} colorPalette="blue">
          <Tag.Label>
            {CHART_TYPE_LABELS[config.chart_type as ChartType] ||
              config.chart_type}
          </Tag.Label>
        </Tag.Root>
      </Card.Header>
      <Card.Body pt={0}>
        {isLoading ? (
          <Box
            h="150px"
            display="flex"
            alignItems="center"
            justifyContent="center"
          >
            <Text color="gray.400" fontSize="sm">
              読み込み中...
            </Text>
          </Box>
        ) : chartData ? (
          <Box h="150px">
            <ChartRenderer
              data={chartData}
              chartType={config.chart_type as ChartType}
              height={150}
            />
          </Box>
        ) : (
          <Box
            h="150px"
            display="flex"
            alignItems="center"
            justifyContent="center"
          >
            <Text color="gray.400" fontSize="sm">
              データなし
            </Text>
          </Box>
        )}
      </Card.Body>
    </Card.Root>
  );
}
