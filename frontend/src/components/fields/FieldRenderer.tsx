import { Field as FieldModel, FieldType } from "@/types";
import {
  Checkbox,
  CheckboxGroup,
  Field,
  Input,
  NativeSelect,
  NumberInput,
  RadioGroup,
  Stack,
  Textarea,
} from "@chakra-ui/react";

interface FieldRendererProps {
  field: FieldModel;
  value: unknown;
  onChange: (value: unknown) => void;
  error?: string;
  isReadOnly?: boolean;
}

export function FieldRenderer({
  field,
  value,
  onChange,
  error,
  isReadOnly = false,
}: FieldRendererProps) {
  const renderField = () => {
    switch (field.field_type as FieldType) {
      case "text":
        return (
          <Input
            value={(value as string) || ""}
            onChange={(e) => onChange(e.target.value)}
            readOnly={isReadOnly}
            placeholder={field.field_name}
          />
        );

      case "textarea":
        return (
          <Textarea
            value={(value as string) || ""}
            onChange={(e) => onChange(e.target.value)}
            readOnly={isReadOnly}
            placeholder={field.field_name}
            rows={4}
          />
        );

      case "number":
        return (
          <NumberInput.Root
            value={String((value as number) || "")}
            onValueChange={(e) =>
              onChange(isNaN(e.valueAsNumber) ? null : e.valueAsNumber)
            }
            readOnly={isReadOnly}
          >
            <NumberInput.Control />
            <NumberInput.Input placeholder={field.field_name} />
          </NumberInput.Root>
        );

      case "date":
        return (
          <Input
            type="date"
            value={(value as string) || ""}
            onChange={(e) => onChange(e.target.value)}
            readOnly={isReadOnly}
          />
        );

      case "datetime":
        return (
          <Input
            type="datetime-local"
            value={(value as string) || ""}
            onChange={(e) => onChange(e.target.value)}
            readOnly={isReadOnly}
          />
        );

      case "select":
        return (
          <NativeSelect.Root disabled={isReadOnly}>
            <NativeSelect.Field
              value={(value as string) || ""}
              onChange={(e) => onChange(e.target.value)}
              placeholder="選択してください"
            >
              {field.options?.choices?.map((choice) => (
                <option key={choice} value={choice}>
                  {choice}
                </option>
              ))}
            </NativeSelect.Field>
            <NativeSelect.Indicator />
          </NativeSelect.Root>
        );

      case "multiselect":
        return (
          <CheckboxGroup
            value={(value as string[]) || []}
            onValueChange={(values) => onChange(values)}
          >
            <Stack gap={2}>
              {field.options?.choices?.map((choice) => (
                <Checkbox.Root
                  key={choice}
                  value={choice}
                  disabled={isReadOnly}
                >
                  <Checkbox.HiddenInput />
                  <Checkbox.Control />
                  <Checkbox.Label>{choice}</Checkbox.Label>
                </Checkbox.Root>
              ))}
            </Stack>
          </CheckboxGroup>
        );

      case "checkbox":
        return (
          <Checkbox.Root
            checked={!!value}
            onCheckedChange={(e) => onChange(!!e.checked)}
            disabled={isReadOnly}
          >
            <Checkbox.HiddenInput />
            <Checkbox.Control />
            <Checkbox.Label>{field.field_name}</Checkbox.Label>
          </Checkbox.Root>
        );

      case "radio":
        return (
          <RadioGroup.Root
            value={(value as string) || ""}
            onValueChange={(e) => onChange(e.value)}
          >
            <Stack gap={2}>
              {field.options?.choices?.map((choice) => (
                <RadioGroup.Item
                  key={choice}
                  value={choice}
                  disabled={isReadOnly}
                >
                  <RadioGroup.ItemHiddenInput />
                  <RadioGroup.ItemIndicator />
                  <RadioGroup.ItemText>{choice}</RadioGroup.ItemText>
                </RadioGroup.Item>
              ))}
            </Stack>
          </RadioGroup.Root>
        );

      case "link":
        return (
          <Input
            type={field.options?.link_type === "email" ? "email" : "url"}
            value={(value as string) || ""}
            onChange={(e) => onChange(e.target.value)}
            readOnly={isReadOnly}
            placeholder={
              field.options?.link_type === "email"
                ? "email@example.com"
                : "https://example.com"
            }
          />
        );

      case "attachment":
        return (
          <Input
            type="file"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) {
                onChange({ name: file.name, size: file.size, type: file.type });
              }
            }}
            disabled={isReadOnly}
          />
        );

      default:
        return (
          <Input
            value={(value as string) || ""}
            onChange={(e) => onChange(e.target.value)}
            readOnly={isReadOnly}
          />
        );
    }
  };

  // Checkbox type doesn't need a separate label
  if (field.field_type === "checkbox") {
    return (
      <Field.Root invalid={!!error} required={field.required}>
        {renderField()}
        <Field.ErrorText>{error}</Field.ErrorText>
      </Field.Root>
    );
  }

  return (
    <Field.Root invalid={!!error} required={field.required}>
      <Field.Label>{field.field_name}</Field.Label>
      {renderField()}
      <Field.ErrorText>{error}</Field.ErrorText>
    </Field.Root>
  );
}
