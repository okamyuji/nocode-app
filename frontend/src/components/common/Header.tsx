/**
 * ヘッダーコンポーネント
 * アプリケーションのナビゲーションバーとユーザーメニュー
 */

import { useAuth } from "@/hooks";
import { AddIcon } from "@chakra-ui/icons/Add";
import { CloseIcon } from "@chakra-ui/icons/Close";
import { HamburgerIcon } from "@chakra-ui/icons/Hamburger";
import {
  Avatar,
  Box,
  Button,
  Flex,
  HStack,
  IconButton,
  Menu,
  Portal,
  Stack,
  Text,
  useDisclosure,
} from "@chakra-ui/react";
import { Link as RouterLink, useNavigate } from "react-router-dom";

export function Header() {
  const { open: isOpen, onOpen, onClose } = useDisclosure();
  const { user, logout, isAuthenticated, isAdmin } = useAuth();
  const navigate = useNavigate();

  return (
    <Box bg="brand.500" px={4} position="sticky" top={0} zIndex={100}>
      <Flex h={16} alignItems="center" justifyContent="space-between">
        {/* モバイルメニューボタン */}
        <IconButton
          size="md"
          aria-label="メニュー"
          display={{ md: "none" }}
          onClick={isOpen ? onClose : onOpen}
          variant="ghost"
          color="white"
          _hover={{ bg: "brand.600" }}
        >
          {isOpen ? <CloseIcon /> : <HamburgerIcon />}
        </IconButton>

        {/* ロゴとナビゲーション */}
        <HStack gap={8} alignItems="center">
          <Box
            asChild
            fontWeight="bold"
            fontSize="xl"
            color="white"
            _hover={{ textDecoration: "none" }}
          >
            <RouterLink to="/">Nocode App</RouterLink>
          </Box>

          {isAuthenticated && (
            <HStack as="nav" gap={4} display={{ base: "none", md: "flex" }}>
              <Button
                asChild
                variant="ghost"
                color="white"
                _hover={{ bg: "brand.600" }}
              >
                <RouterLink to="/">ダッシュボード</RouterLink>
              </Button>
              <Button
                asChild
                variant="ghost"
                color="white"
                _hover={{ bg: "brand.600" }}
              >
                <RouterLink to="/apps">アプリ一覧</RouterLink>
              </Button>
            </HStack>
          )}
        </HStack>

        {/* ユーザーメニュー */}
        {isAuthenticated ? (
          <Flex alignItems="center">
            {isAdmin && (
              <Button
                variant="outline"
                color="white"
                borderColor="whiteAlpha.500"
                _hover={{ bg: "whiteAlpha.200" }}
                size="sm"
                mr={4}
                display={{ base: "none", md: "flex" }}
                onClick={() => navigate("/apps/new")}
              >
                <AddIcon />
                新規アプリ
              </Button>
            )}

            <Menu.Root>
              <Menu.Trigger asChild>
                <Button
                  rounded="full"
                  variant="plain"
                  cursor="pointer"
                  minW={0}
                >
                  <Avatar.Root size="sm" bg="white" color="brand.500">
                    <Avatar.Fallback name={user?.name} />
                  </Avatar.Root>
                </Button>
              </Menu.Trigger>
              <Portal>
                <Menu.Positioner>
                  <Menu.Content>
                    <Box px={3} py={2}>
                      <Text fontWeight="bold">{user?.name}</Text>
                      <Text fontSize="sm" color="gray.500">
                        {user?.email}
                      </Text>
                    </Box>
                    <Menu.Separator />
                    <Menu.Item value="logout" onClick={logout}>
                      ログアウト
                    </Menu.Item>
                  </Menu.Content>
                </Menu.Positioner>
              </Portal>
            </Menu.Root>
          </Flex>
        ) : (
          <HStack gap={2}>
            <Button
              asChild
              variant="ghost"
              color="white"
              _hover={{ bg: "brand.600" }}
            >
              <RouterLink to="/login">ログイン</RouterLink>
            </Button>
            <Button
              asChild
              variant="outline"
              color="white"
              borderColor="whiteAlpha.500"
              _hover={{ bg: "whiteAlpha.200" }}
            >
              <RouterLink to="/register">登録</RouterLink>
            </Button>
          </HStack>
        )}
      </Flex>

      {/* モバイルナビゲーション */}
      {isOpen && (
        <Box pb={4} display={{ md: "none" }}>
          <Stack as="nav" gap={4}>
            {isAuthenticated && (
              <>
                <Button
                  asChild
                  variant="ghost"
                  color="white"
                  justifyContent="flex-start"
                >
                  <RouterLink to="/">ダッシュボード</RouterLink>
                </Button>
                <Button
                  asChild
                  variant="ghost"
                  color="white"
                  justifyContent="flex-start"
                >
                  <RouterLink to="/apps">アプリ一覧</RouterLink>
                </Button>
                {isAdmin && (
                  <Button
                    asChild
                    variant="ghost"
                    color="white"
                    justifyContent="flex-start"
                  >
                    <RouterLink to="/apps/new">新規アプリ作成</RouterLink>
                  </Button>
                )}
              </>
            )}
          </Stack>
        </Box>
      )}
    </Box>
  );
}
