/**
 * レコードフォームコンポーネント
 * レコードの作成・編集用フォームを提供する
 */

import { FieldRenderer } from "@/components/fields";
import { Field, RecordData, RecordItem } from "@/types";
import { Box, Button, HStack, useToast, VStack } from "@chakra-ui/react";
import { useEffect, useState } from "react";

interface RecordFormProps {
  fields: Field[];
  record?: RecordItem;
  onSubmit: (data: RecordData) => Promise<void>;
  onCancel: () => void;
  isSubmitting?: boolean;
}

export function RecordForm({
  fields,
  record,
  onSubmit,
  onCancel,
  isSubmitting = false,
}: RecordFormProps) {
  const [data, setData] = useState<RecordData>({});
  const [errors, setErrors] = useState<Record<string, string>>({});
  const toast = useToast();

  useEffect(() => {
    if (record) {
      setData(record.data || {});
    } else {
      // 空の値で初期化
      const initialData: RecordData = {};
      fields.forEach((field) => {
        initialData[field.field_code] = getDefaultValue(field.field_type);
      });
      setData(initialData);
    }
  }, [record, fields]);

  /** フィールドタイプに応じたデフォルト値を返す */
  const getDefaultValue = (fieldType: string): unknown => {
    switch (fieldType) {
      case "multiselect":
        return [];
      case "checkbox":
        return false;
      case "number":
        return null;
      default:
        return "";
    }
  };

  /** フィールド値変更ハンドラ */
  const handleChange = (fieldCode: string, value: unknown) => {
    setData((prev) => ({ ...prev, [fieldCode]: value }));
    // フィールド変更時にエラーをクリア
    if (errors[fieldCode]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[fieldCode];
        return newErrors;
      });
    }
  };

  /** フォームのバリデーションを実行 */
  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    fields.forEach((field) => {
      if (field.required) {
        const value = data[field.field_code];
        if (value === null || value === undefined || value === "") {
          newErrors[field.field_code] = "必須項目です";
        }
        if (Array.isArray(value) && value.length === 0) {
          newErrors[field.field_code] = "1つ以上選択してください";
        }
      }
    });

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  /** フォーム送信ハンドラ */
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      toast({
        title: "入力エラーがあります",
        status: "error",
        duration: 3000,
      });
      return;
    }

    await onSubmit(data);
  };

  return (
    <Box as="form" onSubmit={handleSubmit}>
      <VStack spacing={4} align="stretch">
        {fields.map((field) => (
          <FieldRenderer
            key={field.id}
            field={field}
            value={data[field.field_code]}
            onChange={(value) => handleChange(field.field_code, value)}
            error={errors[field.field_code]}
          />
        ))}

        <HStack justify="flex-end" spacing={4} pt={4}>
          <Button variant="outline" onClick={onCancel}>
            キャンセル
          </Button>
          <Button
            type="submit"
            colorScheme="brand"
            isLoading={isSubmitting}
            loadingText="保存中..."
          >
            {record ? "更新" : "作成"}
          </Button>
        </HStack>
      </VStack>
    </Box>
  );
}
