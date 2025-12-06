import { useDashboardApi } from "@/api";
import { useApps } from "@/hooks";
import {
  Box,
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
  VStack,
} from "@chakra-ui/react";
import { useQuery } from "@tanstack/react-query";
import { FiActivity, FiDatabase, FiGrid, FiUsers } from "react-icons/fi";
import { Link as RouterLink } from "react-router-dom";

export function DashboardPage() {
  const { data: appsData } = useApps();
  const dashboardApi = useDashboardApi();

  const { data: statsData, isLoading: statsLoading } = useQuery({
    queryKey: ["dashboard", "stats"],
    queryFn: () => dashboardApi.getStats(),
    staleTime: 30000, // 30 seconds
  });

  const stats = [
    {
      label: "アプリ数",
      value: statsLoading
        ? null
        : (statsData?.stats.app_count ?? appsData?.apps.length ?? 0),
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

  return (
    <Box>
      <Heading size="lg" mb={6}>
        ダッシュボード
      </Heading>

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

      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6}>
        <Card>
          <CardBody>
            <Heading size="md" mb={4}>
              最近のアプリ
            </Heading>
            {appsData?.apps.length === 0 ? (
              <Text color="gray.500">アプリがありません</Text>
            ) : (
              <VStack align="stretch" spacing={2}>
                {appsData?.apps.slice(0, 5).map((app) => (
                  <HStack
                    key={app.id}
                    as={RouterLink}
                    to={`/apps/${app.id}/records`}
                    p={3}
                    borderRadius="md"
                    _hover={{ bg: "gray.50" }}
                    justify="space-between"
                  >
                    <HStack>
                      <Icon as={FiGrid} color="brand.500" />
                      <Text fontWeight="medium">{app.name}</Text>
                    </HStack>
                    <Text fontSize="sm" color="gray.500">
                      {new Date(app.updated_at).toLocaleDateString("ja-JP")}
                    </Text>
                  </HStack>
                ))}
              </VStack>
            )}
          </CardBody>
        </Card>

        <Card>
          <CardBody>
            <Heading size="md" mb={4}>
              アクティビティ
            </Heading>
            <Text color="gray.500">まだアクティビティがありません</Text>
          </CardBody>
        </Card>
      </SimpleGrid>
    </Box>
  );
}
