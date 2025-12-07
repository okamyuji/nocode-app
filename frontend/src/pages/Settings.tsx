import { useApiClient } from "@/api";
import { DataSourceList } from "@/components/datasources";
import { useAuthStore } from "@/stores";
import type {
  App,
  ChangePasswordRequest,
  CreateFieldRequest,
  CreateUserRequest,
  Field,
  FieldOptions,
  FieldType,
  UpdateAppRequest,
  UpdateFieldRequest,
  UpdateProfileRequest,
  UpdateUserRequest,
  User,
} from "@/types";
import { FIELD_TYPE_LABELS } from "@/types";
import {
  AlertDialog,
  AlertDialogBody,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogOverlay,
  Badge,
  Box,
  Button,
  Card,
  CardBody,
  CardHeader,
  Flex,
  FormControl,
  FormErrorMessage,
  FormLabel,
  Grid,
  Heading,
  Icon,
  IconButton,
  Input,
  InputGroup,
  InputLeftElement,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Select,
  Skeleton,
  SkeletonText,
  Switch,
  Tab,
  Table,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Tag,
  Tbody,
  Td,
  Text,
  Th,
  Thead,
  Tr,
  useDisclosure,
  useToast,
  VStack,
} from "@chakra-ui/react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  FiArrowLeft,
  FiDatabase,
  FiEdit2,
  FiPlus,
  FiSearch,
  FiSettings,
  FiTrash2,
} from "react-icons/fi";
import { useSearchParams } from "react-router-dom";

export function SettingsPage() {
  const { user } = useAuthStore();
  const isAdmin = user?.role === "admin";
  const [searchParams, setSearchParams] = useSearchParams();

  // Get initial tab from URL query param
  const tabParam = searchParams.get("tab");
  const getInitialTabIndex = () => {
    if (isAdmin) {
      // admin: profile(0), password(1), users(2), apps(3), datasources(4)
      if (tabParam === "datasources") return 4;
      if (tabParam === "apps") return 3;
      if (tabParam === "users") return 2;
      if (tabParam === "password") return 1;
      return 0; // profile
    } else {
      // user: profile(0), password(1)
      if (tabParam === "password") return 1;
      return 0; // profile
    }
  };
  const initialTabIndex = getInitialTabIndex();

  const handleTabChange = (index: number) => {
    const tabNames = isAdmin
      ? ["profile", "password", "users", "apps", "datasources"]
      : ["profile", "password"];
    setSearchParams({ tab: tabNames[index] });
  };

  return (
    <Box p={6}>
      <Heading size="lg" mb={6}>
        設定
      </Heading>

      <Tabs
        colorScheme="brand"
        defaultIndex={initialTabIndex}
        onChange={handleTabChange}
      >
        <TabList>
          <Tab>プロフィール</Tab>
          <Tab>パスワード変更</Tab>
          {isAdmin && <Tab>ユーザー管理</Tab>}
          {isAdmin && <Tab>アプリ設定</Tab>}
          {isAdmin && <Tab>データソース</Tab>}
        </TabList>

        <TabPanels>
          <TabPanel>
            <ProfileSettings />
          </TabPanel>
          <TabPanel>
            <PasswordSettings />
          </TabPanel>
          {isAdmin && (
            <TabPanel>
              <UserManagement />
            </TabPanel>
          )}
          {isAdmin && (
            <TabPanel>
              <AppSettings />
            </TabPanel>
          )}
          {isAdmin && (
            <TabPanel>
              <DataSourceList />
            </TabPanel>
          )}
        </TabPanels>
      </Tabs>
    </Box>
  );
}

