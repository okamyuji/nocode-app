import { FieldEditor, FieldFormData, FieldPalette } from "@/components/fields";
import { useCreateApp } from "@/hooks";
import { FieldType } from "@/types";
import {
  Box,
  Button,
  Card,
  CardBody,
  CardHeader,
  FormControl,
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
import { useState } from "react";
import { useNavigate } from "react-router-dom";

export function AppFormBuilder() {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [fields, setFields] = useState<FieldFormData[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);

  const navigate = useNavigate();
  const createApp = useCreateApp();
  const toast = useToast();

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const addField = (fieldType: FieldType) => {
    setFields([
      ...fields,
      {
        tempId: crypto.randomUUID(),
        field_code: "",
        field_name: "",
        field_type: fieldType,
        required: false,
        display_order: fields.length + 1,
      },
    ]);
  };

  const removeField = (tempId: string) => {
    setFields(fields.filter((f) => f.tempId !== tempId));
  };

  const updateField = (tempId: string, updates: Partial<FieldFormData>) => {
    setFields(
      fields.map((f) => (f.tempId === tempId ? { ...f, ...updates } : f))
    );
  };

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveId(null);

    // Handle dropping from palette
    if (active.data.current?.type === "new-field") {
      const fieldType = active.data.current.fieldType as FieldType;
      addField(fieldType);
      return;
    }

    // Handle reordering existing fields
    if (over && active.id !== over.id) {
      const oldIndex = fields.findIndex((f) => f.tempId === active.id);
      const newIndex = fields.findIndex((f) => f.tempId === over.id);
      setFields(arrayMove(fields, oldIndex, newIndex));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      toast({
        title: "アプリ名を入力してください",
        status: "error",
        duration: 3000,
      });
      return;
    }

    if (fields.length === 0) {
      toast({
        title: "少なくとも1つのフィールドを追加してください",
        status: "error",
        duration: 3000,
      });
      return;
    }

    const invalidFields = fields.filter(
      (f) => !f.field_code.trim() || !f.field_name.trim()
    );
    if (invalidFields.length > 0) {
      toast({
        title: "すべてのフィールドにコードと名前を入力してください",
        status: "error",
        duration: 3000,
      });
      return;
    }

    // Validate choices for select fields
    const selectFields = fields.filter(
      (f) =>
        (f.field_type === "select" ||
          f.field_type === "multiselect" ||
          f.field_type === "radio") &&
        (!f.options?.choices || f.options.choices.length === 0)
    );
    if (selectFields.length > 0) {
      toast({
        title: "選択フィールドには選択肢を追加してください",
        status: "error",
        duration: 3000,
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
        status: "error",
        duration: 5000,
      });
    }
  };

  const activeField = fields.find((f) => f.tempId === activeId);

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
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
                    <FormControl isRequired>
                      <FormLabel>アプリ名</FormLabel>
                      <Input
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        placeholder="例: 顧客管理"
                      />
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
                      borderColor="gray.300"
                      borderRadius="lg"
                      textAlign="center"
                    >
                      <Text color="gray.500" mb={2}>
                        左のパレットからフィールドをクリックまたはドラッグして追加してください
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

      <DragOverlay>
        {activeField ? (
          <Box
            p={4}
            bg="brand.50"
            border="2px"
            borderColor="brand.500"
            borderRadius="md"
            shadow="lg"
          >
            <Text fontWeight="bold">
              {activeField.field_name || "新しいフィールド"}
            </Text>
          </Box>
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
