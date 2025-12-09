import { ChartConfigForm, ChartRenderer } from "@/components/charts";
import { Loading } from "@/components/common";
import {
  useApp,
  useChartConfigs,
  useChartData,
  useDeleteChartConfig,
  useFields,
  useSaveChartConfig,
} from "@/hooks";
import { ChartConfig, ChartDataRequest, ChartType, CHART_TYPE_LABELS } from "@/types";
import { ChevronRightIcon } from "@chakra-ui/icons";
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
  CardHeader,
  Flex,
  FormControl,
  FormLabel,
  Heading,
  HStack,
  IconButton,
  Input,
  SimpleGrid,
  Tag,
  Text,
  useDisclosure,
  useToast,
  VStack,
} from "@chakra-ui/react";
import { useRef, useState } from "react";
import { FiEdit2, FiPlus, FiTrash2 } from "react-icons/fi";
import { Link as RouterLink, useParams } from "react-router-dom";

export function AppChartPage() {
  const { appId } = useParams<{ appId: string }>();
  const numericAppId = Number(appId);
  const toast = useToast();
  const cancelRef = useRef<HTMLButtonElement>(null);

  // 編集モード管理
  const [isCreating, setIsCreating] = useState(false);
  const [editingConfig, setEditingConfig] = useState<ChartConfig | null>(null);
  const [chartName, setChartName] = useState("");
  const [chartConfig, setChartConfig] = useState<ChartDataRequest | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ChartConfig | null>(null);

  // 削除確認ダイアログ
  const {
    isOpen: isDeleteOpen,
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
      toast({
        title: "グラフ名を入力してください",
        status: "warning",
        duration: 3000,
      });
      return;
    }
    if (!chartConfig?.x_axis?.field) {
      toast({
        title: "X軸フィールドを選択してください",
        status: "warning",
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
      toast({
        title: editingConfig ? "グラフ設定を更新しました" : "グラフ設定を保存しました",
        status: "success",
        duration: 3000,
      });
      handleCancel();
    } catch {
      toast({
        title: "保存に失敗しました",
        status: "error",
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
      toast({
        title: "グラフ設定を削除しました",
        status: "success",
        duration: 3000,
      });
      onDeleteClose();
      setDeleteTarget(null);
    } catch {
      toast({
        title: "削除に失敗しました",
        status: "error",
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
        <BreadcrumbItem>
          <BreadcrumbLink as={RouterLink} to={`/apps/${appId}/records`}>
            {app.name}
          </BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbItem isCurrentPage>
          <BreadcrumbLink>グラフ設定</BreadcrumbLink>
        </BreadcrumbItem>
      </Breadcrumb>

      <Flex justify="space-between" align="center" mb={6}>
        <Heading size="lg">{app.name} - グラフ設定</Heading>
        {!isEditing && (
          <Button
            leftIcon={<FiPlus />}
            colorScheme="brand"
            onClick={handleStartCreate}
          >
            新規グラフ
          </Button>
        )}
      </Flex>

      {isEditing ? (
        // 編集フォーム
        <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6}>
          <Card>
            <CardHeader>
              <Heading size="md">
                {editingConfig ? "グラフ設定を編集" : "新規グラフ設定"}
              </Heading>
            </CardHeader>
            <CardBody>
              <VStack spacing={4} align="stretch">
                <FormControl isRequired>
                  <FormLabel>グラフ名</FormLabel>
                  <Input
                    value={chartName}
                    onChange={(e) => setChartName(e.target.value)}
                    placeholder="例: 月別売上推移"
                  />
                </FormControl>
                <ChartConfigForm
                  fields={fields}
                  onSubmit={setChartConfig}
                  initialConfig={chartConfig || editingConfig?.config || undefined}
                  isLoading={isChartLoading}
                />
              </VStack>
            </CardBody>
          </Card>

          <Card>
            <CardHeader>
              <Heading size="md">プレビュー</Heading>
            </CardHeader>
            <CardBody>
              {isChartLoading ? (
                <Loading message="グラフを生成中..." />
              ) : chartData ? (
                <Box>
                  <ChartRenderer
                    data={chartData}
                    chartType={chartConfig?.chart_type || "bar"}
                    height={350}
                  />
                  <HStack mt={4} justify="flex-end" spacing={3}>
                    <Button variant="ghost" onClick={handleCancel}>
                      キャンセル
                    </Button>
                    <Button
                      colorScheme="brand"
                      onClick={handleSave}
                      isLoading={saveChartConfigMutation.isPending}
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
            </CardBody>
          </Card>
        </SimpleGrid>
      ) : (
        // 保存済みグラフ一覧
        <Box>
          {savedConfigs.length === 0 ? (
            <Card>
              <CardBody>
                <VStack py={10} spacing={4}>
                  <Text color="gray.500">保存済みのグラフ設定がありません</Text>
                  <Button
                    leftIcon={<FiPlus />}
                    colorScheme="brand"
                    onClick={handleStartCreate}
                  >
                    最初のグラフを作成
                  </Button>
                </VStack>
              </CardBody>
            </Card>
          ) : (
            <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} spacing={4}>
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
      <AlertDialog
        isOpen={isDeleteOpen}
        leastDestructiveRef={cancelRef}
        onClose={onDeleteClose}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader>グラフ設定の削除</AlertDialogHeader>
            <AlertDialogBody>
              「{deleteTarget?.name}」を削除しますか？この操作は取り消せません。
            </AlertDialogBody>
            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={onDeleteClose}>
                キャンセル
              </Button>
              <Button
                colorScheme="red"
                onClick={handleDelete}
                ml={3}
                isLoading={deleteChartConfigMutation.isPending}
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

// 保存済みグラフカードコンポーネント
interface SavedChartCardProps {
  config: ChartConfig;
  appId: number;
  onEdit: (config: ChartConfig) => void;
  onDelete: (config: ChartConfig) => void;
}

function SavedChartCard({ config, appId, onEdit, onDelete }: SavedChartCardProps) {
  const { data: chartData, isLoading } = useChartData(appId, config.config);

  return (
    <Card>
      <CardHeader pb={2}>
        <Flex justify="space-between" align="center">
          <Heading size="sm" noOfLines={1}>
            {config.name}
          </Heading>
          <HStack spacing={1}>
            <IconButton
              aria-label="編集"
              icon={<FiEdit2 />}
              size="sm"
              variant="ghost"
              onClick={() => onEdit(config)}
            />
            <IconButton
              aria-label="削除"
              icon={<FiTrash2 />}
              size="sm"
              variant="ghost"
              colorScheme="red"
              onClick={() => onDelete(config)}
            />
          </HStack>
        </Flex>
        <Tag size="sm" mt={2} colorScheme="blue">
          {CHART_TYPE_LABELS[config.chart_type as ChartType] || config.chart_type}
        </Tag>
      </CardHeader>
      <CardBody pt={0}>
        {isLoading ? (
          <Box h="150px" display="flex" alignItems="center" justifyContent="center">
            <Text color="gray.400" fontSize="sm">読み込み中...</Text>
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
          <Box h="150px" display="flex" alignItems="center" justifyContent="center">
            <Text color="gray.400" fontSize="sm">データなし</Text>
          </Box>
        )}
      </CardBody>
    </Card>
  );
}
