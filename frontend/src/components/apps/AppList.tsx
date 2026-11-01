import { Loading } from "@/components/common";
import { useApps } from "@/hooks";
import { AddIcon } from "@chakra-ui/icons/Add";
import { Box, Button, SimpleGrid, Text, VStack } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import { AppCard } from "./AppCard";

interface AppListProps {
  page?: number;
  limit?: number;
}

export function AppList({ page = 1, limit = 20 }: AppListProps) {
  const { data, isLoading, error } = useApps(page, limit);

  if (isLoading) {
    return <Loading message="アプリを読み込み中..." />;
  }

  if (error) {
    return (
      <Box textAlign="center" py={10}>
        <Text color="red.500">アプリの読み込みに失敗しました</Text>
      </Box>
    );
  }

  if (!data?.apps.length) {
    return (
      <VStack gap={4} py={10}>
        <Text color="gray.500">アプリがありません</Text>
        <Button asChild colorPalette="brand">
          <RouterLink to="/apps/new">
            <AddIcon />
            最初のアプリを作成
          </RouterLink>
        </Button>
      </VStack>
    );
  }

  return (
    <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} gap={4}>
      {data.apps.map((app) => (
        <AppCard key={app.id} app={app} />
      ))}
    </SimpleGrid>
  );
}
