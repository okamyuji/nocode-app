import { useDashboardApi } from "@/api";
import { Loading } from "@/components/common";
import { AddWidgetModal, DashboardWidgetGrid } from "@/components/dashboard";
import { useDashboardWidgets } from "@/hooks";
import { AddIcon } from "@chakra-ui/icons/Add";
import {
  Box,
  Button,
  Card,
  Heading,
  HStack,
  Icon,
  SimpleGrid,
  Spinner,
  Stat,
  Text,
  useDisclosure,
} from "@chakra-ui/react";
import { useQuery } from "@tanstack/react-query";
import { FiActivity, FiDatabase, FiGrid, FiUsers } from "react-icons/fi";

export function DashboardPage() {
  const dashboardApi = useDashboardApi();
  const { open: isOpen, onOpen, onClose } = useDisclosure();

  // 統計データの取得
  const { data: statsData, isLoading: statsLoading } = useQuery({
    queryKey: ["dashboard", "stats"],
    queryFn: () => dashboardApi.getStats(),
    staleTime: 30000, // 30 seconds
  });

  // ウィジェットデータの取得（全ウィジェット）
  const { data: widgetsData, isLoading: widgetsLoading } =
    useDashboardWidgets(false);

  const stats = [
    {
      label: "アプリ数",
      value: statsLoading ? null : (statsData?.stats.app_count ?? 0),
      icon: FiGrid,
      color: "brand.500",
    },
    {
      label: "総レコード数",
      value: statsLoading ? null : (statsData?.stats.total_records ?? 0),
      icon: FiDatabase,
      color: "green.500",
    },
    {
      label: "ユーザー数",
      value: statsLoading ? null : (statsData?.stats.user_count ?? 0),
      icon: FiUsers,
      color: "purple.500",
    },
    {
      label: "今日の更新",
      value: statsLoading ? null : (statsData?.stats.todays_updates ?? 0),
      icon: FiActivity,
      color: "orange.500",
    },
  ];

  const visibleWidgets = widgetsData?.widgets.filter((w) => w.is_visible) || [];

  return (
    <Box>
      <HStack justify="space-between" mb={6}>
        <Heading size="lg">ダッシュボード</Heading>
        <Button colorPalette="brand" onClick={onOpen}>
          <AddIcon />
          ウィジェットを追加
        </Button>
      </HStack>

      {/* 統計カード */}
      <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} gap={4} mb={8}>
        {stats.map((stat) => (
          <Card.Root key={stat.label}>
            <Card.Body>
              <Stat.Root>
                <HStack justify="space-between">
                  <Box>
                    <Stat.Label color="gray.500">{stat.label}</Stat.Label>
                    <Stat.ValueText fontSize="3xl">
                      {stat.value === null ? (
                        <Spinner size="sm" />
                      ) : (
                        stat.value.toLocaleString()
                      )}
                    </Stat.ValueText>
                    <Stat.HelpText mb={0}>前月比 -</Stat.HelpText>
                  </Box>
                  <Icon as={stat.icon} boxSize={10} color={stat.color} />
                </HStack>
              </Stat.Root>
            </Card.Body>
          </Card.Root>
        ))}
      </SimpleGrid>

      {/* ウィジェットグリッド */}
      <Box>
        <Heading size="md" mb={4}>
          アプリウィジェット
        </Heading>

        {widgetsLoading ? (
          <Loading message="ウィジェットを読み込み中..." />
        ) : visibleWidgets.length === 0 ? (
          <Card.Root>
            <Card.Body>
              <Text color="gray.500" textAlign="center" py={8}>
                表示中のウィジェットがありません。
                <br />
                「ウィジェットを追加」ボタンからアプリを追加してください。
              </Text>
            </Card.Body>
          </Card.Root>
        ) : (
          <DashboardWidgetGrid widgets={visibleWidgets} />
        )}
      </Box>

      {/* ウィジェット追加モーダル */}
      <AddWidgetModal
        isOpen={isOpen}
        onClose={onClose}
        existingWidgets={widgetsData?.widgets || []}
      />
    </Box>
  );
}
