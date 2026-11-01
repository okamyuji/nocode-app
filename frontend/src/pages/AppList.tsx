import { AppList } from "@/components/apps";
import { useAuth } from "@/hooks";
import { AddIcon } from "@chakra-ui/icons/Add";
import { SearchIcon } from "@chakra-ui/icons/Search";
import {
  Box,
  Button,
  Heading,
  HStack,
  Input,
  InputGroup,
} from "@chakra-ui/react";
import { useState } from "react";
import { Link as RouterLink } from "react-router-dom";

export function AppListPage() {
  const [searchTerm, setSearchTerm] = useState("");
  const { isAdmin } = useAuth();

  return (
    <Box>
      <HStack justify="space-between" mb={6}>
        <Heading size="lg">アプリ一覧</Heading>
        {isAdmin && (
          <Button asChild colorPalette="brand">
            <RouterLink to="/apps/new">
              <AddIcon />
              新規作成
            </RouterLink>
          </Button>
        )}
      </HStack>

      <Box mb={6}>
        <InputGroup maxW="400px" startElement={<SearchIcon color="gray.400" />}>
          <Input
            placeholder="アプリを検索..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            bg="white"
          />
        </InputGroup>
      </Box>

      <AppList />
    </Box>
  );
}
