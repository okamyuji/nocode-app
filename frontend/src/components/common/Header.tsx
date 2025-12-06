/**
 * ヘッダーコンポーネント
 * アプリケーションのナビゲーションバーとユーザーメニュー
 */

import { useAuth } from "@/hooks";
import { AddIcon, CloseIcon, HamburgerIcon } from "@chakra-ui/icons";
import {
  Avatar,
  Box,
  Button,
  Flex,
  HStack,
  IconButton,
  Menu,
  MenuButton,
  MenuDivider,
  MenuItem,
  MenuList,
  Stack,
  Text,
  useDisclosure,
} from "@chakra-ui/react";
import { Link as RouterLink, useNavigate } from "react-router-dom";

export function Header() {
  const { isOpen, onOpen, onClose } = useDisclosure();
  const { user, logout, isAuthenticated, isAdmin } = useAuth();
  const navigate = useNavigate();

  return (
    <Box bg="brand.500" px={4} position="sticky" top={0} zIndex={100}>
      <Flex h={16} alignItems="center" justifyContent="space-between">
        {/* モバイルメニューボタン */}
        <IconButton
          size="md"
          icon={isOpen ? <CloseIcon /> : <HamburgerIcon />}
          aria-label="メニュー"
          display={{ md: "none" }}
          onClick={isOpen ? onClose : onOpen}
          variant="ghost"
          color="white"
          _hover={{ bg: "brand.600" }}
        />

        {/* ロゴとナビゲーション */}
        <HStack spacing={8} alignItems="center">
          <Box
            as={RouterLink}
            to="/"
            fontWeight="bold"
            fontSize="xl"
            color="white"
            _hover={{ textDecoration: "none" }}
          >
            Nocode App
          </Box>

          {isAuthenticated && (
            <HStack as="nav" spacing={4} display={{ base: "none", md: "flex" }}>
              <Button
                as={RouterLink}
                to="/"
                variant="ghost"
                color="white"
                _hover={{ bg: "brand.600" }}
              >
                ダッシュボード
              </Button>
              <Button
                as={RouterLink}
                to="/apps"
                variant="ghost"
                color="white"
                _hover={{ bg: "brand.600" }}
              >
                アプリ一覧
              </Button>
            </HStack>
          )}
        </HStack>

        {/* ユーザーメニュー */}
        {isAuthenticated ? (
          <Flex alignItems="center">
            {isAdmin && (
              <Button
                leftIcon={<AddIcon />}
                colorScheme="whiteAlpha"
                variant="outline"
                size="sm"
                mr={4}
                display={{ base: "none", md: "flex" }}
                onClick={() => navigate("/apps/new")}
              >
                新規アプリ
              </Button>
            )}

            <Menu>
              <MenuButton
                as={Button}
                rounded="full"
                variant="link"
                cursor="pointer"
                minW={0}
              >
                <Avatar
                  size="sm"
                  name={user?.name}
                  bg="white"
                  color="brand.500"
                />
              </MenuButton>
              <MenuList>
                <Box px={3} py={2}>
                  <Text fontWeight="bold">{user?.name}</Text>
                  <Text fontSize="sm" color="gray.500">
                    {user?.email}
                  </Text>
                </Box>
                <MenuDivider />
                <MenuItem onClick={logout}>ログアウト</MenuItem>
              </MenuList>
            </Menu>
          </Flex>
        ) : (
          <HStack spacing={2}>
            <Button
              as={RouterLink}
              to="/login"
              variant="ghost"
              color="white"
              _hover={{ bg: "brand.600" }}
            >
              ログイン
            </Button>
            <Button
              as={RouterLink}
              to="/register"
              colorScheme="whiteAlpha"
              variant="outline"
            >
              登録
            </Button>
          </HStack>
        )}
      </Flex>

      {/* モバイルナビゲーション */}
      {isOpen && (
        <Box pb={4} display={{ md: "none" }}>
          <Stack as="nav" spacing={4}>
            {isAuthenticated && (
              <>
                <Button
                  as={RouterLink}
                  to="/"
                  variant="ghost"
                  color="white"
                  justifyContent="flex-start"
                >
                  ダッシュボード
                </Button>
                <Button
                  as={RouterLink}
                  to="/apps"
                  variant="ghost"
                  color="white"
                  justifyContent="flex-start"
                >
                  アプリ一覧
                </Button>
                {isAdmin && (
                  <Button
                    as={RouterLink}
                    to="/apps/new"
                    variant="ghost"
                    color="white"
                    justifyContent="flex-start"
                  >
                    新規アプリ作成
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