// Profile Settings Component
function ProfileSettings() {
  const { user, setUser } = useAuthStore();
  const { profile } = useApiClient();
  const toast = useToast();
  const [name, setName] = useState(user?.name || "");
  const [error, setError] = useState("");

  const updateProfileMutation = useMutation({
    mutationFn: (data: UpdateProfileRequest) => profile.updateProfile(data),
    onSuccess: (updatedUser) => {
      setUser(updatedUser);
      toast({
        title: "プロフィールを更新しました",
        status: "success",
        duration: 3000,
      });
    },
    onError: () => {
      toast({
        title: "プロフィールの更新に失敗しました",
        status: "error",
        duration: 3000,
      });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!name.trim()) {
      setError("名前を入力してください");
      return;
    }

    updateProfileMutation.mutate({ name: name.trim() });
  };

  return (
    <Card maxW="md">
      <CardHeader>
        <Heading size="md">プロフィール設定</Heading>
      </CardHeader>
      <CardBody>
        <form onSubmit={handleSubmit}>
          <VStack spacing={4}>
            <FormControl>
              <FormLabel>メールアドレス</FormLabel>
              <Input value={user?.email || ""} isReadOnly bg="gray.50" />
            </FormControl>

            <FormControl>
              <FormLabel>ロール</FormLabel>
              <Input
                value={user?.role === "admin" ? "管理者" : "一般ユーザー"}
                isReadOnly
                bg="gray.50"
              />
            </FormControl>

            <FormControl isInvalid={!!error}>
              <FormLabel>名前</FormLabel>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="名前を入力"
              />
              <FormErrorMessage>{error}</FormErrorMessage>
            </FormControl>

            <Button
              type="submit"
              colorScheme="brand"
              w="full"
              isLoading={updateProfileMutation.isPending}
            >
              保存
            </Button>
          </VStack>
        </form>
      </CardBody>
    </Card>
  );
}

// Password Settings Component
function PasswordSettings() {
  const { profile } = useApiClient();
  const toast = useToast();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  const changePasswordMutation = useMutation({
    mutationFn: (data: ChangePasswordRequest) => profile.changePassword(data),
    onSuccess: () => {
      toast({
        title: "パスワードを変更しました",
        status: "success",
        duration: 3000,
      });
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    },
    onError: () => {
      toast({
        title: "パスワードの変更に失敗しました",
        description: "現在のパスワードが正しくない可能性があります",
        status: "error",
        duration: 3000,
      });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const newErrors: Record<string, string> = {};

    if (!currentPassword) {
      newErrors.currentPassword = "現在のパスワードを入力してください";
    }
    if (!newPassword) {
      newErrors.newPassword = "新しいパスワードを入力してください";
    } else if (newPassword.length < 6) {
      newErrors.newPassword = "パスワードは6文字以上である必要があります";
    }
    if (newPassword !== confirmPassword) {
      newErrors.confirmPassword = "パスワードが一致しません";
    }

    setErrors(newErrors);

    if (Object.keys(newErrors).length === 0) {
      changePasswordMutation.mutate({
        current_password: currentPassword,
        new_password: newPassword,
      });
    }
  };

  return (
    <Card maxW="md">
      <CardHeader>
        <Heading size="md">パスワード変更</Heading>
      </CardHeader>
      <CardBody>
        <form onSubmit={handleSubmit}>
          <VStack spacing={4}>
            <FormControl isInvalid={!!errors.currentPassword}>
              <FormLabel>現在のパスワード</FormLabel>
              <Input
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
              />
              <FormErrorMessage>{errors.currentPassword}</FormErrorMessage>
            </FormControl>

            <FormControl isInvalid={!!errors.newPassword}>
              <FormLabel>新しいパスワード</FormLabel>
              <Input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
              />
              <FormErrorMessage>{errors.newPassword}</FormErrorMessage>
            </FormControl>

            <FormControl isInvalid={!!errors.confirmPassword}>
              <FormLabel>新しいパスワード（確認）</FormLabel>
              <Input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
              />
              <FormErrorMessage>{errors.confirmPassword}</FormErrorMessage>
            </FormControl>

            <Button
              type="submit"
              colorScheme="brand"
              w="full"
              isLoading={changePasswordMutation.isPending}
            >
              パスワードを変更
            </Button>
          </VStack>
        </form>
      </CardBody>
    </Card>
  );
}

// User Management Component (Admin Only)
function UserManagement() {
  const { users } = useApiClient();
  const { user: currentUser } = useAuthStore();
  const queryClient = useQueryClient();
  const toast = useToast();
  const [page, setPage] = useState(1);
  const limit = 20;

  const {
    isOpen: isCreateOpen,
    onOpen: onCreateOpen,
    onClose: onCreateClose,
  } = useDisclosure();
  const {
    isOpen: isEditOpen,
    onOpen: onEditOpen,
    onClose: onEditClose,
  } = useDisclosure();
  const {
    isOpen: isDeleteOpen,
    onOpen: onDeleteOpen,
    onClose: onDeleteClose,
  } = useDisclosure();

  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const cancelRef = useRef<HTMLButtonElement>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["users", page, limit],
    queryFn: () => users.getAll(page, limit),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => users.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toast({
        title: "ユーザーを削除しました",
        status: "success",
        duration: 3000,
      });
      onDeleteClose();
    },
    onError: () => {
      toast({
        title: "ユーザーの削除に失敗しました",
        status: "error",
        duration: 3000,
      });
    },
  });

  const handleEdit = (user: User) => {
    setSelectedUser(user);
    onEditOpen();
  };

  const handleDelete = (user: User) => {
    setSelectedUser(user);
    onDeleteOpen();
  };

  const confirmDelete = () => {
    if (selectedUser) {
      deleteMutation.mutate(selectedUser.id);
    }
  };

  return (
    <Box>
      <Flex justify="space-between" align="center" mb={4}>
        <Heading size="md">ユーザー管理</Heading>
        <Button
          leftIcon={<FiPlus />}
          colorScheme="brand"
          onClick={onCreateOpen}
        >
          新規ユーザー
        </Button>
      </Flex>

      <Card>
        <CardBody p={0}>
          <Table variant="simple">
            <Thead>
              <Tr>
                <Th>名前</Th>
                <Th>メールアドレス</Th>
                <Th>ロール</Th>
                <Th>作成日</Th>
                <Th w="100px">操作</Th>
              </Tr>
            </Thead>
            <Tbody>
              {isLoading ? (
                <Tr>
                  <Td colSpan={5} textAlign="center">
                    読み込み中...
                  </Td>
                </Tr>
              ) : data?.users?.length === 0 ? (
                <Tr>
                  <Td colSpan={5} textAlign="center">
                    ユーザーがいません
                  </Td>
                </Tr>
              ) : (
                data?.users?.map((user) => (
                  <Tr key={user.id}>
                    <Td>{user.name}</Td>
                    <Td>{user.email}</Td>
                    <Td>
                      <Badge
                        colorScheme={user.role === "admin" ? "purple" : "gray"}
                      >
                        {user.role === "admin" ? "管理者" : "一般"}
                      </Badge>
                    </Td>
                    <Td>
                      {new Date(user.created_at).toLocaleDateString("ja-JP")}
                    </Td>
                    <Td>
                      <Flex gap={2}>
                        <IconButton
                          aria-label="編集"
                          icon={<FiEdit2 />}
                          size="sm"
                          variant="ghost"
                          onClick={() => handleEdit(user)}
                        />
                        <IconButton
                          aria-label="削除"
                          icon={<FiTrash2 />}
                          size="sm"
                          variant="ghost"
                          colorScheme="red"
                          isDisabled={user.id === currentUser?.id}
                          onClick={() => handleDelete(user)}
                        />
                      </Flex>
                    </Td>
                  </Tr>
                ))
              )}
            </Tbody>
          </Table>
        </CardBody>
      </Card>

      {data?.pagination && data.pagination.total_pages > 1 && (
        <Flex justify="center" mt={4} gap={2}>
          <Button
            size="sm"
            isDisabled={page === 1}
            onClick={() => setPage((p) => p - 1)}
          >
            前へ
          </Button>
          <Text alignSelf="center">
            {page} / {data.pagination.total_pages}
          </Text>
          <Button
            size="sm"
            isDisabled={page === data.pagination.total_pages}
            onClick={() => setPage((p) => p + 1)}
          >
            次へ
          </Button>
        </Flex>
      )}

      {/* Create User Modal */}
      <CreateUserModal isOpen={isCreateOpen} onClose={onCreateClose} />

      {/* Edit User Modal */}
      {selectedUser && (
        <EditUserModal
          isOpen={isEditOpen}
          onClose={onEditClose}
          user={selectedUser}
        />
      )}

      {/* Delete Confirmation */}
      <AlertDialog
        isOpen={isDeleteOpen}
        leastDestructiveRef={cancelRef}
        onClose={onDeleteClose}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader>ユーザーの削除</AlertDialogHeader>
            <AlertDialogBody>
              「{selectedUser?.name}」を削除してもよろしいですか？
              この操作は取り消せません。
            </AlertDialogBody>
            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={onDeleteClose}>
                キャンセル
              </Button>
              <Button
                colorScheme="red"
                onClick={confirmDelete}
                ml={3}
                isLoading={deleteMutation.isPending}
              >
                削除
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
    </Box>
  );
}

