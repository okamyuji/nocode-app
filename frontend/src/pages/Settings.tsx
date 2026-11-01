import { useApiClient, useDashboardWidgetsApi } from "@/api";
import { DataSourceList } from "@/components/datasources";
import { useChartConfigs, useDashboardWidgets } from "@/hooks";
import { useAuthStore } from "@/stores";
import type {
  App,
  AppListResponse,
  ChangePasswordRequest,
  ColumnInfo,
  CreateDashboardWidgetRequest,
  CreateFieldRequest,
  CreateUserRequest,
  Field as AppField,
  FieldOptions,
  FieldType,
  UpdateAppRequest,
  UpdateDashboardWidgetRequest,
  UpdateFieldRequest,
  UpdateProfileRequest,
  UpdateUserRequest,
  User,
  WidgetSize,
  WidgetViewType,
} from "@/types";
import { FIELD_TYPE_LABELS, mapDataTypeToFieldType } from "@/types";
import { getAppIcon } from "@/utils";
import { generateFieldCode } from "@/utils/fieldCodeGenerator";
import { toaster } from "@/components/ui/toaster";
import {
  Badge,
  Box,
  Button,
  Card,
  CloseButton,
  Dialog,
  Field,
  Flex,
  Grid,
  Heading,
  Icon,
  IconButton,
  Input,
  InputGroup,
  NativeSelect,
  Portal,
  Skeleton,
  SkeletonText,
  Switch,
  Table,
  Tabs,
  Tag,
  Text,
  Textarea,
  useDisclosure,
  VStack,
} from "@chakra-ui/react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  FiArrowLeft,
  FiBarChart2,
  FiDatabase,
  FiEdit2,
  FiGrid,
  FiList,
  FiPlus,
  FiSearch,
  FiTrash2,
} from "react-icons/fi";
import { Link as RouterLink, useSearchParams } from "react-router-dom";

export function SettingsPage() {
  const { user } = useAuthStore();
  const isAdmin = user?.role === "admin";
  const [searchParams, setSearchParams] = useSearchParams();

  // URL の tab パラメータから初期タブを決める。管理者専用タブは一般ユーザーでは無効にする
  const tabParam = searchParams.get("tab");
  const allowedTabs = isAdmin
    ? ["profile", "password", "users", "apps", "datasources"]
    : ["profile", "password"];
  const initialTab =
    tabParam && allowedTabs.includes(tabParam) ? tabParam : "profile";

  return (
    <Box p={6}>
      <Heading size="lg" mb={6}>
        設定
      </Heading>

      <Tabs.Root
        colorPalette="brand"
        defaultValue={initialTab}
        onValueChange={(e) => setSearchParams({ tab: e.value })}
      >
        <Tabs.List>
          <Tabs.Trigger value="profile">プロフィール</Tabs.Trigger>
          <Tabs.Trigger value="password">パスワード変更</Tabs.Trigger>
          {isAdmin && <Tabs.Trigger value="users">ユーザー管理</Tabs.Trigger>}
          {isAdmin && <Tabs.Trigger value="apps">アプリ設定</Tabs.Trigger>}
          {isAdmin && (
            <Tabs.Trigger value="datasources">データソース</Tabs.Trigger>
          )}
        </Tabs.List>

        <Tabs.Content value="profile">
          <ProfileSettings />
        </Tabs.Content>
        <Tabs.Content value="password">
          <PasswordSettings />
        </Tabs.Content>
        {isAdmin && (
          <Tabs.Content value="users">
            <UserManagement />
          </Tabs.Content>
        )}
        {isAdmin && (
          <Tabs.Content value="apps">
            <AppSettings />
          </Tabs.Content>
        )}
        {isAdmin && (
          <Tabs.Content value="datasources">
            <DataSourceList />
          </Tabs.Content>
        )}
      </Tabs.Root>
    </Box>
  );
}

