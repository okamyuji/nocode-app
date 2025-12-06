import {
  FieldEditor,
  FieldFormData,
  FieldItemOverlay,
  FieldPalette,
  PaletteItemOverlay,
} from "@/components/fields";
import { useCreateApp } from "@/hooks";
import { FIELD_TYPE_LABELS, FieldType } from "@/types";
import {
  Box,
  Button,
  Card,
  CardBody,
  CardHeader,
  FormControl,
  FormErrorMessage,
  FormLabel,
  Grid,
  GridItem,
  Heading,
  HStack,
  Input,
  Text,
  Textarea,
  useToast,
  VStack,
} from "@chakra-ui/react";
import {
  closestCenter,
  DndContext,
  DragEndEvent,
  DragOverlay,
  DragStartEvent,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { useCallback, useState } from "react";
import { useNavigate } from "react-router-dom";

// ドラッグ中のアイテムの状態を表す型
interface DragState {
  type: "palette" | "field";
  fieldType: FieldType;
  label: string;
  fieldName?: string;
  fieldCode?: string;
  fieldIndex?: number;
}

export function AppFormBuilder() {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [fields, setFields] = useState<FieldFormData[]>([]);
  const [dragState, setDragState] = useState<DragState | null>(null);
  const [touched, setTouched] = useState({ name: false });

  const navigate = useNavigate();
  const createApp = useCreateApp();
  const toast = useToast();

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8, // ドラッグ開始までの移動距離（誤タップ防止）
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const addField = useCallback((fieldType: FieldType) => {
    setFields((prev) => [
      ...prev,
      {
        tempId: crypto.randomUUID(),
        field_code: "",
        field_name: "",
        field_type: fieldType,
        required: false,
        display_order: prev.length + 1,
      },
    ]);
  }, []);

  const removeField = useCallback((tempId: string) => {
    setFields((prev) => prev.filter((f) => f.tempId !== tempId));
  }, []);

  const updateField = useCallback(
    (tempId: string, updates: Partial<FieldFormData>) => {
      setFields((prev) =>
        prev.map((f) => (f.tempId === tempId ? { ...f, ...updates } : f))
      );
    },
    []
  );

  const handleDragStart = useCallback(
    (event: DragStartEvent) => {
      const { active } = event;

      // パレットからのドラッグの場合
      if (active.data.current?.type === "new-field") {
        const fieldType = active.data.current.fieldType as FieldType;
        const label = active.data.current.label as string;
        setDragState({
          type: "palette",
          fieldType,
          label: label || FIELD_TYPE_LABELS[fieldType],
        });
        return;
      }

      // 既存フィールドのドラッグの場合
      const fieldIndex = fields.findIndex((f) => f.tempId === active.id);
      const field = fieldIndex >= 0 ? fields[fieldIndex] : null;
      if (field) {
        setDragState({
          type: "field",
          fieldType: field.field_type,
          label: FIELD_TYPE_LABELS[field.field_type],
          fieldName: field.field_name,
          fieldCode: field.field_code,
          fieldIndex,
        });
      }
    },
    [fields]
  );

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event;
      setDragState(null);

      // パレットからのドロップ
      if (active.data.current?.type === "new-field") {
        const fieldType = active.data.current.fieldType as FieldType;
        addField(fieldType);
        return;
      }

      // 既存フィールドの並び替え
      if (over && active.id !== over.id) {
        setFields((prev) => {
          const oldIndex = prev.findIndex((f) => f.tempId === active.id);
          const newIndex = prev.findIndex((f) => f.tempId === over.id);
          return arrayMove(prev, oldIndex, newIndex);
        });
      }
    },
    [addField]
  );

  const handleDragCancel = useCallback(() => {
    setDragState(null);
  }, []);

  const validateForm = useCallback(() => {
    const errors: string[] = [];

    if (!name.trim()) {
      errors.push("アプリ名を入力してください");
    }

    if (fields.length === 0) {
      errors.push("少なくとも1つのフィールドを追加してください");
    }

    const invalidFields = fields.filter(
      (f) => !f.field_code.trim() || !f.field_name.trim()
    );
    if (invalidFields.length > 0) {
      errors.push("すべてのフィールドにコードと名前を入力してください");
    }

    // フィールドコードのバリデーション
    const invalidCodeFields = fields.filter(
      (f) =>
        f.field_code.trim() && !/^[a-zA-Z][a-zA-Z0-9_]*$/.test(f.field_code)
    );
    if (invalidCodeFields.length > 0) {
      errors.push(
        "フィールドコードは英字で始まり、英数字とアンダースコアのみ使用可能です"
      );
    }

    // 選択フィールドの選択肢チェック
    const selectFields = fields.filter(
      (f) =>
        (f.field_type === "select" ||
          f.field_type === "multiselect" ||
          f.field_type === "radio") &&
        (!f.options?.choices || f.options.choices.length === 0)
    );
    if (selectFields.length > 0) {
      errors.push("選択フィールドには選択肢を追加してください");
    }

    return errors;
  }, [name, fields]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setTouched({ name: true });

    const errors = validateForm();
    if (errors.length > 0) {
      toast({
        title: "入力エラー",
        description: errors[0],
        status: "error",
        duration: 5000,
        isClosable: true,
      });
      return;
    }

    try {
      const app = await createApp.mutateAsync({
        name,
        description,
        fields: fields.map(({ tempId: _tempId, ...field }, index) => ({
          ...field,
          display_order: index + 1,
        })),
      });

      toast({
        title: "アプリを作成しました",
        status: "success",
        duration: 3000,
      });

      navigate(`/apps/${app.id}/records`);
    } catch {
      toast({
        title: "アプリの作成に失敗しました",
        description: "入力内容を確認してもう一度お試しください",
        status: "error",
        duration: 5000,
        isClosable: true,
      });
    }
  };

  const nameError =
    touched.name && !name.trim() ? "アプリ名は必須です" : undefined;

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
    >
      <Box as="form" onSubmit={handleSubmit}>
        <Grid templateColumns={{ base: "1fr", lg: "250px 1fr" }} gap={6}>
          {/* Field Palette */}
          <GridItem>
            <Box position="sticky" top={4} bg="gray.50" p={4} borderRadius="lg">
              <FieldPalette onFieldSelect={addField} />
            </Box>
          </GridItem>

          {/* Main Content */}
          <GridItem>
            <VStack spacing={6} align="stretch">
              <Card>
                <CardHeader>
                  <Heading size="md">基本情報</Heading>
                </CardHeader>
                <CardBody>
                  <VStack spacing={4}>
                    <FormControl isRequired isInvalid={!!nameError}>
                      <FormLabel>アプリ名</FormLabel>
                      <Input
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        onBlur={() =>
                          setTouched((prev) => ({ ...prev, name: true }))
                        }
                        placeholder="例: 顧客管理"
                        borderColor={nameError ? "red.300" : undefined}
                      />
                      {nameError && (
                        <FormErrorMessage>{nameError}</FormErrorMessage>
                      )}
                    </FormControl>

                    <FormControl>
                      <FormLabel>説明</FormLabel>
                      <Textarea
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        placeholder="アプリの説明を入力"
                        rows={3}
                      />
                    </FormControl>
                  </VStack>
                </CardBody>
              </Card>

              <Card>
                <CardHeader>
                  <Heading size="md">フィールド設定</Heading>
                </CardHeader>
                <CardBody>
                  {fields.length === 0 ? (
                    <Box
                      p={8}
                      border="2px dashed"
                      borderColor={dragState ? "brand.400" : "gray.300"}
                      bg={dragState ? "brand.50" : undefined}
                      borderRadius="lg"
                      textAlign="center"
                      transition="all 0.2s"
                    >
                      <Text color={dragState ? "brand.600" : "gray.500"} mb={2}>
                        {dragState
                          ? "ここにドロップしてフィールドを追加"
                          : "左のパレットからフィールドをクリックまたはドラッグして追加してください"}
                      </Text>
                      <Text fontSize="sm" color="gray.400">
                        フィールドはドラッグで並べ替え可能です
                      </Text>
                    </Box>
                  ) : (
                    <SortableContext
                      items={fields.map((f) => f.tempId)}
                      strategy={verticalListSortingStrategy}
                    >
                      <VStack spacing={4} align="stretch">
                        {fields.map((field, index) => (
                          <FieldEditor
                            key={field.tempId}
                            field={field}
                            index={index}
                            onUpdate={(updates) =>
                              updateField(field.tempId, updates)
                            }
                            onDelete={() => removeField(field.tempId)}
                          />
                        ))}
                      </VStack>
                    </SortableContext>
                  )}
                </CardBody>
              </Card>

              <HStack justify="flex-end" spacing={4}>
                <Button variant="outline" onClick={() => navigate("/apps")}>
                  キャンセル
                </Button>
                <Button
                  type="submit"
                  colorScheme="brand"
                  isLoading={createApp.isPending}
                  loadingText="作成中..."
                  isDisabled={fields.length === 0}
                >
                  アプリを作成
                </Button>
              </HStack>
            </VStack>
          </GridItem>
        </Grid>
      </Box>

      {/* ドラッグオーバーレイ */}
      <DragOverlay dropAnimation={null} style={{ zIndex: 9999 }}>
        {dragState ? (
          <Box
            pointerEvents="none"
            style={{
              cursor: "grabbing",
            }}
          >
            {dragState.type === "palette" ? (
              <PaletteItemOverlay
                fieldType={dragState.fieldType}
                label={dragState.label}
              />
            ) : (
              <FieldItemOverlay
                fieldName={dragState.fieldName || ""}
                fieldCode={dragState.fieldCode || ""}
                fieldType={dragState.fieldType}
                fieldIndex={dragState.fieldIndex ?? 0}
              />
            )}
          </Box>
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