// Create User Modal
function CreateUserModal({
  isOpen,
  onClose,
}: {
  isOpen: boolean;
  onClose: () => void;
}) {
  const { users } = useApiClient();
  const queryClient = useQueryClient();
  const toast = useToast();
  const [formData, setFormData] = useState({
    email: "",
    password: "",
    name: "",
    role: "user" as "admin" | "user",
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const createMutation = useMutation({
    mutationFn: (data: CreateUserRequest) => users.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toast({
        title: "ユーザーを作成しました",
        status: "success",
        duration: 3000,
      });
      onClose();
      setFormData({ email: "", password: "", name: "", role: "user" });
    },
    onError: () => {
      toast({
        title: "ユーザーの作成に失敗しました",
        description: "メールアドレスが既に使用されている可能性があります",
        status: "error",
        duration: 3000,
      });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const newErrors: Record<string, string> = {};

    if (!formData.email) newErrors.email = "メールアドレスを入力してください";
    if (!formData.password) newErrors.password = "パスワードを入力してください";
    else if (formData.password.length < 6)
      newErrors.password = "パスワードは6文字以上である必要があります";
    if (!formData.name) newErrors.name = "名前を入力してください";

    setErrors(newErrors);

    if (Object.keys(newErrors).length === 0) {
      createMutation.mutate(formData);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <ModalOverlay />
      <ModalContent>
        <form onSubmit={handleSubmit}>
          <ModalHeader>新規ユーザー作成</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <FormControl isInvalid={!!errors.email}>
                <FormLabel>メールアドレス</FormLabel>
                <Input
                  type="email"
                  value={formData.email}
                  onChange={(e) =>
                    setFormData({ ...formData, email: e.target.value })
                  }
                />
                <FormErrorMessage>{errors.email}</FormErrorMessage>
              </FormControl>

              <FormControl isInvalid={!!errors.password}>
                <FormLabel>パスワード</FormLabel>
                <Input
                  type="password"
                  value={formData.password}
                  onChange={(e) =>
                    setFormData({ ...formData, password: e.target.value })
                  }
                />
                <FormErrorMessage>{errors.password}</FormErrorMessage>
              </FormControl>

              <FormControl isInvalid={!!errors.name}>
                <FormLabel>名前</FormLabel>
                <Input
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                />
                <FormErrorMessage>{errors.name}</FormErrorMessage>
              </FormControl>

              <FormControl>
                <FormLabel>ロール</FormLabel>
                <Select
                  value={formData.role}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      role: e.target.value as "admin" | "user",
                    })
                  }
                >
                  <option value="user">一般ユーザー</option>
                  <option value="admin">管理者</option>
                </Select>
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onClose}>
              キャンセル
            </Button>
            <Button
              type="submit"
              colorScheme="brand"
              isLoading={createMutation.isPending}
            >
              作成
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
}

// Edit User Modal
function EditUserModal({
  isOpen,
  onClose,
  user,
}: {
  isOpen: boolean;
  onClose: () => void;
  user: User;
}) {
  const { users } = useApiClient();
  const { user: currentUser } = useAuthStore();
  const queryClient = useQueryClient();
  const toast = useToast();
  const [formData, setFormData] = useState({
    name: user.name,
    role: user.role as "admin" | "user",
  });
  const [error, setError] = useState("");

  const updateMutation = useMutation({
    mutationFn: (data: UpdateUserRequest) => users.update(user.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toast({
        title: "ユーザーを更新しました",
        status: "success",
        duration: 3000,
      });
      onClose();
    },
    onError: () => {
      toast({
        title: "ユーザーの更新に失敗しました",
        status: "error",
        duration: 3000,
      });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!formData.name.trim()) {
      setError("名前を入力してください");
      return;
    }

    updateMutation.mutate(formData);
  };

  const isSelf = user.id === currentUser?.id;

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <ModalOverlay />
      <ModalContent>
        <form onSubmit={handleSubmit}>
          <ModalHeader>ユーザー編集</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <FormControl>
                <FormLabel>メールアドレス</FormLabel>
                <Input value={user.email} isReadOnly bg="gray.50" />
              </FormControl>

              <FormControl isInvalid={!!error}>
                <FormLabel>名前</FormLabel>
                <Input
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                />
                <FormErrorMessage>{error}</FormErrorMessage>
              </FormControl>

              <FormControl>
                <FormLabel>ロール</FormLabel>
                <Select
                  value={formData.role}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      role: e.target.value as "admin" | "user",
                    })
                  }
                  isDisabled={isSelf}
                >
                  <option value="user">一般ユーザー</option>
                  <option value="admin">管理者</option>
                </Select>
                {isSelf && (
                  <Text fontSize="sm" color="gray.500" mt={1}>
                    自分のロールは変更できません
                  </Text>
                )}
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onClose}>
              キャンセル
            </Button>
            <Button
              type="submit"
              colorScheme="brand"
              isLoading={updateMutation.isPending}
            >
              保存
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
}

