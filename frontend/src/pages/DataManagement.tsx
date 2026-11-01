import { useApps, useAuth } from "@/hooks";
import type { App } from "@/types";
import { getAppIcon } from "@/utils";
import {
  Badge,
  Box,
  Card,
  Flex,
  Grid,
  Heading,
  Icon,
  IconButton,
  Input,
  InputGroup,
  Skeleton,
  SkeletonText,
  Text,
  VStack,
} from "@chakra-ui/react";
import { Tooltip } from "@/components/ui/tooltip";
import { useMemo, useState } from "react";
import { FiDatabase, FiSearch, FiSettings } from "react-icons/fi";
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
      <VStack gap={6} align="stretch">
        <Flex justify="space-between" align="center">
          <Box>
            <Heading size="lg" mb={2}>
              データ管理
            </Heading>
            <Text color="gray.600">アプリを選択してレコードを管理します</Text>
          </Box>
        </Flex>

        <InputGroup
          maxW="400px"
          startElement={<Icon as={FiSearch} color="gray.400" />}
        >
          <Input
            placeholder="アプリを検索..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </InputGroup>

        {isLoading ? (
          <Grid templateColumns="repeat(auto-fill, minmax(300px, 1fr))" gap={4}>
            {[1, 2, 3, 4].map((i) => (
              <Card.Root key={i}>
                <Card.Body>
                  <Skeleton height="20px" width="60%" mb={2} />
                  <SkeletonText noOfLines={2} />
                </Card.Body>
              </Card.Root>
            ))}
          </Grid>
        ) : filteredApps.length === 0 ? (
          <Card.Root>
            <Card.Body textAlign="center" py={10}>
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
            </Card.Body>
          </Card.Root>
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
  return (
    <Card.Root
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
      <Card.Body>
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
            <Icon as={getAppIcon(app.icon)} color="brand.500" boxSize={5} />
          </Flex>
          <Box flex={1} minW={0}>
            <Flex align="center" justify="space-between" mb={1}>
              <Flex align="center" gap={2} flex={1} minW={0}>
                <Heading size="sm" lineClamp={1}>
                  {app.name}
                </Heading>
                <Badge colorPalette="brand" fontSize="xs">
                  {app.field_count} フィールド
                </Badge>
              </Flex>
              {isAdmin && (
                <Tooltip
                  content="アプリ設定"
                  positioning={{ placement: "top" }}
                >
                  <IconButton
                    aria-label="アプリ設定"
                    size="sm"
                    variant="ghost"
                    onClick={onSettingsClick}
                    _hover={{ bg: "gray.100" }}
                  >
                    <FiSettings />
                  </IconButton>
                </Tooltip>
              )}
            </Flex>
            <Text fontSize="sm" color="gray.600" lineClamp={2}>
              {app.description || "No description"}
            </Text>
            <Text fontSize="xs" color="gray.400" mt={2}>
              更新: {new Date(app.updated_at).toLocaleDateString("ja-JP")}
            </Text>
          </Box>
        </Flex>
      </Card.Body>
    </Card.Root>
  );
}
