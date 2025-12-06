import { CreateFieldRequest, FIELD_TYPE_LABELS, FieldType } from "@/types";
import { DeleteIcon, DragHandleIcon } from "@chakra-ui/icons";
import {
  Box,
  Button,
  Checkbox,
  FormControl,
  FormLabel,
  HStack,
  IconButton,
  Input,
  Select,
  Tag,
  TagCloseButton,
  TagLabel,
  Text,
  VStack,
  Wrap,
  WrapItem,
} from "@chakra-ui/react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useState } from "react";

interface FieldFormData extends CreateFieldRequest {
  tempId: string;
  options?: {
    choices?: string[];
    link_type?: "url" | "email";
  };
}

interface FieldEditorProps {
  field: FieldFormData;
  index: number;
  onUpdate: (updates: Partial<FieldFormData>) => void;
  onDelete: () => void;
}

export function FieldEditor({
  field,
  index,
  onUpdate,
  onDelete,
}: FieldEditorProps) {
  const [newChoice, setNewChoice] = useState("");

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: field.tempId });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const handleAddChoice = () => {
    if (!newChoice.trim()) return;
    const currentChoices = field.options?.choices || [];
    if (currentChoices.includes(newChoice.trim())) return;

    onUpdate({
      options: {
        ...field.options,
        choices: [...currentChoices, newChoice.trim()],
      },
    });
    setNewChoice("");
  };

  const handleRemoveChoice = (choice: string) => {
    const currentChoices = field.options?.choices || [];
    onUpdate({
      options: {
        ...field.options,
        choices: currentChoices.filter((c) => c !== choice),
      },
    });
  };

  const showChoicesEditor =
    field.field_type === "select" ||
    field.field_type === "multiselect" ||
    field.field_type === "radio";

  const showLinkTypeEditor = field.field_type === "link";

  return (
    <Box
      ref={setNodeRef}
      style={style}
      bg="white"
      border="1px"
      borderColor={isDragging ? "brand.400" : "gray.200"}
      borderRadius="md"
      p={4}
      shadow={isDragging ? "lg" : "sm"}
    >
      <HStack mb={3} justify="space-between">
        <HStack spacing={2}>
          <IconButton
            {...attributes}
            {...listeners}
            icon={<DragHandleIcon />}
            aria-label="ドラッグ"
            size="sm"
            variant="ghost"
            cursor="grab"
          />
          <Text fontWeight="bold" color="gray.600">
            フィールド {index + 1}
          </Text>
          <Tag size="sm" colorScheme="brand">
            {FIELD_TYPE_LABELS[field.field_type]}
          </Tag>
        </HStack>
        <IconButton
          icon={<DeleteIcon />}
          aria-label="削除"
          size="sm"
          variant="ghost"
          colorScheme="red"
          onClick={onDelete}
        />
      </HStack>

      <VStack spacing={3} align="stretch">
        {/* Row 1: Field Code, Field Name */}
        <HStack spacing={4}>
          <FormControl isRequired flex={1}>
            <FormLabel fontSize="sm">フィールドコード</FormLabel>
            <Input
              size="sm"
              value={field.field_code}
              onChange={(e) =>
                onUpdate({
                  field_code: e.target.value.replace(/[^a-zA-Z0-9_]/g, ""),
                })
              }
              placeholder="customer_name"
            />
          </FormControl>

          <FormControl isRequired flex={1}>
            <FormLabel fontSize="sm">表示名</FormLabel>
            <Input
              size="sm"
              value={field.field_name}
              onChange={(e) => onUpdate({ field_name: e.target.value })}
              placeholder="顧客名"
            />
          </FormControl>
        </HStack>

        {/* Row 2: Field Type, Required */}
        <HStack spacing={4}>
          <FormControl flex={1}>
            <FormLabel fontSize="sm">フィールドタイプ</FormLabel>
            <Select
              size="sm"
              value={field.field_type}
              onChange={(e) =>
                onUpdate({ field_type: e.target.value as FieldType })
              }
            >
              {Object.entries(FIELD_TYPE_LABELS).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </Select>
          </FormControl>

          <FormControl flex={1}>
            <FormLabel fontSize="sm">&nbsp;</FormLabel>
            <Checkbox
              isChecked={field.required}
              onChange={(e) => onUpdate({ required: e.target.checked })}
            >
              必須フィールド
            </Checkbox>
          </FormControl>
        </HStack>

        {/* Choices editor for select/multiselect/radio */}
        {showChoicesEditor && (
          <FormControl>
            <FormLabel fontSize="sm">選択肢</FormLabel>
            <HStack mb={2}>
              <Input
                size="sm"
                value={newChoice}
                onChange={(e) => setNewChoice(e.target.value)}
                placeholder="選択肢を入力"
                onKeyPress={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    handleAddChoice();
                  }
                }}
              />
              <Button size="sm" onClick={handleAddChoice}>
                追加
              </Button>
            </HStack>
            <Wrap spacing={2}>
              {(field.options?.choices || []).map((choice) => (
                <WrapItem key={choice}>
                  <Tag size="sm" colorScheme="gray">
                    <TagLabel>{choice}</TagLabel>
                    <TagCloseButton
                      onClick={() => handleRemoveChoice(choice)}
                    />
                  </Tag>
                </WrapItem>
              ))}
            </Wrap>
            {(field.options?.choices || []).length === 0 && (
              <Text fontSize="xs" color="gray.500">
                選択肢を追加してください
              </Text>
            )}
          </FormControl>
        )}

        {/* Link type editor */}
        {showLinkTypeEditor && (
          <FormControl>
            <FormLabel fontSize="sm">リンクタイプ</FormLabel>
            <Select
              size="sm"
              value={field.options?.link_type || "url"}
              onChange={(e) =>
                onUpdate({
                  options: {
                    ...field.options,
                    link_type: e.target.value as "url" | "email",
                  },
                })
              }
            >
              <option value="url">URL</option>
              <option value="email">メールアドレス</option>
            </Select>
          </FormControl>
        )}
      </VStack>
    </Box>
  );
}

export type { FieldFormData };
