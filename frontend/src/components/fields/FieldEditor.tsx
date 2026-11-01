import { CreateFieldRequest, FIELD_TYPE_LABELS, FieldType } from "@/types";
import { DeleteIcon } from "@chakra-ui/icons/Delete";
import { DragHandleIcon } from "@chakra-ui/icons/DragHandle";
import { WarningIcon } from "@chakra-ui/icons/Warning";
import {
  Box,
  Button,
  Checkbox,
  Field,
  HStack,
  IconButton,
  Input,
  NativeSelect,
  Tag,
  Text,
  VStack,
  Wrap,
  WrapItem,
} from "@chakra-ui/react";
import { Tooltip } from "@/components/ui/tooltip";
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
      // 不正な文字を除去
      const sanitized = value.replace(/[^a-zA-Z0-9_]/g, "");
      // 空、または英字で始まる場合のみ更新（数字やアンダースコアで始まる入力を防止）
      if (sanitized === "" || /^[a-zA-Z]/.test(sanitized)) {
        onUpdate({ field_code: sanitized });
      }
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
      borderWidth="1px"
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
            content="このフィールドに入力エラーがあります"
            positioning={{ placement: "top" }}
            showArrow
          >
            <Box position="absolute" top={2} right={12} color="red.500">
              <WarningIcon />
            </Box>
          </Tooltip>
        )}

      <HStack mb={3} justify="space-between">
        <HStack gap={2}>
          <IconButton
            {...attributes}
            {...listeners}
            aria-label="ドラッグ"
            size="sm"
            variant="ghost"
            cursor="grab"
          >
            <DragHandleIcon />
          </IconButton>
          <Text fontWeight="bold" color="gray.600">
            フィールド {index + 1}
          </Text>
          <Tag.Root size="sm" colorPalette="brand">
            <Tag.Label>{FIELD_TYPE_LABELS[field.field_type]}</Tag.Label>
          </Tag.Root>
        </HStack>
        <IconButton
          aria-label="削除"
          size="sm"
          variant="ghost"
          colorPalette="red"
          onClick={onDelete}
        >
          <DeleteIcon />
        </IconButton>
      </HStack>

      <VStack gap={3} align="stretch">
        {/* Row 1: Field Code, Field Name */}
        <HStack gap={4} align="flex-start">
          <Field.Root
            required
            flex={1}
            invalid={touched.field_code && !!errors.field_code}
          >
            <Field.Label fontSize="sm">
              フィールドコード
              <Field.RequiredIndicator />
            </Field.Label>
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
              <Field.ErrorText fontSize="xs">
                {errors.field_code}
              </Field.ErrorText>
            ) : (
              <Field.HelperText fontSize="xs" color="gray.500">
                英字で始まり、英数字と_のみ
              </Field.HelperText>
            )}
          </Field.Root>

          <Field.Root
            required
            flex={1}
            invalid={touched.field_name && !!errors.field_name}
          >
            <Field.Label fontSize="sm">
              表示名
              <Field.RequiredIndicator />
            </Field.Label>
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
              <Field.ErrorText fontSize="xs">
                {errors.field_name}
              </Field.ErrorText>
            )}
          </Field.Root>
        </HStack>

        {/* Row 2: Field Type, Required */}
        <HStack gap={4}>
          <Field.Root flex={1}>
            <Field.Label fontSize="sm">フィールドタイプ</Field.Label>
            <NativeSelect.Root size="sm">
              <NativeSelect.Field
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
              </NativeSelect.Field>
              <NativeSelect.Indicator />
            </NativeSelect.Root>
          </Field.Root>

          <Field.Root flex={1}>
            <Field.Label fontSize="sm">&nbsp;</Field.Label>
            <Checkbox.Root
              checked={field.required}
              onCheckedChange={(e) => onUpdate({ required: !!e.checked })}
            >
              <Checkbox.HiddenInput />
              <Checkbox.Control />
              <Checkbox.Label>必須フィールド</Checkbox.Label>
            </Checkbox.Root>
          </Field.Root>
        </HStack>

        {/* Choices editor for select/multiselect/radio */}
        {showChoicesEditor && (
          <Field.Root invalid={touched.choices && !!errors.choices}>
            <Field.Label fontSize="sm">
              選択肢
              <Text as="span" color="red.500" ml={1}>
                *
              </Text>
            </Field.Label>
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
              <Button size="sm" onClick={handleAddChoice} colorPalette="brand">
                追加
              </Button>
            </HStack>
            <Wrap gap={2}>
              {(field.options?.choices || []).map((choice) => (
                <WrapItem key={choice}>
                  <Tag.Root size="sm" colorPalette="gray">
                    <Tag.Label>{choice}</Tag.Label>
                    <Tag.EndElement>
                      <Tag.CloseTrigger
                        onClick={() => handleRemoveChoice(choice)}
                      />
                    </Tag.EndElement>
                  </Tag.Root>
                </WrapItem>
              ))}
            </Wrap>
            {touched.choices && errors.choices ? (
              <Field.ErrorText fontSize="xs">{errors.choices}</Field.ErrorText>
            ) : (field.options?.choices || []).length === 0 ? (
              <Field.HelperText fontSize="xs" color="gray.500">
                選択肢を追加してください
              </Field.HelperText>
            ) : null}
          </Field.Root>
        )}

        {/* Link type editor */}
        {showLinkTypeEditor && (
          <Field.Root>
            <Field.Label fontSize="sm">リンクタイプ</Field.Label>
            <NativeSelect.Root size="sm">
              <NativeSelect.Field
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
              </NativeSelect.Field>
              <NativeSelect.Indicator />
            </NativeSelect.Root>
          </Field.Root>
        )}
      </VStack>
    </Box>
  );
}

export type { FieldFormData };
