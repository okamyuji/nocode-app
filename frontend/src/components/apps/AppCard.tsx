import { useAuth, useDeleteApp } from "@/hooks";
import { App } from "@/types";
import {
  AlertDialog,
  AlertDialogBody,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogOverlay,
  Button,
  Card,
  CardBody,
  Heading,
  HStack,
  Icon,
  IconButton,
  Menu,
  MenuButton,
  MenuItem,
  MenuList,
  Portal,
  Text,
  useDisclosure,
  useToast,
  VStack,
} from "@chakra-ui/react";
import { useRef } from "react";
import {
  FiEdit2,
  FiGrid,
  FiMoreVertical,
  FiSettings,
  FiTrash2,
} from "react-icons/fi";
import { useNavigate } from "react-router-dom";

interface AppCardProps {
  app: App;
}

export function AppCard({ app }: AppCardProps) {
  const navigate = useNavigate();
  const { isAdmin } = useAuth();
  const { isOpen, onOpen, onClose } = useDisclosure();
  const cancelRef = useRef<HTMLButtonElement>(null);
  const deleteApp = useDeleteApp();
  const toast = useToast();

  const handleDelete = async () => {
    try {
      await deleteApp.mutateAsync(app.id);
      toast({
        title: "アプリを削除しました",
        status: "success",
        duration: 3000,
        isClosable: true,
      });
      onClose();
    } catch {
      toast({
        title: "削除に失敗しました",
        status: "error",
        duration: 5000,
        isClosable: true,
      });
    }
  };

  return (
    <>
      <Card
        cursor="pointer"
        _hover={{ shadow: "md", transform: "translateY(-2px)" }}
        transition="all 0.2s"
        onClick={() => navigate(`/apps/${app.id}/records`)}
      >
        <CardBody>
          <HStack justify="space-between" align="start">
            <HStack spacing={4}>
              <Icon as={FiGrid} boxSize={10} color="brand.500" />
              <VStack align="start" spacing={1}>
                <Heading size="md">{app.name}</Heading>
                <Text fontSize="sm" color="gray.500" noOfLines={2}>
                  {app.description || "説明なし"}
                </Text>
              </VStack>
            </HStack>

            <Menu>
              <MenuButton
                as={IconButton}
                icon={<FiMoreVertical />}
                variant="ghost"
                size="sm"
                onClick={(e) => e.stopPropagation()}
              />
              <Portal>
                <MenuList
                  onClick={(e) => e.stopPropagation()}
                  zIndex="dropdown"
                >
                  <MenuItem
                    icon={<FiEdit2 />}
                    onClick={() => navigate(`/apps/${app.id}/records`)}
                  >
                    レコード管理
                  </MenuItem>
                  {isAdmin && (
                    <>
                      <MenuItem
                        icon={<FiSettings />}
                        onClick={() =>
                          navigate(`/settings?tab=apps&appId=${app.id}`)
                        }
                      >
                        アプリ設定
                      </MenuItem>
                      <MenuItem
                        icon={<FiTrash2 />}
                        color="red.500"
                        onClick={onOpen}
                      >
                        削除
                      </MenuItem>
                    </>
                  )}
                </MenuList>
              </Portal>
            </Menu>
          </HStack>
        </CardBody>
      </Card>

      <AlertDialog
        isOpen={isOpen}
        leastDestructiveRef={cancelRef}
        onClose={onClose}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader fontSize="lg" fontWeight="bold">
              アプリの削除
            </AlertDialogHeader>

            <AlertDialogBody>
              「{app.name}」を削除しますか？
              この操作は取り消せません。アプリ内のすべてのデータも削除されます。
            </AlertDialogBody>

            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={onClose}>
                キャンセル
              </Button>
              <Button
                colorScheme="red"
                onClick={handleDelete}
                ml={3}
                isLoading={deleteApp.isPending}
              >
                削除
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
    </>
  );
}