// Profile Settings Component
function ProfileSettings() {
  const { user, setUser } = useAuthStore();
  const { profile } = useApiClient();
  const [name, setName] = useState(user?.name || "");
  const [error, setError] = useState("");

  const updateProfileMutation = useMutation({
    mutationFn: (data: UpdateProfileRequest) => profile.updateProfile(data),
    onSuccess: (updatedUser) => {
      setUser(updatedUser);
      toaster.create({
        title: "プロフィールを更新しました",
        type: "success",
        duration: 3000,
      });
    },
    onError: () => {
      toaster.create({
        title: "プロフィールの更新に失敗しました",
        type: "error",
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
    <Card.Root maxW="md">
      <Card.Header>
        <Heading size="md">プロフィール設定</Heading>
      </Card.Header>
      <Card.Body>
        <form onSubmit={handleSubmit}>
          <VStack gap={4}>
            <Field.Root>
              <Field.Label>メールアドレス</Field.Label>
              <Input value={user?.email || ""} readOnly bg="gray.50" />
            </Field.Root>

            <Field.Root>
              <Field.Label>ロール</Field.Label>
              <Input
                value={user?.role === "admin" ? "管理者" : "一般ユーザー"}
                readOnly
                bg="gray.50"
              />
            </Field.Root>

            <Field.Root invalid={!!error}>
              <Field.Label>名前</Field.Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="名前を入力"
              />
              <Field.ErrorText>{error}</Field.ErrorText>
            </Field.Root>

            <Button
              type="submit"
              colorPalette="brand"
              w="full"
              loading={updateProfileMutation.isPending}
            >
              保存
            </Button>
          </VStack>
        </form>
      </Card.Body>
    </Card.Root>
  );
}

// Password Settings Component
function PasswordSettings() {
  const { profile } = useApiClient();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  const changePasswordMutation = useMutation({
    mutationFn: (data: ChangePasswordRequest) => profile.changePassword(data),
    onSuccess: () => {
      toaster.create({
        title: "パスワードを変更しました",
        type: "success",
        duration: 3000,
      });
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    },
    onError: () => {
      toaster.create({
        title: "パスワードの変更に失敗しました",
        description: "現在のパスワードが正しくない可能性があります",
        type: "error",
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
    <Card.Root maxW="md">
      <Card.Header>
        <Heading size="md">パスワード変更</Heading>
      </Card.Header>
      <Card.Body>
        <form onSubmit={handleSubmit}>
          <VStack gap={4}>
            <Field.Root invalid={!!errors.currentPassword}>
              <Field.Label>現在のパスワード</Field.Label>
              <Input
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
              />
              <Field.ErrorText>{errors.currentPassword}</Field.ErrorText>
            </Field.Root>

            <Field.Root invalid={!!errors.newPassword}>
              <Field.Label>新しいパスワード</Field.Label>
              <Input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
              />
              <Field.ErrorText>{errors.newPassword}</Field.ErrorText>
            </Field.Root>

            <Field.Root invalid={!!errors.confirmPassword}>
              <Field.Label>新しいパスワード（確認）</Field.Label>
              <Input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
              />
              <Field.ErrorText>{errors.confirmPassword}</Field.ErrorText>
            </Field.Root>

            <Button
              type="submit"
              colorPalette="brand"
              w="full"
              loading={changePasswordMutation.isPending}
            >
              パスワードを変更
            </Button>
          </VStack>
        </form>
      </Card.Body>
    </Card.Root>
  );
}

// User Management Component (Admin Only)
function UserManagement() {
  const { users } = useApiClient();
  const { user: currentUser } = useAuthStore();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const limit = 20;

  const {
    open: isCreateOpen,
    onOpen: onCreateOpen,
    onClose: onCreateClose,
  } = useDisclosure();
  const {
    open: isEditOpen,
    onOpen: onEditOpen,
    onClose: onEditClose,
  } = useDisclosure();
  const {
    open: isDeleteOpen,
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
      toaster.create({
        title: "ユーザーを削除しました",
        type: "success",
        duration: 3000,
      });
      onDeleteClose();
    },
    onError: () => {
      toaster.create({
        title: "ユーザーの削除に失敗しました",
        type: "error",
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
        <Button colorPalette="brand" onClick={onCreateOpen}>
          <FiPlus />
          新規ユーザー
        </Button>
      </Flex>

      <Card.Root>
        <Card.Body p={0}>
          <Table.Root variant="line">
            <Table.Header>
              <Table.Row>
                <Table.ColumnHeader>名前</Table.ColumnHeader>
                <Table.ColumnHeader>メールアドレス</Table.ColumnHeader>
                <Table.ColumnHeader>ロール</Table.ColumnHeader>
                <Table.ColumnHeader>作成日</Table.ColumnHeader>
                <Table.ColumnHeader w="100px">操作</Table.ColumnHeader>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {isLoading ? (
                <Table.Row>
                  <Table.Cell colSpan={5} textAlign="center">
                    読み込み中...
                  </Table.Cell>
                </Table.Row>
              ) : data?.users?.length === 0 ? (
                <Table.Row>
                  <Table.Cell colSpan={5} textAlign="center">
                    ユーザーがいません
                  </Table.Cell>
                </Table.Row>
              ) : (
                data?.users?.map((user) => (
                  <Table.Row key={user.id}>
                    <Table.Cell>{user.name}</Table.Cell>
                    <Table.Cell>{user.email}</Table.Cell>
                    <Table.Cell>
                      <Badge
                        colorPalette={user.role === "admin" ? "purple" : "gray"}
                      >
                        {user.role === "admin" ? "管理者" : "一般"}
                      </Badge>
                    </Table.Cell>
                    <Table.Cell>
                      {new Date(user.created_at).toLocaleDateString("ja-JP")}
                    </Table.Cell>
                    <Table.Cell>
                      <Flex gap={2}>
                        <IconButton
                          aria-label="編集"
                          size="sm"
                          variant="ghost"
                          onClick={() => handleEdit(user)}
                        >
                          <FiEdit2 />
                        </IconButton>
                        <IconButton
                          aria-label="削除"
                          size="sm"
                          variant="ghost"
                          colorPalette="red"
                          disabled={user.id === currentUser?.id}
                          onClick={() => handleDelete(user)}
                        >
                          <FiTrash2 />
                        </IconButton>
                      </Flex>
                    </Table.Cell>
                  </Table.Row>
                ))
              )}
            </Table.Body>
          </Table.Root>
        </Card.Body>
      </Card.Root>

      {data?.pagination && data.pagination.total_pages > 1 && (
        <Flex justify="center" mt={4} gap={2}>
          <Button
            size="sm"
            disabled={page === 1}
            onClick={() => setPage((p) => p - 1)}
          >
            前へ
          </Button>
          <Text alignSelf="center">
            {page} / {data.pagination.total_pages}
          </Text>
          <Button
            size="sm"
            disabled={page === data.pagination.total_pages}
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
      <Dialog.Root
        role="alertdialog"
        open={isDeleteOpen}
        onOpenChange={(e) => !e.open && onDeleteClose()}
        initialFocusEl={() => cancelRef.current}
      >
        <Portal>
          <Dialog.Backdrop />
          <Dialog.Positioner>
            <Dialog.Content>
              <Dialog.Header>
                <Dialog.Title>ユーザーの削除</Dialog.Title>
              </Dialog.Header>
              <Dialog.Body>
                「{selectedUser?.name}」を削除してもよろしいですか？
                この操作は取り消せません。
              </Dialog.Body>
              <Dialog.Footer>
                <Button ref={cancelRef} onClick={onDeleteClose}>
                  キャンセル
                </Button>
                <Button
                  colorPalette="red"
                  onClick={confirmDelete}
                  ml={3}
                  loading={deleteMutation.isPending}
                >
                  削除
                </Button>
              </Dialog.Footer>
            </Dialog.Content>
          </Dialog.Positioner>
        </Portal>
      </Dialog.Root>
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
      toaster.create({
        title: "ユーザーを作成しました",
        type: "success",
        duration: 3000,
      });
      onClose();
      setFormData({ email: "", password: "", name: "", role: "user" });
    },
    onError: () => {
      toaster.create({
        title: "ユーザーの作成に失敗しました",
        description: "メールアドレスが既に使用されている可能性があります",
        type: "error",
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
    <Dialog.Root open={isOpen} onOpenChange={(e) => !e.open && onClose()}>
      <Portal>
        <Dialog.Backdrop />
        <Dialog.Positioner>
          <Dialog.Content>
            <form onSubmit={handleSubmit}>
              <Dialog.Header>
                <Dialog.Title>新規ユーザー作成</Dialog.Title>
              </Dialog.Header>
              <Dialog.CloseTrigger asChild>
                <CloseButton size="sm" />
              </Dialog.CloseTrigger>
              <Dialog.Body>
                <VStack gap={4}>
                  <Field.Root invalid={!!errors.email}>
                    <Field.Label>メールアドレス</Field.Label>
                    <Input
                      type="email"
                      value={formData.email}
                      onChange={(e) =>
                        setFormData({ ...formData, email: e.target.value })
                      }
                    />
                    <Field.ErrorText>{errors.email}</Field.ErrorText>
                  </Field.Root>

                  <Field.Root invalid={!!errors.password}>
                    <Field.Label>パスワード</Field.Label>
                    <Input
                      type="password"
                      value={formData.password}
                      onChange={(e) =>
                        setFormData({ ...formData, password: e.target.value })
                      }
                    />
                    <Field.ErrorText>{errors.password}</Field.ErrorText>
                  </Field.Root>

                  <Field.Root invalid={!!errors.name}>
                    <Field.Label>名前</Field.Label>
                    <Input
                      value={formData.name}
                      onChange={(e) =>
                        setFormData({ ...formData, name: e.target.value })
                      }
                    />
                    <Field.ErrorText>{errors.name}</Field.ErrorText>
                  </Field.Root>

                  <Field.Root>
                    <Field.Label>ロール</Field.Label>
                    <NativeSelect.Root>
                      <NativeSelect.Field
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
                      </NativeSelect.Field>
                      <NativeSelect.Indicator />
                    </NativeSelect.Root>
                  </Field.Root>
                </VStack>
              </Dialog.Body>
              <Dialog.Footer>
                <Button variant="ghost" mr={3} onClick={onClose}>
                  キャンセル
                </Button>
                <Button
                  type="submit"
                  colorPalette="brand"
                  loading={createMutation.isPending}
                >
                  作成
                </Button>
              </Dialog.Footer>
            </form>
          </Dialog.Content>
        </Dialog.Positioner>
      </Portal>
    </Dialog.Root>
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
  const [formData, setFormData] = useState({
    name: user.name,
    role: user.role as "admin" | "user",
  });
  const [error, setError] = useState("");

  const updateMutation = useMutation({
    mutationFn: (data: UpdateUserRequest) => users.update(user.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toaster.create({
        title: "ユーザーを更新しました",
        type: "success",
        duration: 3000,
      });
      onClose();
    },
    onError: () => {
      toaster.create({
        title: "ユーザーの更新に失敗しました",
        type: "error",
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
    <Dialog.Root open={isOpen} onOpenChange={(e) => !e.open && onClose()}>
      <Portal>
        <Dialog.Backdrop />
        <Dialog.Positioner>
          <Dialog.Content>
            <form onSubmit={handleSubmit}>
              <Dialog.Header>
                <Dialog.Title>ユーザー編集</Dialog.Title>
              </Dialog.Header>
              <Dialog.CloseTrigger asChild>
                <CloseButton size="sm" />
              </Dialog.CloseTrigger>
              <Dialog.Body>
                <VStack gap={4}>
                  <Field.Root>
                    <Field.Label>メールアドレス</Field.Label>
                    <Input value={user.email} readOnly bg="gray.50" />
                  </Field.Root>

                  <Field.Root invalid={!!error}>
                    <Field.Label>名前</Field.Label>
                    <Input
                      value={formData.name}
                      onChange={(e) =>
                        setFormData({ ...formData, name: e.target.value })
                      }
                    />
                    <Field.ErrorText>{error}</Field.ErrorText>
                  </Field.Root>

                  <Field.Root>
                    <Field.Label>ロール</Field.Label>
                    <NativeSelect.Root disabled={isSelf}>
                      <NativeSelect.Field
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
                      </NativeSelect.Field>
                      <NativeSelect.Indicator />
                    </NativeSelect.Root>
                    {isSelf && (
                      <Text fontSize="sm" color="gray.500" mt={1}>
                        自分のロールは変更できません
                      </Text>
                    )}
                  </Field.Root>
                </VStack>
              </Dialog.Body>
              <Dialog.Footer>
                <Button variant="ghost" mr={3} onClick={onClose}>
                  キャンセル
                </Button>
                <Button
                  type="submit"
                  colorPalette="brand"
                  loading={updateMutation.isPending}
                >
                  保存
                </Button>
              </Dialog.Footer>
            </form>
          </Dialog.Content>
        </Dialog.Positioner>
      </Portal>
    </Dialog.Root>
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

      <InputGroup
        maxW="400px"
        mb={4}
        startElement={<Icon as={FiSearch} color="gray.400" />}
      >
        <Input
          placeholder="アプリを検索..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
        />
      </InputGroup>

      {isLoading ? (
        <Grid templateColumns="repeat(auto-fill, minmax(280px, 1fr))" gap={4}>
          {[1, 2, 3, 4].map((i) => (
            <Card.Root key={i}>
              <Card.Body>
                <Skeleton height="20px" width="60%" mb={2} />
                <SkeletonText noOfLines={2} />
              </Card.Body>
            </Card.Root>
          ))}
        </Grid>
      ) : filteredApps.length === 0 ? (
        <Card.Root>
          <Card.Body textAlign="center" py={10}>
            <Icon as={FiDatabase} boxSize={12} color="gray.300" mb={4} />
            <Text fontSize="lg" color="gray.500">
              {searchQuery
                ? "検索条件に一致するアプリがありません"
                : "アプリがありません"}
            </Text>
          </Card.Body>
        </Card.Root>
      ) : (
        <Grid templateColumns="repeat(auto-fill, minmax(280px, 1fr))" gap={4}>
          {filteredApps.map((app) => (
            <Card.Root
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
              <Card.Body>
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
                    <Icon
                      as={getAppIcon(app.icon)}
                      color="brand.500"
                      boxSize={5}
                    />
                  </Flex>
                  <Box flex={1} minW={0}>
                    <Flex align="center" gap={2} mb={1}>
                      <Heading size="sm" lineClamp={1}>
                        {app.name}
                      </Heading>
                      <Badge colorPalette="brand" fontSize="xs">
                        {app.field_count} フィールド
                      </Badge>
                    </Flex>
                    <Text fontSize="sm" color="gray.600" lineClamp={2}>
                      {app.description || "No description"}
                    </Text>
                  </Box>
                </Flex>
              </Card.Body>
            </Card.Root>
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
  const dashboardWidgetsApi = useDashboardWidgetsApi();
  const queryClient = useQueryClient();
  const cancelRef = useRef<HTMLButtonElement>(null);

  // Form state for basic info
  const [name, setName] = useState(app.name);
  const [description, setDescription] = useState(app.description || "");
  const [icon, setIcon] = useState(app.icon || "default");

  // ダッシュボードウィジェットの取得
  const { data: widgetsData } = useDashboardWidgets(false);
  const existingWidget = widgetsData?.widgets.find((w) => w.app_id === app.id);

  // ダッシュボード表示設定のステート
  const [showOnDashboard, setShowOnDashboard] = useState(false);
  const [viewType, setViewType] = useState<WidgetViewType>("table");
  const [widgetSize, setWidgetSize] = useState<WidgetSize>("medium");
  const [selectedChartConfigId, setSelectedChartConfigId] = useState<
    number | null
  >(null);

  // チャート設定一覧を取得
  const { data: chartConfigsData } = useChartConfigs(app.id);
  const chartConfigs = chartConfigsData?.configs || [];

  // 既存のウィジェット設定をステートに反映
  useEffect(() => {
    if (existingWidget) {
      setShowOnDashboard(existingWidget.is_visible);
      setViewType(existingWidget.view_type);
      setWidgetSize(existingWidget.widget_size);
      // チャート設定IDを復元
      const chartConfigId = existingWidget.config?.chart_config_id as
        | number
        | undefined;
      setSelectedChartConfigId(chartConfigId || null);
    } else {
      setShowOnDashboard(false);
      setViewType("table");
      setWidgetSize("medium");
      setSelectedChartConfigId(null);
    }
  }, [existingWidget]);

  // propsのappが変更された場合にローカルstateを同期
  useEffect(() => {
    setName(app.name);
    setDescription(app.description || "");
    setIcon(app.icon || "default");
  }, [app]);

  // Modal states
  const {
    open: isDeleteAppOpen,
    onOpen: onDeleteAppOpen,
    onClose: onDeleteAppClose,
  } = useDisclosure();
  const {
    open: isAddFieldOpen,
    onOpen: onAddFieldOpen,
    onClose: onAddFieldClose,
  } = useDisclosure();
  const {
    open: isEditFieldOpen,
    onOpen: onEditFieldOpen,
    onClose: onEditFieldClose,
  } = useDisclosure();
  const {
    open: isDeleteFieldOpen,
    onOpen: onDeleteFieldOpen,
    onClose: onDeleteFieldClose,
  } = useDisclosure();

  const [selectedField, setSelectedField] = useState<AppField | null>(null);

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
      queryClient.setQueriesData<AppListResponse>(
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

      toaster.create({
        title: "アプリを更新しました",
        type: "success",
        duration: 3000,
      });
    },
    onError: () => {
      toaster.create({
        title: "アプリの更新に失敗しました",
        type: "error",
        duration: 3000,
      });
    },
  });

  // Delete app mutation
  const deleteAppMutation = useMutation({
    mutationFn: () => apps.delete(app.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["apps"] });
      toaster.create({
        title: "アプリを削除しました",
        type: "success",
        duration: 3000,
      });
      onBack();
    },
    onError: () => {
      toaster.create({
        title: "アプリの削除に失敗しました",
        type: "error",
        duration: 3000,
      });
    },
  });

  // ダッシュボードウィジェット作成
  const createWidgetMutation = useMutation({
    mutationFn: (data: CreateDashboardWidgetRequest) =>
      dashboardWidgetsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dashboard", "widgets"] });
      toaster.create({
        title: "ダッシュボード設定を保存しました",
        type: "success",
        duration: 3000,
      });
    },
    onError: () => {
      toaster.create({
        title: "ダッシュボード設定の保存に失敗しました",
        type: "error",
        duration: 3000,
      });
    },
  });

  // ダッシュボードウィジェット更新
  const updateWidgetMutation = useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: number;
      data: UpdateDashboardWidgetRequest;
    }) => dashboardWidgetsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dashboard", "widgets"] });
      toaster.create({
        title: "ダッシュボード設定を保存しました",
        type: "success",
        duration: 3000,
      });
    },
    onError: () => {
      toaster.create({
        title: "ダッシュボード設定の保存に失敗しました",
        type: "error",
        duration: 3000,
      });
    },
  });

  // ダッシュボードウィジェット削除
  const deleteWidgetMutation = useMutation({
    mutationFn: (id: number) => dashboardWidgetsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dashboard", "widgets"] });
      toaster.create({
        title: "ダッシュボードから削除しました",
        type: "success",
        duration: 3000,
      });
    },
    onError: () => {
      toaster.create({
        title: "削除に失敗しました",
        type: "error",
        duration: 3000,
      });
    },
  });

  // Delete field mutation
  const deleteFieldMutation = useMutation({
    mutationFn: (fieldId: number) => fields.delete(app.id, fieldId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["fields", app.id] });
      toaster.create({
        title: "フィールドを削除しました",
        type: "success",
        duration: 3000,
      });
      onDeleteFieldClose();
    },
    onError: () => {
      toaster.create({
        title: "フィールドの削除に失敗しました",
        type: "error",
        duration: 3000,
      });
    },
  });

  const handleSaveBasicInfo = () => {
    if (!name.trim()) {
      toaster.create({
        title: "アプリ名を入力してください",
        type: "error",
        duration: 3000,
      });
      return;
    }
    updateAppMutation.mutate({ name: name.trim(), description, icon });
  };

  const handleEditField = (field: AppField) => {
    setSelectedField(field);
    onEditFieldOpen();
  };

  const handleDeleteField = (field: AppField) => {
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

  // ダッシュボード設定を保存
  const handleSaveDashboardSettings = () => {
    // グラフ選択時のconfig設定
    const widgetConfig =
      viewType === "chart" && selectedChartConfigId
        ? { chart_config_id: selectedChartConfigId }
        : undefined;

    if (showOnDashboard) {
      if (existingWidget) {
        // 既存のウィジェットを更新
        updateWidgetMutation.mutate({
          id: existingWidget.id,
          data: {
            view_type: viewType,
            widget_size: widgetSize,
            is_visible: true,
            config: widgetConfig,
          },
        });
      } else {
        // 新規ウィジェットを作成
        createWidgetMutation.mutate({
          app_id: app.id,
          view_type: viewType,
          widget_size: widgetSize,
          is_visible: true,
          config: widgetConfig,
        });
      }
    } else {
      if (existingWidget) {
        // ウィジェットを非表示に
        updateWidgetMutation.mutate({
          id: existingWidget.id,
          data: { is_visible: false },
        });
      }
    }
  };

  const isDashboardSaving =
    createWidgetMutation.isPending ||
    updateWidgetMutation.isPending ||
    deleteWidgetMutation.isPending;

  return (
    <Box>
      <Button variant="ghost" mb={4} onClick={onBack}>
        <FiArrowLeft />
        アプリ一覧に戻る
      </Button>

      <Flex justify="space-between" align="center" mb={6}>
        <Heading size="md">{app.name} の設定</Heading>
        <Button colorPalette="red" variant="outline" onClick={onDeleteAppOpen}>
          <FiTrash2 />
          アプリを削除
        </Button>
      </Flex>

      <Grid templateColumns={{ base: "1fr", lg: "1fr 1fr" }} gap={6}>
        {/* Basic Info Card */}
        <Card.Root>
          <Card.Header>
            <Heading size="sm">基本情報</Heading>
          </Card.Header>
          <Card.Body>
            <VStack gap={4}>
              <Field.Root>
                <Field.Label>アプリ名</Field.Label>
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="アプリ名を入力"
                />
              </Field.Root>

              <Field.Root>
                <Field.Label>説明</Field.Label>
                <Input
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="アプリの説明を入力"
                />
              </Field.Root>

              <Field.Root>
                <Field.Label>アイコン</Field.Label>
                <NativeSelect.Root>
                  <NativeSelect.Field
                    value={icon}
                    onChange={(e) => setIcon(e.target.value)}
                  >
                    <option value="default">デフォルト</option>
                    <option value="grid">グリッド</option>
                    <option value="list">リスト</option>
                    <option value="calendar">カレンダー</option>
                    <option value="database">データベース</option>
                  </NativeSelect.Field>
                  <NativeSelect.Indicator />
                </NativeSelect.Root>
              </Field.Root>

              <Button
                colorPalette="brand"
                w="full"
                onClick={handleSaveBasicInfo}
                loading={updateAppMutation.isPending}
              >
                保存
              </Button>
            </VStack>
          </Card.Body>
        </Card.Root>

        {/* Dashboard Settings Card */}
        <Card.Root>
          <Card.Header>
            <Heading size="sm">ダッシュボード表示設定</Heading>
          </Card.Header>
          <Card.Body>
            <VStack gap={4} align="stretch">
              <Field.Root display="flex" alignItems="center">
                <Field.Label mb="0">ダッシュボードに表示</Field.Label>
                <Switch.Root
                  checked={showOnDashboard}
                  onCheckedChange={(e) => setShowOnDashboard(e.checked)}
                >
                  <Switch.HiddenInput />
                  <Switch.Control>
                    <Switch.Thumb />
                  </Switch.Control>
                </Switch.Root>
              </Field.Root>

              {showOnDashboard && (
                <>
                  <Field.Root>
                    <Field.Label>表示形式</Field.Label>
                    <Grid templateColumns="repeat(3, 1fr)" gap={2}>
                      <Button
                        variant={viewType === "table" ? "solid" : "outline"}
                        colorPalette={viewType === "table" ? "brand" : "gray"}
                        onClick={() => setViewType("table")}
                        size="sm"
                      >
                        <Icon as={FiGrid} />
                        テーブル
                      </Button>
                      <Button
                        variant={viewType === "list" ? "solid" : "outline"}
                        colorPalette={viewType === "list" ? "brand" : "gray"}
                        onClick={() => setViewType("list")}
                        size="sm"
                      >
                        <Icon as={FiList} />
                        リスト
                      </Button>
                      <Button
                        variant={viewType === "chart" ? "solid" : "outline"}
                        colorPalette={viewType === "chart" ? "brand" : "gray"}
                        onClick={() => setViewType("chart")}
                        size="sm"
                      >
                        <Icon as={FiBarChart2} />
                        グラフ
                      </Button>
                    </Grid>
                  </Field.Root>

                  {viewType === "chart" && (
                    <Field.Root>
                      <Field.Label>表示するグラフ</Field.Label>
                      {chartConfigs.length === 0 ? (
                        <VStack gap={2} align="start">
                          <Text fontSize="sm" color="gray.500">
                            グラフ設定がありません。
                          </Text>
                          <Button asChild size="sm" colorPalette="brand">
                            <RouterLink to={`/apps/${app.id}/charts`}>
                              <Icon as={FiBarChart2} />
                              グラフ設定を作成する
                            </RouterLink>
                          </Button>
                        </VStack>
                      ) : (
                        <NativeSelect.Root>
                          <NativeSelect.Field
                            value={selectedChartConfigId || ""}
                            onChange={(e) =>
                              setSelectedChartConfigId(
                                e.target.value ? Number(e.target.value) : null
                              )
                            }
                            placeholder="グラフを選択"
                          >
                            {chartConfigs.map((config) => (
                              <option key={config.id} value={config.id}>
                                {config.name}
                              </option>
                            ))}
                          </NativeSelect.Field>
                          <NativeSelect.Indicator />
                        </NativeSelect.Root>
                      )}
                    </Field.Root>
                  )}

                  <Field.Root>
                    <Field.Label>ウィジェットサイズ</Field.Label>
                    <NativeSelect.Root>
                      <NativeSelect.Field
                        value={widgetSize}
                        onChange={(e) =>
                          setWidgetSize(e.target.value as WidgetSize)
                        }
                      >
                        <option value="small">小</option>
                        <option value="medium">中</option>
                        <option value="large">大</option>
                      </NativeSelect.Field>
                      <NativeSelect.Indicator />
                    </NativeSelect.Root>
                  </Field.Root>
                </>
              )}

              <Button
                colorPalette="brand"
                w="full"
                onClick={handleSaveDashboardSettings}
                loading={isDashboardSaving}
              >
                ダッシュボード設定を保存
              </Button>
            </VStack>
          </Card.Body>
        </Card.Root>

        {/* Fields Card */}
        <Card.Root>
          <Card.Header>
            <Flex justify="space-between" align="center">
              <Heading size="sm">フィールド管理</Heading>
              <Button size="sm" colorPalette="brand" onClick={onAddFieldOpen}>
                <FiPlus />
                フィールド追加
              </Button>
            </Flex>
          </Card.Header>
          <Card.Body p={0}>
            {fieldsLoading ? (
              <Box p={4}>
                <SkeletonText noOfLines={5} />
              </Box>
            ) : !fieldsData?.fields?.length ? (
              <Box p={4} textAlign="center">
                <Text color="gray.500">フィールドがありません</Text>
              </Box>
            ) : (
              <Table.Root variant="line" size="sm">
                <Table.Header>
                  <Table.Row>
                    <Table.ColumnHeader>名前</Table.ColumnHeader>
                    <Table.ColumnHeader>タイプ</Table.ColumnHeader>
                    <Table.ColumnHeader>必須</Table.ColumnHeader>
                    <Table.ColumnHeader w="80px">操作</Table.ColumnHeader>
                  </Table.Row>
                </Table.Header>
                <Table.Body>
                  {fieldsData.fields.map((field) => (
                    <Table.Row key={field.id}>
                      <Table.Cell>
                        <VStack align="start" gap={0}>
                          <Text fontWeight="medium">{field.field_name}</Text>
                          <Text fontSize="xs" color="gray.500">
                            {field.field_code}
                          </Text>
                        </VStack>
                      </Table.Cell>
                      <Table.Cell>
                        <Tag.Root size="sm" colorPalette="blue">
                          {FIELD_TYPE_LABELS[field.field_type] ||
                            field.field_type}
                        </Tag.Root>
                      </Table.Cell>
                      <Table.Cell>
                        {field.required ? (
                          <Badge colorPalette="red">必須</Badge>
                        ) : (
                          <Text color="gray.400">-</Text>
                        )}
                      </Table.Cell>
                      <Table.Cell>
                        <Flex gap={1}>
                          <IconButton
                            aria-label="編集"
                            size="xs"
                            variant="ghost"
                            onClick={() => handleEditField(field)}
                          >
                            <FiEdit2 />
                          </IconButton>
                          <IconButton
                            aria-label="削除"
                            size="xs"
                            variant="ghost"
                            colorPalette="red"
                            onClick={() => handleDeleteField(field)}
                          >
                            <FiTrash2 />
                          </IconButton>
                        </Flex>
                      </Table.Cell>
                    </Table.Row>
                  ))}
                </Table.Body>
              </Table.Root>
            )}
          </Card.Body>
        </Card.Root>
      </Grid>

      {/* Add Field Modal */}
      <AddFieldModal
        isOpen={isAddFieldOpen}
        onClose={onAddFieldClose}
        app={app}
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
      <Dialog.Root
        role="alertdialog"
        open={isDeleteFieldOpen}
        onOpenChange={(e) => !e.open && onDeleteFieldClose()}
        initialFocusEl={() => cancelRef.current}
      >
        <Portal>
          <Dialog.Backdrop />
          <Dialog.Positioner>
            <Dialog.Content>
              <Dialog.Header>
                <Dialog.Title>フィールドの削除</Dialog.Title>
              </Dialog.Header>
              <Dialog.Body>
                「{selectedField?.field_name}」を削除してもよろしいですか？
                このフィールドに保存されたデータも削除されます。
              </Dialog.Body>
              <Dialog.Footer>
                <Button ref={cancelRef} onClick={onDeleteFieldClose}>
                  キャンセル
                </Button>
                <Button
                  colorPalette="red"
                  onClick={confirmDeleteField}
                  ml={3}
                  loading={deleteFieldMutation.isPending}
                >
                  削除
                </Button>
              </Dialog.Footer>
            </Dialog.Content>
          </Dialog.Positioner>
        </Portal>
      </Dialog.Root>

      {/* Delete App Confirmation */}
      <Dialog.Root
        role="alertdialog"
        open={isDeleteAppOpen}
        onOpenChange={(e) => !e.open && onDeleteAppClose()}
        initialFocusEl={() => cancelRef.current}
      >
        <Portal>
          <Dialog.Backdrop />
          <Dialog.Positioner>
            <Dialog.Content>
              <Dialog.Header>
                <Dialog.Title>アプリの削除</Dialog.Title>
              </Dialog.Header>
              <Dialog.Body>
                「{app.name}」を削除してもよろしいですか？
                このアプリに保存されたすべてのレコードとフィールドも削除されます。
                この操作は取り消せません。
              </Dialog.Body>
              <Dialog.Footer>
                <Button ref={cancelRef} onClick={onDeleteAppClose}>
                  キャンセル
                </Button>
                <Button
                  colorPalette="red"
                  onClick={confirmDeleteApp}
                  ml={3}
                  loading={deleteAppMutation.isPending}
                >
                  削除
                </Button>
              </Dialog.Footer>
            </Dialog.Content>
          </Dialog.Positioner>
        </Portal>
      </Dialog.Root>
    </Box>
  );
}

// Add Field Modal
function AddFieldModal({
  isOpen,
  onClose,
  app,
  existingFieldCodes,
}: {
  isOpen: boolean;
  onClose: () => void;
  app: App;
  existingFieldCodes: string[];
}) {
  const { fields, dataSources } = useApiClient();
  const queryClient = useQueryClient();

  // 新規データソース用のフォームステート
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

  // 外部データソース用のステート
  const [selectedColumns, setSelectedColumns] = useState<Set<string>>(
    new Set()
  );

  // 外部データソースのカラム一覧を取得
  const { data: columnsData, isLoading: columnsLoading } = useQuery({
    queryKey: ["columns", app.data_source_id, app.source_table_name],
    queryFn: () =>
      dataSources.getColumns(app.data_source_id!, app.source_table_name!),
    enabled: app.is_external && !!app.data_source_id && !!app.source_table_name,
  });

  // 追加可能なカラム（既存フィールドに追加済みのものを除外）
  const availableColumns = useMemo((): ColumnInfo[] => {
    if (!columnsData?.columns) return [];
    return columnsData.columns.filter(
      (col: ColumnInfo) => !existingFieldCodes.includes(col.name)
    );
  }, [columnsData?.columns, existingFieldCodes]);

  const createMutation = useMutation({
    mutationFn: (data: CreateFieldRequest) => fields.create(app.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["fields", app.id] });
      toaster.create({
        title: "フィールドを追加しました",
        type: "success",
        duration: 3000,
      });
      onClose();
      resetForm();
    },
    onError: () => {
      toaster.create({
        title: "フィールドの追加に失敗しました",
        type: "error",
        duration: 3000,
      });
    },
  });

  // 外部データソース用の一括作成ミューテーション
  const bulkCreateMutation = useMutation({
    mutationFn: async (columnNames: string[]) => {
      // 使用済みフィールドコードを追跡（既存 + 新規追加分）
      const usedFieldCodes = new Set(existingFieldCodes);
      // 現在のカラム一覧からインデックスを取得
      const allColumns = columnsData?.columns || [];

      // 選択されたカラムを順番に追加
      for (const colName of columnNames) {
        const columnIndex = allColumns.findIndex(
          (c: ColumnInfo) => c.name === colName
        );
        const column = allColumns[columnIndex];

        if (column) {
          // フィールドコードを生成（日本語を含む場合は英数字に変換）
          let fieldCode = generateFieldCode(
            colName,
            existingFieldCodes.length + columnIndex
          );
          // 重複を避けるためにサフィックスを付ける
          let suffix = 1;
          const baseCode = fieldCode;
          while (usedFieldCodes.has(fieldCode)) {
            suffix++;
            fieldCode = `${baseCode}_${suffix}`;
          }
          usedFieldCodes.add(fieldCode);

          await fields.create(app.id, {
            field_code: fieldCode,
            field_name: colName, // 表示名はカラム名をそのまま使用
            field_type: mapDataTypeToFieldType(column.data_type) as FieldType,
            source_column_name: colName, // 外部データソースのカラム名を設定
            required: !column.is_nullable,
          });
        }
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["fields", app.id] });
      toaster.create({
        title: "フィールドを追加しました",
        type: "success",
        duration: 3000,
      });
      onClose();
      resetForm();
    },
    onError: () => {
      toaster.create({
        title: "フィールドの追加に失敗しました",
        type: "error",
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
    setSelectedColumns(new Set());
  };

  const needsChoices = ["select", "multiselect", "radio"].includes(
    formData.field_type
  );

  // 新規データソース用のサブミット
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

  // 外部データソース用のサブミット
  const handleExternalSubmit = () => {
    if (selectedColumns.size === 0) {
      toaster.create({
        title: "カラムを選択してください",
        type: "warning",
        duration: 3000,
      });
      return;
    }
    bulkCreateMutation.mutate(Array.from(selectedColumns));
  };

  const handleColumnToggle = (columnName: string) => {
    setSelectedColumns((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(columnName)) {
        newSet.delete(columnName);
      } else {
        newSet.add(columnName);
      }
      return newSet;
    });
  };

  const handleSelectAll = () => {
    if (selectedColumns.size === availableColumns.length) {
      setSelectedColumns(new Set());
    } else {
      setSelectedColumns(new Set(availableColumns.map((col) => col.name)));
    }
  };

  const handleClose = useCallback(() => {
    resetForm();
    onClose();
  }, [onClose]);

  // 外部データソースの場合
  if (app.is_external) {
    return (
      <Dialog.Root
        open={isOpen}
        onOpenChange={(e) => !e.open && handleClose()}
        size="xl"
      >
        <Portal>
          <Dialog.Backdrop />
          <Dialog.Positioner>
            <Dialog.Content>
              <Dialog.Header>
                <Dialog.Title>カラムからフィールドを追加</Dialog.Title>
              </Dialog.Header>
              <Dialog.CloseTrigger asChild>
                <CloseButton size="sm" />
              </Dialog.CloseTrigger>
              <Dialog.Body>
                {columnsLoading ? (
                  <VStack gap={4}>
                    <Skeleton height="40px" width="100%" />
                    <Skeleton height="40px" width="100%" />
                    <Skeleton height="40px" width="100%" />
                  </VStack>
                ) : availableColumns.length === 0 ? (
                  <Text color="gray.500" textAlign="center" py={4}>
                    追加可能なカラムがありません。
                    {columnsData?.columns?.length ===
                      existingFieldCodes.length &&
                      "すべてのカラムが既に追加されています。"}
                  </Text>
                ) : (
                  <VStack gap={4} align="stretch">
                    <Text fontSize="sm" color="gray.600">
                      テーブル「{app.source_table_name}
                      」のカラムから追加するものを選択してください
                    </Text>
                    <Flex justify="space-between" align="center">
                      <Text fontWeight="medium">
                        {selectedColumns.size} / {availableColumns.length}{" "}
                        件選択中
                      </Text>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={handleSelectAll}
                      >
                        {selectedColumns.size === availableColumns.length
                          ? "全て解除"
                          : "全て選択"}
                      </Button>
                    </Flex>
                    <Box
                      maxH="400px"
                      overflowY="auto"
                      borderWidth="1px"
                      borderRadius="md"
                    >
                      <Table.Root size="sm">
                        <Table.Header
                          position="sticky"
                          top={0}
                          bg="white"
                          zIndex={1}
                        >
                          <Table.Row>
                            <Table.ColumnHeader width="40px" />
                            <Table.ColumnHeader>カラム名</Table.ColumnHeader>
                            <Table.ColumnHeader>データ型</Table.ColumnHeader>
                            <Table.ColumnHeader>
                              フィールド型
                            </Table.ColumnHeader>
                            <Table.ColumnHeader>NULL可</Table.ColumnHeader>
                          </Table.Row>
                        </Table.Header>
                        <Table.Body>
                          {availableColumns.map((column) => (
                            <Table.Row
                              key={column.name}
                              cursor="pointer"
                              _hover={{ bg: "gray.50" }}
                              onClick={() => handleColumnToggle(column.name)}
                            >
                              <Table.Cell>
                                <Switch.Root
                                  checked={selectedColumns.has(column.name)}
                                  onCheckedChange={() =>
                                    handleColumnToggle(column.name)
                                  }
                                  onClick={(e) => e.stopPropagation()}
                                  size="sm"
                                >
                                  <Switch.HiddenInput />
                                  <Switch.Control>
                                    <Switch.Thumb />
                                  </Switch.Control>
                                </Switch.Root>
                              </Table.Cell>
                              <Table.Cell fontWeight="medium">
                                {column.name}
                              </Table.Cell>
                              <Table.Cell>
                                <Tag.Root size="sm" colorPalette="gray">
                                  {column.data_type}
                                </Tag.Root>
                              </Table.Cell>
                              <Table.Cell>
                                <Tag.Root size="sm" colorPalette="blue">
                                  {FIELD_TYPE_LABELS[
                                    mapDataTypeToFieldType(
                                      column.data_type
                                    ) as FieldType
                                  ] || mapDataTypeToFieldType(column.data_type)}
                                </Tag.Root>
                              </Table.Cell>
                              <Table.Cell>
                                {column.is_nullable ? (
                                  <Tag.Root size="sm" colorPalette="green">
                                    可
                                  </Tag.Root>
                                ) : (
                                  <Tag.Root size="sm" colorPalette="red">
                                    不可
                                  </Tag.Root>
                                )}
                              </Table.Cell>
                            </Table.Row>
                          ))}
                        </Table.Body>
                      </Table.Root>
                    </Box>
                  </VStack>
                )}
              </Dialog.Body>
              <Dialog.Footer>
                <Button variant="ghost" mr={3} onClick={handleClose}>
                  キャンセル
                </Button>
                <Button
                  colorPalette="brand"
                  onClick={handleExternalSubmit}
                  loading={bulkCreateMutation.isPending}
                  disabled={selectedColumns.size === 0}
                >
                  {selectedColumns.size > 0
                    ? `${selectedColumns.size}件追加`
                    : "追加"}
                </Button>
              </Dialog.Footer>
            </Dialog.Content>
          </Dialog.Positioner>
        </Portal>
      </Dialog.Root>
    );
  }

  // 新規データソースの場合（従来のフォーム）
  return (
    <Dialog.Root open={isOpen} onOpenChange={(e) => !e.open && handleClose()}>
      <Portal>
        <Dialog.Backdrop />
        <Dialog.Positioner>
          <Dialog.Content>
            <form onSubmit={handleSubmit}>
              <Dialog.Header>
                <Dialog.Title>フィールド追加</Dialog.Title>
              </Dialog.Header>
              <Dialog.CloseTrigger asChild>
                <CloseButton size="sm" />
              </Dialog.CloseTrigger>
              <Dialog.Body>
                <VStack gap={4}>
                  <Field.Root invalid={!!errors.field_code}>
                    <Field.Label>フィールドコード</Field.Label>
                    <Input
                      value={formData.field_code}
                      onChange={(e) =>
                        setFormData({ ...formData, field_code: e.target.value })
                      }
                      placeholder="例: customer_name"
                    />
                    <Field.ErrorText>{errors.field_code}</Field.ErrorText>
                  </Field.Root>

                  <Field.Root invalid={!!errors.field_name}>
                    <Field.Label>フィールド名</Field.Label>
                    <Input
                      value={formData.field_name}
                      onChange={(e) =>
                        setFormData({ ...formData, field_name: e.target.value })
                      }
                      placeholder="例: 顧客名"
                    />
                    <Field.ErrorText>{errors.field_name}</Field.ErrorText>
                  </Field.Root>

                  <Field.Root>
                    <Field.Label>フィールドタイプ</Field.Label>
                    <NativeSelect.Root>
                      <NativeSelect.Field
                        value={formData.field_type}
                        onChange={(e) =>
                          setFormData({
                            ...formData,
                            field_type: e.target.value as FieldType,
                          })
                        }
                      >
                        {Object.entries(FIELD_TYPE_LABELS).map(
                          ([value, label]) => (
                            <option key={value} value={value}>
                              {label}
                            </option>
                          )
                        )}
                      </NativeSelect.Field>
                      <NativeSelect.Indicator />
                    </NativeSelect.Root>
                  </Field.Root>

                  {needsChoices && (
                    <Field.Root invalid={!!errors.choices}>
                      <Field.Label>選択肢（1行に1つ）</Field.Label>
                      <Textarea
                        value={choicesText}
                        onChange={(e) => setChoicesText(e.target.value)}
                        placeholder="選択肢1&#10;選択肢2&#10;選択肢3"
                        minH="100px"
                      />
                      <Field.ErrorText>{errors.choices}</Field.ErrorText>
                    </Field.Root>
                  )}

                  <Field.Root display="flex" alignItems="center">
                    <Field.Label mb="0">必須項目</Field.Label>
                    <Switch.Root
                      checked={formData.required}
                      onCheckedChange={(e) =>
                        setFormData({ ...formData, required: e.checked })
                      }
                    >
                      <Switch.HiddenInput />
                      <Switch.Control>
                        <Switch.Thumb />
                      </Switch.Control>
                    </Switch.Root>
                  </Field.Root>
                </VStack>
              </Dialog.Body>
              <Dialog.Footer>
                <Button variant="ghost" mr={3} onClick={handleClose}>
                  キャンセル
                </Button>
                <Button
                  type="submit"
                  colorPalette="brand"
                  loading={createMutation.isPending}
                >
                  追加
                </Button>
              </Dialog.Footer>
            </form>
          </Dialog.Content>
        </Dialog.Positioner>
      </Portal>
    </Dialog.Root>
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
  field: AppField;
}) {
  const { fields } = useApiClient();
  const queryClient = useQueryClient();
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
      toaster.create({
        title: "フィールドを更新しました",
        type: "success",
        duration: 3000,
      });
      onClose();
    },
    onError: () => {
      toaster.create({
        title: "フィールドの更新に失敗しました",
        type: "error",
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
    <Dialog.Root open={isOpen} onOpenChange={(e) => !e.open && onClose()}>
      <Portal>
        <Dialog.Backdrop />
        <Dialog.Positioner>
          <Dialog.Content>
            <form onSubmit={handleSubmit}>
              <Dialog.Header>
                <Dialog.Title>フィールド編集</Dialog.Title>
              </Dialog.Header>
              <Dialog.CloseTrigger asChild>
                <CloseButton size="sm" />
              </Dialog.CloseTrigger>
              <Dialog.Body>
                <VStack gap={4}>
                  <Field.Root>
                    <Field.Label>フィールドコード</Field.Label>
                    <Input value={field.field_code} readOnly bg="gray.50" />
                  </Field.Root>

                  <Field.Root>
                    <Field.Label>フィールドタイプ</Field.Label>
                    <Input
                      value={
                        FIELD_TYPE_LABELS[field.field_type] || field.field_type
                      }
                      readOnly
                      bg="gray.50"
                    />
                  </Field.Root>

                  <Field.Root invalid={!!error}>
                    <Field.Label>フィールド名</Field.Label>
                    <Input
                      value={formData.field_name}
                      onChange={(e) =>
                        setFormData({ ...formData, field_name: e.target.value })
                      }
                    />
                    <Field.ErrorText>{error}</Field.ErrorText>
                  </Field.Root>

                  {needsChoices && (
                    <Field.Root>
                      <Field.Label>選択肢（1行に1つ）</Field.Label>
                      <Textarea
                        value={choicesText}
                        onChange={(e) => setChoicesText(e.target.value)}
                        minH="100px"
                      />
                    </Field.Root>
                  )}

                  <Field.Root display="flex" alignItems="center">
                    <Field.Label mb="0">必須項目</Field.Label>
                    <Switch.Root
                      checked={formData.required}
                      onCheckedChange={(e) =>
                        setFormData({ ...formData, required: e.checked })
                      }
                    >
                      <Switch.HiddenInput />
                      <Switch.Control>
                        <Switch.Thumb />
                      </Switch.Control>
                    </Switch.Root>
                  </Field.Root>
                </VStack>
              </Dialog.Body>
              <Dialog.Footer>
                <Button variant="ghost" mr={3} onClick={onClose}>
                  キャンセル
                </Button>
                <Button
                  type="submit"
                  colorPalette="brand"
                  loading={updateMutation.isPending}
                >
                  保存
                </Button>
              </Dialog.Footer>
            </form>
          </Dialog.Content>
        </Dialog.Positioner>
      </Portal>
    </Dialog.Root>
  );
}
