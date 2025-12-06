import { Loading } from "@/components/common";
import { useApps } from "@/hooks";
import { AddIcon } from "@chakra-ui/icons";
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
      <VStack spacing={4} py={10}>
        <Text color="gray.500">アプリがありません</Text>
        <Button
          as={RouterLink}
          to="/apps/new"
          leftIcon={<AddIcon />}
          colorScheme="brand"
        >
          最初のアプリを作成
        </Button>
      </VStack>
    );
  }

  return (
    <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} spacing={4}>
      {data.apps.map((app) => (
        <AppCard key={app.id} app={app} />
      ))}
    </SimpleGrid>
  );
}
