import { AppList } from "@/components/apps";
import { useAuth } from "@/hooks";
import { AddIcon, SearchIcon } from "@chakra-ui/icons";
import {
  Box,
  Button,
  Heading,
  HStack,
  Input,
  InputGroup,
  InputLeftElement,
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
          <Button
            as={RouterLink}
            to="/apps/new"
            leftIcon={<AddIcon />}
            colorScheme="brand"
          >
            新規作成
          </Button>
        )}
      </HStack>

      <Box mb={6}>
        <InputGroup maxW="400px">
          <InputLeftElement pointerEvents="none">
            <SearchIcon color="gray.400" />
          </InputLeftElement>
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
