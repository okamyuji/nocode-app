import { useDashboardApi } from "@/api";
import { Loading } from "@/components/common";
import { AddWidgetModal, DashboardWidgetGrid } from "@/components/dashboard";
import { useDashboardWidgets } from "@/hooks";
import { AddIcon } from "@chakra-ui/icons";
import {
  Box,
  Button,
  Card,
  CardBody,
  Heading,
  HStack,
  Icon,
  SimpleGrid,
  Spinner,
  Stat,
  StatHelpText,
  StatLabel,
  StatNumber,
  Text,
  useDisclosure,
} from "@chakra-ui/react";
import { useQuery } from "@tanstack/react-query";
import { FiActivity, FiDatabase, FiGrid, FiUsers } from "react-icons/fi";

export function DashboardPage() {
  const dashboardApi = useDashboardApi();
  const { isOpen, onOpen, onClose } = useDisclosure();

  // 統計データの取得
  const { data: statsData, isLoading: statsLoading } = useQuery({
    queryKey: ["dashboard", "stats"],
    queryFn: () => dashboardApi.getStats(),
    staleTime: 30000, // 30 seconds
  });

  // ウィジェットデータの取得（表示中のもののみ）
  const { data: widgetsData, isLoading: widgetsLoading } =
    useDashboardWidgets(false);

  // 全ウィジェットデータの取得（追加モーダル用）
  const { data: allWidgetsData } = useDashboardWidgets(false);

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
        <Button leftIcon={<AddIcon />} colorScheme="brand" onClick={onOpen}>
          ウィジェットを追加
        </Button>
      </HStack>

      {/* 統計カード */}
      <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4} mb={8}>
        {stats.map((stat) => (
          <Card key={stat.label}>
            <CardBody>
              <Stat>
                <HStack justify="space-between">
                  <Box>
                    <StatLabel color="gray.500">{stat.label}</StatLabel>
                    <StatNumber fontSize="3xl">
                      {stat.value === null ? (
                        <Spinner size="sm" />
                      ) : (
                        stat.value.toLocaleString()
                      )}
                    </StatNumber>
                    <StatHelpText mb={0}>前月比 -</StatHelpText>
                  </Box>
                  <Icon as={stat.icon} boxSize={10} color={stat.color} />
                </HStack>
              </Stat>
            </CardBody>
          </Card>
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
          <Card>
            <CardBody>
              <Text color="gray.500" textAlign="center" py={8}>
                表示中のウィジェットがありません。
                <br />
                「ウィジェットを追加」ボタンからアプリを追加してください。
              </Text>
            </CardBody>
          </Card>
        ) : (
          <DashboardWidgetGrid widgets={visibleWidgets} />
        )}
      </Box>

      {/* ウィジェット追加モーダル */}
      <AddWidgetModal
        isOpen={isOpen}
        onClose={onClose}
        existingWidgets={allWidgetsData?.widgets || []}
      />
    </Box>
  );
}
