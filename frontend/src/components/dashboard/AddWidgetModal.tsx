/**
 * ウィジェット追加モーダルコンポーネント
 * 利用可能なアプリからダッシュボードに追加するアプリを選択
 */

import { useDashboardWidgetsApi } from "@/api";
import { useApps } from "@/hooks";
import type {
  App,
  CreateDashboardWidgetRequest,
  DashboardWidget,
} from "@/types";
import { getAppIcon } from "@/utils";
import {
  Badge,
  Box,
  Button,
  FormControl,
  FormLabel,
  Heading,
  HStack,
  Icon,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Radio,
  RadioGroup,
  Select,
  SimpleGrid,
  Stack,
  Text,
  useToast,
  VStack,
} from "@chakra-ui/react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { FiCheck } from "react-icons/fi";

interface AddWidgetModalProps {
  isOpen: boolean;
  onClose: () => void;
  existingWidgets: DashboardWidget[];
}

export function AddWidgetModal({
  isOpen,
  onClose,
  existingWidgets,
}: AddWidgetModalProps) {
  const toast = useToast();
  const queryClient = useQueryClient();
  const dashboardWidgetsApi = useDashboardWidgetsApi();
  const { data: appsData, isLoading: isAppsLoading } = useApps();

  const [selectedAppId, setSelectedAppId] = useState<number | null>(null);
  const [viewType, setViewType] = useState<string>("table");
  const [widgetSize, setWidgetSize] = useState<string>("medium");

  // 既にウィジェットに追加されているアプリIDのセット
  const existingAppIds = new Set(existingWidgets.map((w) => w.app_id));

  // 追加可能なアプリ（まだウィジェットに追加されていないもの）
  const availableApps =
    appsData?.apps.filter((app) => !existingAppIds.has(app.id)) || [];

  const createMutation = useMutation({
    mutationFn: (data: CreateDashboardWidgetRequest) =>
      dashboardWidgetsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dashboard", "widgets"] });
      toast({
        title: "ウィジェットを追加しました",
        status: "success",
        duration: 3000,
      });
      handleClose();
    },
    onError: () => {
      toast({
        title: "追加に失敗しました",
        status: "error",
        duration: 3000,
      });
    },
  });

  const handleClose = () => {
    setSelectedAppId(null);
    setViewType("table");
    setWidgetSize("medium");
    onClose();
  };

  const handleSubmit = () => {
    if (!selectedAppId) return;

    createMutation.mutate({
      app_id: selectedAppId,
      view_type: viewType as "table" | "list" | "chart",
      widget_size: widgetSize as "small" | "medium" | "large",
      is_visible: true,
    });
  };

  const AppCard = ({ app, isSelected }: { app: App; isSelected: boolean }) => (
    <Box
      p={4}
      borderWidth="2px"
      borderRadius="md"
      borderColor={isSelected ? "brand.500" : "gray.200"}
      bg={isSelected ? "brand.50" : "white"}
      cursor="pointer"
      transition="all 0.2s"
      _hover={{
        borderColor: "brand.400",
        transform: "translateY(-2px)",
        shadow: "sm",
      }}
      onClick={() => setSelectedAppId(app.id)}
      position="relative"
    >
      {isSelected && (
        <Icon
          as={FiCheck}
          position="absolute"
          top={2}
          right={2}
          color="brand.500"
          boxSize={5}
        />
      )}
      <HStack spacing={3}>
        <Icon as={getAppIcon(app.icon)} boxSize={8} color="brand.500" />
        <VStack align="start" spacing={0}>
          <HStack>
            <Text fontWeight="semibold">{app.name}</Text>
            {app.is_external && (
              <Badge colorScheme="purple" fontSize="xs">
                外部
              </Badge>
            )}
          </HStack>
          <Text fontSize="sm" color="gray.500" noOfLines={1}>
            {app.description || "説明なし"}
          </Text>
        </VStack>
      </HStack>
    </Box>
  );

  return (
    <Modal isOpen={isOpen} onClose={handleClose} size="xl">
      <ModalOverlay />
      <ModalContent maxH="80vh">
        <ModalHeader>ダッシュボードにウィジェットを追加</ModalHeader>
        <ModalCloseButton />
        <ModalBody overflowY="auto">
          <VStack spacing={6} align="stretch">
            {/* アプリ選択 */}
            <Box>
              <Heading size="sm" mb={3}>
                アプリを選択
              </Heading>
              {isAppsLoading ? (
                <Text color="gray.500">読み込み中...</Text>
              ) : availableApps.length === 0 ? (
                <Text color="gray.500">
                  追加可能なアプリがありません。すべてのアプリが既にダッシュボードに追加されています。
                </Text>
              ) : (
                <SimpleGrid columns={{ base: 1, md: 2 }} spacing={3}>
                  {availableApps.map((app) => (
                    <AppCard
                      key={app.id}
                      app={app}
                      isSelected={selectedAppId === app.id}
                    />
                  ))}
                </SimpleGrid>
              )}
            </Box>

            {/* 表示形式選択 */}
            {selectedAppId && (
              <>
                <FormControl>
                  <FormLabel>表示形式</FormLabel>
                  <RadioGroup value={viewType} onChange={setViewType}>
                    <Stack direction="row" spacing={4}>
                      <Radio value="table">テーブル</Radio>
                      <Radio value="list">リスト</Radio>
                      <Radio value="chart">グラフ</Radio>
                    </Stack>
                  </RadioGroup>
                </FormControl>

                <FormControl>
                  <FormLabel>ウィジェットサイズ</FormLabel>
                  <Select
                    value={widgetSize}
                    onChange={(e) => setWidgetSize(e.target.value)}
                  >
                    <option value="small">小</option>
                    <option value="medium">中</option>
                    <option value="large">大</option>
                  </Select>
                </FormControl>
              </>
            )}
          </VStack>
        </ModalBody>
        <ModalFooter>
          <Button variant="ghost" mr={3} onClick={handleClose}>
            キャンセル
          </Button>
          <Button
            colorScheme="brand"
            onClick={handleSubmit}
            isLoading={createMutation.isPending}
            isDisabled={!selectedAppId}
          >
            追加
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
}
