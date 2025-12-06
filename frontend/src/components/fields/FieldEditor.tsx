import { CreateFieldRequest, FIELD_TYPE_LABELS, FieldType } from "@/types";
import { DeleteIcon, DragHandleIcon, WarningIcon } from "@chakra-ui/icons";
import {
  Box,
  Button,
  Checkbox,
  FormControl,
  FormErrorMessage,
  FormHelperText,
  FormLabel,
  HStack,
  IconButton,
  Input,
  Select,
  Tag,
  TagCloseButton,
  TagLabel,
  Text,
  Tooltip,
  VStack,
  Wrap,
  WrapItem,
} from "@chakra-ui/react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useCallback, useMemo, useState } from "react";

interface FieldFormData extends CreateFieldRequest {
  tempId: string;
  options?: {
    choices?: string[];
    link_type?: "url" | "email";
  };
}

// バリデーションエラーの型
interface FieldValidationErrors {
  field_code?: string;
  field_name?: string;
  choices?: string;
}

// バリデーションルール
const validateFieldCode = (value: string): string | undefined => {
  if (!value.trim()) {
    return "フィールドコードは必須です";
  }
  if (value.length > 64) {
    return "フィールドコードは64文字以内で入力してください";
  }
  if (!/^[a-zA-Z][a-zA-Z0-9_]*$/.test(value)) {
    return "英字で始まり、英数字とアンダースコアのみ使用可能です";
  }
  return undefined;
};

const validateFieldName = (value: string): string | undefined => {
  if (!value.trim()) {
    return "表示名は必須です";
  }
  if (value.length > 100) {
    return "表示名は100文字以内で入力してください";
  }
  return undefined;
};

const validateChoices = (
  fieldType: FieldType,
  choices?: string[]
): string | undefined => {
  if (
    (fieldType === "select" ||
      fieldType === "multiselect" ||
      fieldType === "radio") &&
    (!choices || choices.length === 0)
  ) {
    return "選択肢を1つ以上追加してください";
  }
  return undefined;
};

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
  const [touched, setTouched] = useState<{
    field_code: boolean;
    field_name: boolean;
    choices: boolean;
  }>({
    field_code: false,
    field_name: false,
    choices: false,
  });

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

  // バリデーションエラーの計算
  const errors: FieldValidationErrors = useMemo(
    () => ({
      field_code: validateFieldCode(field.field_code),
      field_name: validateFieldName(field.field_name),
      choices: validateChoices(field.field_type, field.options?.choices),
    }),
    [
      field.field_code,
      field.field_name,
      field.field_type,
      field.options?.choices,
    ]
  );

  // フィールドにエラーがあるかどうか
  const hasErrors = useMemo(
    () => Object.values(errors).some((e) => e !== undefined),
    [errors]
  );

  const handleBlur = useCallback((fieldName: keyof typeof touched) => {
    setTouched((prev) => ({ ...prev, [fieldName]: true }));
  }, []);

  const handleFieldCodeChange = useCallback(
    (value: string) => {
      // 英字で始まる場合のみ、または空の場合は許可
      const sanitized = value.replace(/[^a-zA-Z0-9_]/g, "");
      onUpdate({ field_code: sanitized });
    },
    [onUpdate]
  );

  const handleAddChoice = useCallback(() => {
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
    setTouched((prev) => ({ ...prev, choices: true }));
  }, [newChoice, field.options, onUpdate]);

  const handleRemoveChoice = useCallback(
    (choice: string) => {
      const currentChoices = field.options?.choices || [];
      onUpdate({
        options: {
          ...field.options,
          choices: currentChoices.filter((c) => c !== choice),
        },
      });
      setTouched((prev) => ({ ...prev, choices: true }));
    },
    [field.options, onUpdate]
  );

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
      borderColor={
        hasErrors &&
        (touched.field_code || touched.field_name || touched.choices)
          ? "red.300"
          : isDragging
            ? "brand.400"
            : "gray.200"
      }
      borderRadius="md"
      p={4}
      shadow={isDragging ? "lg" : "sm"}
      position="relative"
    >
      {/* エラーインジケーター */}
      {hasErrors &&
        (touched.field_code || touched.field_name || touched.choices) && (
          <Tooltip
            label="このフィールドに入力エラーがあります"
            placement="top"
            hasArrow
          >
            <Box position="absolute" top={2} right={12} color="red.500">
              <WarningIcon />
            </Box>
          </Tooltip>
        )}

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
        <HStack spacing={4} align="flex-start">
          <FormControl
            isRequired
            flex={1}
            isInvalid={touched.field_code && !!errors.field_code}
          >
            <FormLabel fontSize="sm">フィールドコード</FormLabel>
            <Input
              size="sm"
              value={field.field_code}
              onChange={(e) => handleFieldCodeChange(e.target.value)}
              onBlur={() => handleBlur("field_code")}
              placeholder="customer_name"
              borderColor={
                touched.field_code && errors.field_code ? "red.300" : undefined
              }
              _focus={{
                borderColor:
                  touched.field_code && errors.field_code
                    ? "red.500"
                    : "brand.500",
                boxShadow:
                  touched.field_code && errors.field_code
                    ? "0 0 0 1px var(--chakra-colors-red-500)"
                    : "0 0 0 1px var(--chakra-colors-brand-500)",
              }}
            />
            {touched.field_code && errors.field_code ? (
              <FormErrorMessage fontSize="xs">
                {errors.field_code}
              </FormErrorMessage>
            ) : (
              <FormHelperText fontSize="xs" color="gray.500">
                英字で始まり、英数字と_のみ
              </FormHelperText>
            )}
          </FormControl>

          <FormControl
            isRequired
            flex={1}
            isInvalid={touched.field_name && !!errors.field_name}
          >
            <FormLabel fontSize="sm">表示名</FormLabel>
            <Input
              size="sm"
              value={field.field_name}
              onChange={(e) => onUpdate({ field_name: e.target.value })}
              onBlur={() => handleBlur("field_name")}
              placeholder="顧客名"
              borderColor={
                touched.field_name && errors.field_name ? "red.300" : undefined
              }
              _focus={{
                borderColor:
                  touched.field_name && errors.field_name
                    ? "red.500"
                    : "brand.500",
                boxShadow:
                  touched.field_name && errors.field_name
                    ? "0 0 0 1px var(--chakra-colors-red-500)"
                    : "0 0 0 1px var(--chakra-colors-brand-500)",
              }}
            />
            {touched.field_name && errors.field_name && (
              <FormErrorMessage fontSize="xs">
                {errors.field_name}
              </FormErrorMessage>
            )}
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
          <FormControl isInvalid={touched.choices && !!errors.choices}>
            <FormLabel fontSize="sm">
              選択肢
              <Text as="span" color="red.500" ml={1}>
                *
              </Text>
            </FormLabel>
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
                onBlur={() => handleBlur("choices")}
              />
              <Button size="sm" onClick={handleAddChoice} colorScheme="brand">
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
            {touched.choices && errors.choices ? (
              <FormErrorMessage fontSize="xs">
                {errors.choices}
              </FormErrorMessage>
            ) : (field.options?.choices || []).length === 0 ? (
              <FormHelperText fontSize="xs" color="gray.500">
                選択肢を追加してください
              </FormHelperText>
            ) : null}
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
