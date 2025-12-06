import { Field, FieldType } from "@/types";
import {
  Checkbox,
  CheckboxGroup,
  FormControl,
  FormErrorMessage,
  FormLabel,
  Input,
  NumberInput,
  NumberInputField,
  Radio,
  RadioGroup,
  Select,
  Stack,
  Textarea,
} from "@chakra-ui/react";

interface FieldRendererProps {
  field: Field;
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
            isReadOnly={isReadOnly}
            placeholder={field.field_name}
          />
        );

      case "textarea":
        return (
          <Textarea
            value={(value as string) || ""}
            onChange={(e) => onChange(e.target.value)}
            isReadOnly={isReadOnly}
            placeholder={field.field_name}
            rows={4}
          />
        );

      case "number":
        return (
          <NumberInput
            value={(value as number) || ""}
            onChange={(_, num) => onChange(isNaN(num) ? null : num)}
            isReadOnly={isReadOnly}
          >
            <NumberInputField placeholder={field.field_name} />
          </NumberInput>
        );

      case "date":
        return (
          <Input
            type="date"
            value={(value as string) || ""}
            onChange={(e) => onChange(e.target.value)}
            isReadOnly={isReadOnly}
          />
        );

      case "datetime":
        return (
          <Input
            type="datetime-local"
            value={(value as string) || ""}
            onChange={(e) => onChange(e.target.value)}
            isReadOnly={isReadOnly}
          />
        );

      case "select":
        return (
          <Select
            value={(value as string) || ""}
            onChange={(e) => onChange(e.target.value)}
            isReadOnly={isReadOnly}
            placeholder="選択してください"
          >
            {field.options?.choices?.map((choice) => (
              <option key={choice} value={choice}>
                {choice}
              </option>
            ))}
          </Select>
        );

      case "multiselect":
        return (
          <CheckboxGroup
            value={(value as string[]) || []}
            onChange={(values) => onChange(values)}
          >
            <Stack spacing={2}>
              {field.options?.choices?.map((choice) => (
                <Checkbox key={choice} value={choice} isDisabled={isReadOnly}>
                  {choice}
                </Checkbox>
              ))}
            </Stack>
          </CheckboxGroup>
        );

      case "checkbox":
        return (
          <Checkbox
            isChecked={!!value}
            onChange={(e) => onChange(e.target.checked)}
            isDisabled={isReadOnly}
          >
            {field.field_name}
          </Checkbox>
        );

      case "radio":
        return (
          <RadioGroup
            value={(value as string) || ""}
            onChange={(val) => onChange(val)}
          >
            <Stack spacing={2}>
              {field.options?.choices?.map((choice) => (
                <Radio key={choice} value={choice} isDisabled={isReadOnly}>
                  {choice}
                </Radio>
              ))}
            </Stack>
          </RadioGroup>
        );

      case "link":
        return (
          <Input
            type={field.options?.link_type === "email" ? "email" : "url"}
            value={(value as string) || ""}
            onChange={(e) => onChange(e.target.value)}
            isReadOnly={isReadOnly}
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
            isDisabled={isReadOnly}
          />
        );

      default:
        return (
          <Input
            value={(value as string) || ""}
            onChange={(e) => onChange(e.target.value)}
            isReadOnly={isReadOnly}
          />
        );
    }
  };

  // Checkbox type doesn't need a separate label
  if (field.field_type === "checkbox") {
    return (
      <FormControl isInvalid={!!error} isRequired={field.required}>
        {renderField()}
        <FormErrorMessage>{error}</FormErrorMessage>
      </FormControl>
    );
  }

  return (
    <FormControl isInvalid={!!error} isRequired={field.required}>
      <FormLabel>{field.field_name}</FormLabel>
      {renderField()}
      <FormErrorMessage>{error}</FormErrorMessage>
    </FormControl>
  );
}
