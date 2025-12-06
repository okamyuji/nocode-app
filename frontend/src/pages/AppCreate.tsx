import { AppFormBuilder } from "@/components/apps";
import { useAuth } from "@/hooks";
import { ChevronRightIcon } from "@chakra-ui/icons";
import {
  Alert,
  AlertDescription,
  AlertIcon,
  AlertTitle,
  Box,
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  Heading,
} from "@chakra-ui/react";
import { Navigate, Link as RouterLink } from "react-router-dom";

export function AppCreatePage() {
  const { isAdmin, isAuthenticated } = useAuth();

  // Redirect non-admin users
  if (isAuthenticated && !isAdmin) {
    return (
      <Box p={6}>
        <Alert status="warning" borderRadius="md">
          <AlertIcon />
          <Box>
            <AlertTitle>アクセス権限がありません</AlertTitle>
            <AlertDescription>
              アプリの作成は管理者のみ可能です。
            </AlertDescription>
          </Box>
        </Alert>
        <Box mt={4}>
          <Navigate to="/apps" replace />
        </Box>
      </Box>
    );
  }

  return (
    <Box>
      <Breadcrumb
        spacing="8px"
        separator={<ChevronRightIcon color="gray.500" />}
        mb={4}
      >
        <BreadcrumbItem>
          <BreadcrumbLink as={RouterLink} to="/apps">
            アプリ一覧
          </BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbItem isCurrentPage>
          <BreadcrumbLink>新規作成</BreadcrumbLink>
        </BreadcrumbItem>
      </Breadcrumb>

      <Heading size="lg" mb={6}>
        アプリを作成
      </Heading>

      <AppFormBuilder />
    </Box>
  );
}
