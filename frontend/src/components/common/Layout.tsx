import { MIN_SIDEBAR_WIDTH, useUIStore } from "@/stores";
import { Box, Flex } from "@chakra-ui/react";
import { Outlet } from "react-router-dom";
import { Header } from "./Header";
import { Sidebar } from "./Sidebar";

interface LayoutProps {
  showSidebar?: boolean;
}

export function Layout({ showSidebar = true }: LayoutProps) {
  const { sidebarWidth, sidebarCollapsed } = useUIStore();
  const effectiveWidth = sidebarCollapsed ? MIN_SIDEBAR_WIDTH : sidebarWidth;

  return (
    <Box minH="100vh">
      <Header />
      <Flex>
        {showSidebar && <Sidebar />}
        <Box
          as="main"
          flex={1}
          p={6}
          maxW={showSidebar ? `calc(100% - ${effectiveWidth}px)` : "100%"}
          ml={{ base: 0, lg: 0 }}
          transition="max-width 0.2s"
        >
          <Outlet />
        </Box>
      </Flex>
    </Box>
  );
}
