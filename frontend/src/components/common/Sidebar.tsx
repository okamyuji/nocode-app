/**
 * サイドバーコンポーネント
 * ナビゲーションリンクとリサイズ機能を持つサイドバー
 */

import { useAuth } from "@/hooks";
import { MAX_SIDEBAR_WIDTH, MIN_SIDEBAR_WIDTH, useUIStore } from "@/stores";
import {
  Box,
  Button,
  Divider,
  Icon,
  IconButton,
  Text,
  Tooltip,
  VStack,
} from "@chakra-ui/react";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  FiChevronLeft,
  FiChevronRight,
  FiDatabase,
  FiGrid,
  FiHome,
  FiPlus,
  FiSettings,
} from "react-icons/fi";
import { Link as RouterLink, useLocation } from "react-router-dom";

interface SidebarItemProps {
  to: string;
  icon: React.ElementType;
  children: React.ReactNode;
  isActive?: boolean;
  isCollapsed?: boolean;
}

/**
 * サイドバーのナビゲーションアイテム
 */
function SidebarItem({
  to,
  icon,
  children,
  isActive,
  isCollapsed,
}: SidebarItemProps) {
  const button = (
    <Button
      as={RouterLink}
      to={to}
      variant="ghost"
      justifyContent={isCollapsed ? "center" : "flex-start"}
      w="full"
      leftIcon={isCollapsed ? undefined : <Icon as={icon} />}
      bg={isActive ? "brand.50" : "transparent"}
      color={isActive ? "brand.600" : "gray.600"}
      _hover={{ bg: "brand.50", color: "brand.600" }}
      px={isCollapsed ? 2 : 3}
    >
      {isCollapsed ? <Icon as={icon} /> : children}
    </Button>
  );

  // 折りたたみ時はツールチップを表示
  if (isCollapsed) {
    return (
      <Tooltip label={children} placement="right" hasArrow>
        {button}
      </Tooltip>
    );
  }

  return button;
}

export function Sidebar() {
  const location = useLocation();
  const { isAdmin } = useAuth();
  const {
    sidebarWidth,
    sidebarCollapsed,
    setSidebarWidth,
    toggleSidebarCollapsed,
  } = useUIStore();
  const [isResizing, setIsResizing] = useState(false);
  const sidebarRef = useRef<HTMLDivElement>(null);

  /**
   * リサイズ開始
   */
  const startResizing = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsResizing(true);
  }, []);

  /**
   * リサイズ終了
   */
  const stopResizing = useCallback(() => {
    setIsResizing(false);
  }, []);

  /**
   * リサイズ処理
   */
  const resize = useCallback(
    (e: MouseEvent) => {
      if (isResizing && sidebarRef.current) {
        const newWidth =
          e.clientX - sidebarRef.current.getBoundingClientRect().left;
        if (newWidth >= MIN_SIDEBAR_WIDTH && newWidth <= MAX_SIDEBAR_WIDTH) {
          setSidebarWidth(newWidth);
        }
      }
    },
    [isResizing, setSidebarWidth]
  );

  // リサイズイベントリスナーの登録
  useEffect(() => {
    if (isResizing) {
      window.addEventListener("mousemove", resize);
      window.addEventListener("mouseup", stopResizing);
    }
    return () => {
      window.removeEventListener("mousemove", resize);
      window.removeEventListener("mouseup", stopResizing);
    };
  }, [isResizing, resize, stopResizing]);

  const effectiveWidth = sidebarCollapsed ? MIN_SIDEBAR_WIDTH : sidebarWidth;

  return (
    <Box
      ref={sidebarRef}
      as="aside"
      w={`${effectiveWidth}px`}
      minW={`${effectiveWidth}px`}
      bg="white"
      borderRight="1px"
      borderColor="gray.200"
      h="calc(100vh - 64px)"
      position="sticky"
      top="64px"
      display={{ base: "none", lg: "flex" }}
      flexDirection="column"
      transition={isResizing ? "none" : "width 0.2s"}
    >
      {/* サイドバーコンテンツ */}
      <VStack
        spacing={1}
        align="stretch"
        p={sidebarCollapsed ? 2 : 4}
        flex={1}
        overflow="hidden"
      >
        <SidebarItem
          to="/"
          icon={FiHome}
          isActive={location.pathname === "/"}
          isCollapsed={sidebarCollapsed}
        >
          ダッシュボード
        </SidebarItem>

        <Divider my={2} />

        {!sidebarCollapsed && (
          <Text fontSize="xs" fontWeight="bold" color="gray.500" px={3} py={2}>
            アプリ
          </Text>
        )}

        <SidebarItem
          to="/apps"
          icon={FiGrid}
          isActive={location.pathname === "/apps"}
          isCollapsed={sidebarCollapsed}
        >
          アプリ一覧
        </SidebarItem>

        {isAdmin && (
          <SidebarItem
            to="/apps/new"
            icon={FiPlus}
            isActive={location.pathname === "/apps/new"}
            isCollapsed={sidebarCollapsed}
          >
            新規作成
          </SidebarItem>
        )}

        <Divider my={2} />

        {!sidebarCollapsed && (
          <Text fontSize="xs" fontWeight="bold" color="gray.500" px={3} py={2}>
            管理
          </Text>
        )}

        <SidebarItem
          to="/settings"
          icon={FiSettings}
          isActive={location.pathname === "/settings"}
          isCollapsed={sidebarCollapsed}
        >
          設定
        </SidebarItem>

        <SidebarItem
          to="/data"
          icon={FiDatabase}
          isActive={location.pathname === "/data"}
          isCollapsed={sidebarCollapsed}
        >
          データ管理
        </SidebarItem>
      </VStack>

      {/* 折りたたみボタン */}
      <Box p={2} borderTop="1px" borderColor="gray.200">
        <Tooltip
          label={
            sidebarCollapsed ? "サイドバーを展開" : "サイドバーを折りたたむ"
          }
          placement="right"
        >
          <IconButton
            aria-label={sidebarCollapsed ? "展開" : "折りたたむ"}
            icon={
              <Icon as={sidebarCollapsed ? FiChevronRight : FiChevronLeft} />
            }
            variant="ghost"
            size="sm"
            w="full"
            onClick={toggleSidebarCollapsed}
          />
        </Tooltip>
      </Box>

      {/* リサイズハンドル */}
      {!sidebarCollapsed && (
        <Box
          position="absolute"
          right={0}
          top={0}
          bottom={0}
          w="4px"
          cursor="col-resize"
          bg={isResizing ? "brand.300" : "transparent"}
          _hover={{ bg: "brand.200" }}
          onMouseDown={startResizing}
          transition="background 0.2s"
        />
      )}
    </Box>
  );
}
