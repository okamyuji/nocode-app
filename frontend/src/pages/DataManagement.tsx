import { useApps, useAuth } from "@/hooks";
import type { App } from "@/types";
import {
  Badge,
  Box,
  Card,
  CardBody,
  Flex,
  Grid,
  Heading,
  Icon,
  IconButton,
  Input,
  InputGroup,
  InputLeftElement,
  Skeleton,
  SkeletonText,
  Text,
  Tooltip,
  VStack,
} from "@chakra-ui/react";
import { useMemo, useState } from "react";
import {
  FiCalendar,
  FiDatabase,
  FiGrid,
  FiList,
  FiSearch,
  FiSettings,
} from "react-icons/fi";
import { useNavigate } from "react-router-dom";

export function DataManagementPage() {
  const navigate = useNavigate();
  const { isAdmin } = useAuth();
  const { data, isLoading } = useApps(1, 100);
  const [searchQuery, setSearchQuery] = useState("");

  const filteredApps = useMemo(() => {
    if (!data?.apps) return [];
    if (!searchQuery.trim()) return data.apps;

    const query = searchQuery.toLowerCase();
    return data.apps.filter(
      (app) =>
        app.name.toLowerCase().includes(query) ||
        app.description?.toLowerCase().includes(query)
    );
  }, [data?.apps, searchQuery]);

  const handleAppClick = (app: App) => {
    navigate(`/apps/${app.id}/records`);
  };

  const handleSettingsClick = (e: React.MouseEvent, app: App) => {
    e.stopPropagation();
    navigate(`/settings?tab=apps&appId=${app.id}`);
  };

  return (
    <Box p={6}>
      <VStack spacing={6} align="stretch">
        <Flex justify="space-between" align="center">
          <Box>
            <Heading size="lg" mb={2}>
              データ管理
            </Heading>
            <Text color="gray.600">アプリを選択してレコードを管理します</Text>
          </Box>
        </Flex>

        <InputGroup maxW="400px">
          <InputLeftElement>
            <Icon as={FiSearch} color="gray.400" />
          </InputLeftElement>
          <Input
            placeholder="アプリを検索..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </InputGroup>

        {isLoading ? (
          <Grid templateColumns="repeat(auto-fill, minmax(300px, 1fr))" gap={4}>
            {[1, 2, 3, 4].map((i) => (
              <Card key={i}>
                <CardBody>
                  <Skeleton height="20px" width="60%" mb={2} />
                  <SkeletonText noOfLines={2} />
                </CardBody>
              </Card>
            ))}
          </Grid>
        ) : filteredApps.length === 0 ? (
          <Card>
            <CardBody textAlign="center" py={10}>
              <Icon as={FiDatabase} boxSize={12} color="gray.300" mb={4} />
              <Text fontSize="lg" color="gray.500">
                {searchQuery
                  ? "検索条件に一致するアプリがありません"
                  : "アプリがありません"}
              </Text>
              {!searchQuery && (
                <Text color="gray.400" mt={2}>
                  「アプリ一覧」から新しいアプリを作成してください
                </Text>
              )}
            </CardBody>
          </Card>
        ) : (
          <Grid templateColumns="repeat(auto-fill, minmax(300px, 1fr))" gap={4}>
            {filteredApps.map((app) => (
              <AppDataCard
                key={app.id}
                app={app}
                onClick={() => handleAppClick(app)}
                onSettingsClick={(e) => handleSettingsClick(e, app)}
                isAdmin={isAdmin}
              />
            ))}
          </Grid>
        )}
      </VStack>
    </Box>
  );
}

interface AppDataCardProps {
  app: App;
  onClick: () => void;
  onSettingsClick: (e: React.MouseEvent) => void;
  isAdmin: boolean;
}

function AppDataCard({
  app,
  onClick,
  onSettingsClick,
  isAdmin,
}: AppDataCardProps) {
  const getIconForApp = (iconName: string | undefined) => {
    switch (iconName) {
      case "grid":
        return FiGrid;
      case "list":
        return FiList;
      case "calendar":
        return FiCalendar;
      default:
        return FiDatabase;
    }
  };

  return (
    <Card
      cursor="pointer"
      _hover={{
        transform: "translateY(-2px)",
        shadow: "md",
        borderColor: "brand.300",
      }}
      transition="all 0.2s"
      onClick={onClick}
      borderWidth="1px"
      borderColor="gray.200"
    >
      <CardBody>
        <Flex align="start" gap={3}>
          <Flex
            w={10}
            h={10}
            bg="brand.50"
            borderRadius="lg"
            align="center"
            justify="center"
            flexShrink={0}
          >
            <Icon as={getIconForApp(app.icon)} color="brand.500" boxSize={5} />
          </Flex>
          <Box flex={1} minW={0}>
            <Flex align="center" justify="space-between" mb={1}>
              <Flex align="center" gap={2} flex={1} minW={0}>
                <Heading size="sm" noOfLines={1}>
                  {app.name}
                </Heading>
                <Badge colorScheme="brand" fontSize="xs">
                  {app.field_count} フィールド
                </Badge>
              </Flex>
              {isAdmin && (
                <Tooltip label="アプリ設定" placement="top">
                  <IconButton
                    aria-label="アプリ設定"
                    icon={<FiSettings />}
                    size="sm"
                    variant="ghost"
                    onClick={onSettingsClick}
                    _hover={{ bg: "gray.100" }}
                  />
                </Tooltip>
              )}
            </Flex>
            <Text fontSize="sm" color="gray.600" noOfLines={2}>
              {app.description || "No description"}
            </Text>
            <Text fontSize="xs" color="gray.400" mt={2}>
              更新: {new Date(app.updated_at).toLocaleDateString("ja-JP")}
            </Text>
          </Box>
        </Flex>
      </CardBody>
    </Card>
  );
}