// ==================== App Settings Component ====================

function AppSettings() {
  const [selectedApp, setSelectedApp] = useState<App | null>(null);
  const [searchParams] = useSearchParams();
  const { apps } = useApiClient();

  // Get app ID from URL if provided
  const appIdFromUrl = searchParams.get("appId");

  const { data, isLoading } = useQuery({
    queryKey: ["apps", 1, 100],
    queryFn: () => apps.getAll(1, 100),
  });

  // Auto-select app if appId is in URL
  useEffect(() => {
    if (appIdFromUrl && data?.apps) {
      const app = data.apps.find((a) => a.id === Number(appIdFromUrl));
      if (app) {
        setSelectedApp(app);
      }
    }
  }, [appIdFromUrl, data?.apps]);

  if (selectedApp) {
    return (
      <AppSettingsDetail
        app={selectedApp}
        onBack={() => setSelectedApp(null)}
      />
    );
  }

  return (
    <AppSettingsList
      apps={data?.apps || []}
      isLoading={isLoading}
      onSelectApp={setSelectedApp}
    />
  );
}

// App Settings List Component
interface AppSettingsListProps {
  apps: App[];
  isLoading: boolean;
  onSelectApp: (app: App) => void;
}

function AppSettingsList({
  apps,
  isLoading,
  onSelectApp,
}: AppSettingsListProps) {
  const [searchQuery, setSearchQuery] = useState("");

  const filteredApps = useMemo(() => {
    if (!searchQuery.trim()) return apps;
    const query = searchQuery.toLowerCase();
    return apps.filter(
      (app) =>
        app.name.toLowerCase().includes(query) ||
        app.description?.toLowerCase().includes(query)
    );
  }, [apps, searchQuery]);

  return (
    <Box>
      <Flex justify="space-between" align="center" mb={4}>
        <Box>
          <Heading size="md" mb={1}>
            アプリ設定
          </Heading>
          <Text color="gray.600" fontSize="sm">
            アプリを選択して設定を変更します
          </Text>
        </Box>
      </Flex>

      <InputGroup maxW="400px" mb={4}>
        <InputLeftElement>
          <Icon as={FiSearch} color="gray.400" />
        </InputLeftElement>
        <Input
          placeholder="アプリを検索..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
        />
      </InputGroup>

      {isLoading ? (
        <Grid templateColumns="repeat(auto-fill, minmax(280px, 1fr))" gap={4}>
          {[1, 2, 3, 4].map((i) => (
            <Card key={i}>
              <CardBody>
                <Skeleton height="20px" width="60%" mb={2} />
                <SkeletonText noOfLines={2} />
              </CardBody>
            </Card>
          ))}
        </Grid>
      ) : filteredApps.length === 0 ? (
        <Card>
          <CardBody textAlign="center" py={10}>
            <Icon as={FiDatabase} boxSize={12} color="gray.300" mb={4} />
            <Text fontSize="lg" color="gray.500">
              {searchQuery
                ? "検索条件に一致するアプリがありません"
                : "アプリがありません"}
            </Text>
          </CardBody>
        </Card>
      ) : (
        <Grid templateColumns="repeat(auto-fill, minmax(280px, 1fr))" gap={4}>
          {filteredApps.map((app) => (
            <Card
              key={app.id}
              cursor="pointer"
              _hover={{
                transform: "translateY(-2px)",
                shadow: "md",
                borderColor: "brand.300",
              }}
              transition="all 0.2s"
              onClick={() => onSelectApp(app)}
              borderWidth="1px"
              borderColor="gray.200"
            >
              <CardBody>
                <Flex align="start" gap={3}>
                  <Flex
                    w={10}
                    h={10}
                    bg="brand.50"
                    borderRadius="lg"
                    align="center"
                    justify="center"
                    flexShrink={0}
                  >
                    <Icon as={FiSettings} color="brand.500" boxSize={5} />
                  </Flex>
                  <Box flex={1} minW={0}>
                    <Flex align="center" gap={2} mb={1}>
                      <Heading size="sm" noOfLines={1}>
                        {app.name}
                      </Heading>
                      <Badge colorScheme="brand" fontSize="xs">
                        {app.field_count} フィールド
                      </Badge>
                    </Flex>
                    <Text fontSize="sm" color="gray.600" noOfLines={2}>
                      {app.description || "No description"}
                    </Text>
                  </Box>
                </Flex>
              </CardBody>
            </Card>
          ))}
        </Grid>
      )}
    </Box>
  );
}

