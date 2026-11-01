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
  CloseButton,
  Dialog,
  Field,
  Heading,
  HStack,
  Icon,
  NativeSelect,
  Portal,
  RadioGroup,
  SimpleGrid,
  Stack,
  Text,
  VStack,
} from "@chakra-ui/react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { FiCheck } from "react-icons/fi";
import { toaster } from "@/components/ui/toaster";

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
      toaster.create({
        title: "ウィジェットを追加しました",
        type: "success",
        duration: 3000,
      });
      handleClose();
    },
    onError: () => {
      toaster.create({
        title: "追加に失敗しました",
        type: "error",
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
      <HStack gap={3}>
        <Icon as={getAppIcon(app.icon)} boxSize={8} color="brand.500" />
        <VStack align="start" gap={0}>
          <HStack>
            <Text fontWeight="semibold">{app.name}</Text>
            {app.is_external && (
              <Badge colorPalette="purple" fontSize="xs">
                外部
              </Badge>
            )}
          </HStack>
          <Text fontSize="sm" color="gray.500" lineClamp={1}>
            {app.description || "説明なし"}
          </Text>
        </VStack>
      </HStack>
    </Box>
  );

  return (
    <Dialog.Root
      open={isOpen}
      onOpenChange={(e) => !e.open && handleClose()}
      size="xl"
    >
      <Portal>
        <Dialog.Backdrop />
        <Dialog.Positioner>
          <Dialog.Content maxH="80vh">
            <Dialog.Header>
              <Dialog.Title>ダッシュボードにウィジェットを追加</Dialog.Title>
            </Dialog.Header>
            <Dialog.CloseTrigger asChild>
              <CloseButton size="sm" />
            </Dialog.CloseTrigger>
            <Dialog.Body overflowY="auto">
              <VStack gap={6} align="stretch">
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
                    <SimpleGrid columns={{ base: 1, md: 2 }} gap={3}>
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
                    <Field.Root>
                      <Field.Label>表示形式</Field.Label>
                      <RadioGroup.Root
                        value={viewType}
                        onValueChange={(e) => e.value && setViewType(e.value)}
                      >
                        <Stack direction="row" gap={4}>
                          {[
                            { value: "table", label: "テーブル" },
                            { value: "list", label: "リスト" },
                            { value: "chart", label: "グラフ" },
                          ].map((option) => (
                            <RadioGroup.Item
                              key={option.value}
                              value={option.value}
                            >
                              <RadioGroup.ItemHiddenInput />
                              <RadioGroup.ItemIndicator />
                              <RadioGroup.ItemText>
                                {option.label}
                              </RadioGroup.ItemText>
                            </RadioGroup.Item>
                          ))}
                        </Stack>
                      </RadioGroup.Root>
                    </Field.Root>

                    <Field.Root>
                      <Field.Label>ウィジェットサイズ</Field.Label>
                      <NativeSelect.Root>
                        <NativeSelect.Field
                          value={widgetSize}
                          onChange={(e) => setWidgetSize(e.target.value)}
                        >
                          <option value="small">小</option>
                          <option value="medium">中</option>
                          <option value="large">大</option>
                        </NativeSelect.Field>
                        <NativeSelect.Indicator />
                      </NativeSelect.Root>
                    </Field.Root>
                  </>
                )}
              </VStack>
            </Dialog.Body>
            <Dialog.Footer>
              <Button variant="ghost" mr={3} onClick={handleClose}>
                キャンセル
              </Button>
              <Button
                colorPalette="brand"
                onClick={handleSubmit}
                loading={createMutation.isPending}
                disabled={!selectedAppId}
              >
                追加
              </Button>
            </Dialog.Footer>
          </Dialog.Content>
        </Dialog.Positioner>
      </Portal>
    </Dialog.Root>
  );
}
