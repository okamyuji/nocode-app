import { useAuth, useDeleteApp } from "@/hooks";
import { App } from "@/types";
import { getAppIcon } from "@/utils";
import {
  Badge,
  Button,
  Card,
  Dialog,
  Heading,
  HStack,
  Icon,
  IconButton,
  Menu,
  Portal,
  Text,
  useDisclosure,
  VStack,
} from "@chakra-ui/react";
import { toaster } from "@/components/ui/toaster";
import { useRef } from "react";
import { FiEdit2, FiMoreVertical, FiSettings, FiTrash2 } from "react-icons/fi";
import { useNavigate } from "react-router-dom";

interface AppCardProps {
  app: App;
}

export function AppCard({ app }: AppCardProps) {
  const navigate = useNavigate();
  const { isAdmin } = useAuth();
  const { open: isOpen, onOpen, onClose } = useDisclosure();
  const cancelRef = useRef<HTMLButtonElement>(null);
  const deleteApp = useDeleteApp();

  const handleDelete = async () => {
    try {
      await deleteApp.mutateAsync(app.id);
      toaster.create({
        title: "アプリを削除しました",
        type: "success",
        duration: 3000,
        closable: true,
      });
      onClose();
    } catch {
      toaster.create({
        title: "削除に失敗しました",
        type: "error",
        duration: 5000,
        closable: true,
      });
    }
  };

  return (
    <>
      <Card.Root
        cursor="pointer"
        _hover={{ shadow: "md", transform: "translateY(-2px)" }}
        transition="all 0.2s"
        onClick={() => navigate(`/apps/${app.id}/records`)}
      >
        <Card.Body>
          <HStack justify="space-between" align="start">
            <HStack gap={4}>
              <Icon as={getAppIcon(app.icon)} boxSize={10} color="brand.500" />
              <VStack align="start" gap={1}>
                <HStack>
                  <Heading size="md">{app.name}</Heading>
                  {app.is_external && (
                    <Badge colorPalette="purple" fontSize="xs">
                      外部
                    </Badge>
                  )}
                </HStack>
                <Text fontSize="sm" color="gray.500" lineClamp={2}>
                  {app.description || "説明なし"}
                </Text>
              </VStack>
            </HStack>

            <Menu.Root lazyMount>
              <Menu.Trigger asChild>
                <IconButton
                  aria-label="メニュー"
                  variant="ghost"
                  size="sm"
                  onClick={(e) => e.stopPropagation()}
                >
                  <FiMoreVertical />
                </IconButton>
              </Menu.Trigger>
              <Portal>
                <Menu.Positioner>
                  <Menu.Content
                    onClick={(e) => e.stopPropagation()}
                    zIndex="dropdown"
                  >
                    <Menu.Item
                      value="records"
                      onClick={() => navigate(`/apps/${app.id}/records`)}
                    >
                      <FiEdit2 />
                      レコード管理
                    </Menu.Item>
                    {isAdmin && (
                      <>
                        <Menu.Item
                          value="settings"
                          onClick={() =>
                            navigate(`/settings?tab=apps&appId=${app.id}`)
                          }
                        >
                          <FiSettings />
                          アプリ設定
                        </Menu.Item>
                        <Menu.Item
                          value="delete"
                          color="red.500"
                          onClick={onOpen}
                        >
                          <FiTrash2 />
                          削除
                        </Menu.Item>
                      </>
                    )}
                  </Menu.Content>
                </Menu.Positioner>
              </Portal>
            </Menu.Root>
          </HStack>
        </Card.Body>
      </Card.Root>

      <Dialog.Root
        role="alertdialog"
        open={isOpen}
        onOpenChange={(e) => !e.open && onClose()}
        initialFocusEl={() => cancelRef.current}
      >
        <Portal>
          <Dialog.Backdrop />
          <Dialog.Positioner>
            <Dialog.Content>
              <Dialog.Header fontSize="lg" fontWeight="bold">
                <Dialog.Title>アプリの削除</Dialog.Title>
              </Dialog.Header>

              <Dialog.Body>
                「{app.name}」を削除しますか？
                この操作は取り消せません。アプリ内のすべてのデータも削除されます。
              </Dialog.Body>

              <Dialog.Footer>
                <Button ref={cancelRef} onClick={onClose}>
                  キャンセル
                </Button>
                <Button
                  colorPalette="red"
                  onClick={handleDelete}
                  ml={3}
                  loading={deleteApp.isPending}
                >
                  削除
                </Button>
              </Dialog.Footer>
            </Dialog.Content>
          </Dialog.Positioner>
        </Portal>
      </Dialog.Root>
    </>
  );
}