// App Settings Detail Component
interface AppSettingsDetailProps {
  app: App;
  onBack: () => void;
}

function AppSettingsDetail({ app, onBack }: AppSettingsDetailProps) {
  const { apps, fields } = useApiClient();
  const queryClient = useQueryClient();
  const toast = useToast();
  const cancelRef = useRef<HTMLButtonElement>(null);

  // Form state for basic info
  const [name, setName] = useState(app.name);
  const [description, setDescription] = useState(app.description || "");
  const [icon, setIcon] = useState(app.icon || "default");

  // propsのappが変更された場合にローカルstateを同期
  useEffect(() => {
    setName(app.name);
    setDescription(app.description || "");
    setIcon(app.icon || "default");
  }, [app]);

  // Modal states
  const {
    isOpen: isDeleteAppOpen,
    onOpen: onDeleteAppOpen,
    onClose: onDeleteAppClose,
  } = useDisclosure();
  const {
    isOpen: isAddFieldOpen,
    onOpen: onAddFieldOpen,
    onClose: onAddFieldClose,
  } = useDisclosure();
  const {
    isOpen: isEditFieldOpen,
    onOpen: onEditFieldOpen,
    onClose: onEditFieldClose,
  } = useDisclosure();
  const {
    isOpen: isDeleteFieldOpen,
    onOpen: onDeleteFieldOpen,
    onClose: onDeleteFieldClose,
  } = useDisclosure();

  const [selectedField, setSelectedField] = useState<Field | null>(null);

  // Fetch fields
  const { data: fieldsData, isLoading: fieldsLoading } = useQuery({
    queryKey: ["fields", app.id],
    queryFn: () => fields.getByAppId(app.id),
  });

  // Update app mutation
  const updateAppMutation = useMutation({
    mutationFn: (data: UpdateAppRequest) => apps.update(app.id, data),
    onSuccess: (updatedApp) => {
      // ローカルstateを更新してUIに即時反映
      setName(updatedApp.name);
      setDescription(updatedApp.description || "");
      setIcon(updatedApp.icon || "default");

      // キャッシュを直接更新して他の画面でも即時反映
      queryClient.setQueriesData<{ apps: App[] }>(
        { queryKey: ["apps"] },
        (oldData) => {
          if (!oldData) return oldData;
          return {
            ...oldData,
            apps: oldData.apps.map((a) =>
              a.id === updatedApp.id ? updatedApp : a
            ),
          };
        }
      );

      // 個別アプリのキャッシュも更新
      queryClient.setQueryData(["app", app.id], updatedApp);

      toast({
        title: "アプリを更新しました",
        status: "success",
        duration: 3000,
      });
    },
    onError: () => {
      toast({
        title: "アプリの更新に失敗しました",
        status: "error",
        duration: 3000,
      });
    },
  });

  // Delete app mutation
  const deleteAppMutation = useMutation({
    mutationFn: () => apps.delete(app.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["apps"] });
      toast({
        title: "アプリを削除しました",
        status: "success",
        duration: 3000,
      });
      onBack();
    },
    onError: () => {
      toast({
        title: "アプリの削除に失敗しました",
        status: "error",
        duration: 3000,
      });
    },
  });

  // Delete field mutation
  const deleteFieldMutation = useMutation({
    mutationFn: (fieldId: number) => fields.delete(app.id, fieldId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["fields", app.id] });
      toast({
        title: "フィールドを削除しました",
        status: "success",
        duration: 3000,
      });
      onDeleteFieldClose();
    },
    onError: () => {
      toast({
        title: "フィールドの削除に失敗しました",
        status: "error",
        duration: 3000,
      });
    },
  });

  const handleSaveBasicInfo = () => {
    if (!name.trim()) {
      toast({
        title: "アプリ名を入力してください",
        status: "error",
        duration: 3000,
      });
      return;
    }
    updateAppMutation.mutate({ name: name.trim(), description, icon });
  };

  const handleEditField = (field: Field) => {
    setSelectedField(field);
    onEditFieldOpen();
  };

  const handleDeleteField = (field: Field) => {
    setSelectedField(field);
    onDeleteFieldOpen();
  };

  const confirmDeleteField = () => {
    if (selectedField) {
      deleteFieldMutation.mutate(selectedField.id);
    }
  };

  const confirmDeleteApp = () => {
    deleteAppMutation.mutate();
  };

  return (
    <Box>
      <Button
        leftIcon={<FiArrowLeft />}
        variant="ghost"
        mb={4}
        onClick={onBack}
      >
        アプリ一覧に戻る
      </Button>

      <Flex justify="space-between" align="center" mb={6}>
        <Heading size="md">{app.name} の設定</Heading>
        <Button
          colorScheme="red"
          variant="outline"
          leftIcon={<FiTrash2 />}
          onClick={onDeleteAppOpen}
        >
          アプリを削除
        </Button>
      </Flex>

      <Grid templateColumns={{ base: "1fr", lg: "1fr 1fr" }} gap={6}>
        {/* Basic Info Card */}
        <Card>
          <CardHeader>
            <Heading size="sm">基本情報</Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={4}>
              <FormControl>
                <FormLabel>アプリ名</FormLabel>
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="アプリ名を入力"
                />
              </FormControl>

              <FormControl>
                <FormLabel>説明</FormLabel>
                <Input
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="アプリの説明を入力"
                />
              </FormControl>

              <FormControl>
                <FormLabel>アイコン</FormLabel>
                <Select value={icon} onChange={(e) => setIcon(e.target.value)}>
                  <option value="default">デフォルト</option>
                  <option value="grid">グリッド</option>
                  <option value="list">リスト</option>
                  <option value="calendar">カレンダー</option>
                  <option value="database">データベース</option>
                </Select>
              </FormControl>

              <Button
                colorScheme="brand"
                w="full"
                onClick={handleSaveBasicInfo}
                isLoading={updateAppMutation.isPending}
              >
                保存
              </Button>
            </VStack>
          </CardBody>
        </Card>

        {/* Fields Card */}
        <Card>
          <CardHeader>
            <Flex justify="space-between" align="center">
              <Heading size="sm">フィールド管理</Heading>
              <Button
                size="sm"
                leftIcon={<FiPlus />}
                colorScheme="brand"
                onClick={onAddFieldOpen}
              >
                フィールド追加
              </Button>
            </Flex>
          </CardHeader>
          <CardBody p={0}>
            {fieldsLoading ? (
              <Box p={4}>
                <SkeletonText noOfLines={5} />
              </Box>
            ) : !fieldsData?.fields?.length ? (
              <Box p={4} textAlign="center">
                <Text color="gray.500">フィールドがありません</Text>
              </Box>
            ) : (
              <Table variant="simple" size="sm">
                <Thead>
                  <Tr>
                    <Th>名前</Th>
                    <Th>タイプ</Th>
                    <Th>必須</Th>
                    <Th w="80px">操作</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {fieldsData.fields.map((field) => (
                    <Tr key={field.id}>
                      <Td>
                        <VStack align="start" spacing={0}>
                          <Text fontWeight="medium">{field.field_name}</Text>
                          <Text fontSize="xs" color="gray.500">
                            {field.field_code}
                          </Text>
                        </VStack>
                      </Td>
                      <Td>
                        <Tag size="sm" colorScheme="blue">
                          {FIELD_TYPE_LABELS[field.field_type] ||
                            field.field_type}
                        </Tag>
                      </Td>
                      <Td>
                        {field.required ? (
                          <Badge colorScheme="red">必須</Badge>
                        ) : (
                          <Text color="gray.400">-</Text>
                        )}
                      </Td>
                      <Td>
                        <Flex gap={1}>
                          <IconButton
                            aria-label="編集"
                            icon={<FiEdit2 />}
                            size="xs"
                            variant="ghost"
                            onClick={() => handleEditField(field)}
                          />
                          <IconButton
                            aria-label="削除"
                            icon={<FiTrash2 />}
                            size="xs"
                            variant="ghost"
                            colorScheme="red"
                            onClick={() => handleDeleteField(field)}
                          />
                        </Flex>
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            )}
          </CardBody>
        </Card>
      </Grid>

      {/* Add Field Modal */}
      <AddFieldModal
        isOpen={isAddFieldOpen}
        onClose={onAddFieldClose}
        appId={app.id}
        existingFieldCodes={fieldsData?.fields?.map((f) => f.field_code) || []}
      />

      {/* Edit Field Modal */}
      {selectedField && (
        <EditFieldModal
          isOpen={isEditFieldOpen}
          onClose={onEditFieldClose}
          appId={app.id}
          field={selectedField}
        />
      )}

      {/* Delete Field Confirmation */}
      <AlertDialog
        isOpen={isDeleteFieldOpen}
        leastDestructiveRef={cancelRef}
        onClose={onDeleteFieldClose}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader>フィールドの削除</AlertDialogHeader>
            <AlertDialogBody>
              「{selectedField?.field_name}」を削除してもよろしいですか？
              このフィールドに保存されたデータも削除されます。
            </AlertDialogBody>
            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={onDeleteFieldClose}>
                キャンセル
              </Button>
              <Button
                colorScheme="red"
                onClick={confirmDeleteField}
                ml={3}
                isLoading={deleteFieldMutation.isPending}
              >
                削除
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>

      {/* Delete App Confirmation */}
      <AlertDialog
        isOpen={isDeleteAppOpen}
        leastDestructiveRef={cancelRef}
        onClose={onDeleteAppClose}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader>アプリの削除</AlertDialogHeader>
            <AlertDialogBody>
              「{app.name}」を削除してもよろしいですか？
              このアプリに保存されたすべてのレコードとフィールドも削除されます。
              この操作は取り消せません。
            </AlertDialogBody>
            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={onDeleteAppClose}>
                キャンセル
              </Button>
              <Button
                colorScheme="red"
                onClick={confirmDeleteApp}
                ml={3}
                isLoading={deleteAppMutation.isPending}
              >
                削除
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
    </Box>
  );
}

// Add Field Modal
function AddFieldModal({
  isOpen,
  onClose,
  appId,
  existingFieldCodes,
}: {
  isOpen: boolean;
  onClose: () => void;
  appId: number;
  existingFieldCodes: string[];
}) {
  const { fields } = useApiClient();
  const queryClient = useQueryClient();
  const toast = useToast();
  const [formData, setFormData] = useState<{
    field_code: string;
    field_name: string;
    field_type: FieldType;
    required: boolean;
    options: FieldOptions;
  }>({
    field_code: "",
    field_name: "",
    field_type: "text",
    required: false,
    options: {},
  });
  const [choicesText, setChoicesText] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  const createMutation = useMutation({
    mutationFn: (data: CreateFieldRequest) => fields.create(appId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["fields", appId] });
      toast({
        title: "フィールドを追加しました",
        status: "success",
        duration: 3000,
      });
      onClose();
      resetForm();
    },
    onError: () => {
      toast({
        title: "フィールドの追加に失敗しました",
        status: "error",
        duration: 3000,
      });
    },
  });

  const resetForm = () => {
    setFormData({
      field_code: "",
      field_name: "",
      field_type: "text",
      required: false,
      options: {},
    });
    setChoicesText("");
    setErrors({});
  };

  const needsChoices = ["select", "multiselect", "radio"].includes(
    formData.field_type
  );

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const newErrors: Record<string, string> = {};

    if (!formData.field_code.trim()) {
      newErrors.field_code = "フィールドコードを入力してください";
    } else if (!/^[a-zA-Z][a-zA-Z0-9_]*$/.test(formData.field_code)) {
      newErrors.field_code =
        "英字で始まり、英数字とアンダースコアのみ使用可能です";
    } else if (existingFieldCodes.includes(formData.field_code)) {
      newErrors.field_code = "このフィールドコードは既に使用されています";
    }

    if (!formData.field_name.trim()) {
      newErrors.field_name = "フィールド名を入力してください";
    }

    if (needsChoices && !choicesText.trim()) {
      newErrors.choices = "選択肢を入力してください";
    }

    setErrors(newErrors);

    if (Object.keys(newErrors).length === 0) {
      const options: FieldOptions = {};
      if (needsChoices) {
        options.choices = choicesText
          .split("\n")
          .map((s) => s.trim())
          .filter(Boolean);
      }

      createMutation.mutate({
        field_code: formData.field_code.trim(),
        field_name: formData.field_name.trim(),
        field_type: formData.field_type,
        required: formData.required,
        options: Object.keys(options).length > 0 ? options : undefined,
      });
    }
  };

  const handleClose = useCallback(() => {
    resetForm();
    onClose();
  }, [onClose]);

  return (
    <Modal isOpen={isOpen} onClose={handleClose}>
      <ModalOverlay />
      <ModalContent>
        <form onSubmit={handleSubmit}>
          <ModalHeader>フィールド追加</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <FormControl isInvalid={!!errors.field_code}>
                <FormLabel>フィールドコード</FormLabel>
                <Input
                  value={formData.field_code}
                  onChange={(e) =>
                    setFormData({ ...formData, field_code: e.target.value })
                  }
                  placeholder="例: customer_name"
                />
                <FormErrorMessage>{errors.field_code}</FormErrorMessage>
              </FormControl>

              <FormControl isInvalid={!!errors.field_name}>
                <FormLabel>フィールド名</FormLabel>
                <Input
                  value={formData.field_name}
                  onChange={(e) =>
                    setFormData({ ...formData, field_name: e.target.value })
                  }
                  placeholder="例: 顧客名"
                />
                <FormErrorMessage>{errors.field_name}</FormErrorMessage>
              </FormControl>

              <FormControl>
                <FormLabel>フィールドタイプ</FormLabel>
                <Select
                  value={formData.field_type}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      field_type: e.target.value as FieldType,
                    })
                  }
                >
                  {Object.entries(FIELD_TYPE_LABELS).map(([value, label]) => (
                    <option key={value} value={value}>
                      {label}
                    </option>
                  ))}
                </Select>
              </FormControl>

              {needsChoices && (
                <FormControl isInvalid={!!errors.choices}>
                  <FormLabel>選択肢（1行に1つ）</FormLabel>
                  <Input
                    as="textarea"
                    value={choicesText}
                    onChange={(e) => setChoicesText(e.target.value)}
                    placeholder="選択肢1&#10;選択肢2&#10;選択肢3"
                    minH="100px"
                  />
                  <FormErrorMessage>{errors.choices}</FormErrorMessage>
                </FormControl>
              )}

              <FormControl display="flex" alignItems="center">
                <FormLabel mb="0">必須項目</FormLabel>
                <Switch
                  isChecked={formData.required}
                  onChange={(e) =>
                    setFormData({ ...formData, required: e.target.checked })
                  }
                />
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={handleClose}>
              キャンセル
            </Button>
            <Button
              type="submit"
              colorScheme="brand"
              isLoading={createMutation.isPending}
            >
              追加
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
}

// Edit Field Modal
function EditFieldModal({
  isOpen,
  onClose,
  appId,
  field,
}: {
  isOpen: boolean;
  onClose: () => void;
  appId: number;
  field: Field;
}) {
  const { fields } = useApiClient();
  const queryClient = useQueryClient();
  const toast = useToast();
  const [formData, setFormData] = useState({
    field_name: field.field_name,
    required: field.required,
  });
  const [choicesText, setChoicesText] = useState(
    field.options?.choices?.join("\n") || ""
  );
  const [error, setError] = useState("");

  const updateMutation = useMutation({
    mutationFn: (data: UpdateFieldRequest) =>
      fields.update(appId, field.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["fields", appId] });
      toast({
        title: "フィールドを更新しました",
        status: "success",
        duration: 3000,
      });
      onClose();
    },
    onError: () => {
      toast({
        title: "フィールドの更新に失敗しました",
        status: "error",
        duration: 3000,
      });
    },
  });

  const needsChoices = ["select", "multiselect", "radio"].includes(
    field.field_type
  );

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!formData.field_name.trim()) {
      setError("フィールド名を入力してください");
      return;
    }

    const options: FieldOptions = {};
    if (needsChoices && choicesText.trim()) {
      options.choices = choicesText
        .split("\n")
        .map((s) => s.trim())
        .filter(Boolean);
    }

    updateMutation.mutate({
      field_name: formData.field_name.trim(),
      required: formData.required,
      options: Object.keys(options).length > 0 ? options : undefined,
    });
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <ModalOverlay />
      <ModalContent>
        <form onSubmit={handleSubmit}>
          <ModalHeader>フィールド編集</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <FormControl>
                <FormLabel>フィールドコード</FormLabel>
                <Input value={field.field_code} isReadOnly bg="gray.50" />
              </FormControl>

              <FormControl>
                <FormLabel>フィールドタイプ</FormLabel>
                <Input
                  value={
                    FIELD_TYPE_LABELS[field.field_type] || field.field_type
                  }
                  isReadOnly
                  bg="gray.50"
                />
              </FormControl>

              <FormControl isInvalid={!!error}>
                <FormLabel>フィールド名</FormLabel>
                <Input
                  value={formData.field_name}
                  onChange={(e) =>
                    setFormData({ ...formData, field_name: e.target.value })
                  }
                />
                <FormErrorMessage>{error}</FormErrorMessage>
              </FormControl>

              {needsChoices && (
                <FormControl>
                  <FormLabel>選択肢（1行に1つ）</FormLabel>
                  <Input
                    as="textarea"
                    value={choicesText}
                    onChange={(e) => setChoicesText(e.target.value)}
                    minH="100px"
                  />
                </FormControl>
              )}

              <FormControl display="flex" alignItems="center">
                <FormLabel mb="0">必須項目</FormLabel>
                <Switch
                  isChecked={formData.required}
                  onChange={(e) =>
                    setFormData({ ...formData, required: e.target.checked })
                  }
                />
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onClose}>
              キャンセル
            </Button>
            <Button
              type="submit"
              colorScheme="brand"
              isLoading={updateMutation.isPending}
            >
              保存
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
}
