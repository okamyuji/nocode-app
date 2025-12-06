import { Field, FieldType } from "@/types";
import { formatDate, formatDateTime } from "@/utils";
import { Badge, Checkbox, HStack, Link, Text } from "@chakra-ui/react";

interface FieldValueProps {
  field: Field;
  value: unknown;
}

export function FieldValue({ field, value }: FieldValueProps) {
  if (value === null || value === undefined || value === "") {
    return <Text color="gray.400">-</Text>;
  }

  switch (field.field_type as FieldType) {
    case "text":
    case "textarea":
      return <Text>{String(value)}</Text>;

    case "number":
      return <Text>{Number(value).toLocaleString()}</Text>;

    case "date":
      return <Text>{formatDate(value as string)}</Text>;

    case "datetime":
      return <Text>{formatDateTime(value as string)}</Text>;

    case "select":
    case "radio":
      return <Badge colorScheme="brand">{String(value)}</Badge>;

    case "multiselect":
      return (
        <HStack spacing={1} flexWrap="wrap">
          {(value as string[]).map((v) => (
            <Badge key={v} colorScheme="brand">
              {v}
            </Badge>
          ))}
        </HStack>
      );

    case "checkbox":
      return <Checkbox isChecked={!!value} isReadOnly />;

    case "link": {
      const href =
        field.options?.link_type === "email"
          ? `mailto:${value}`
          : String(value);
      return (
        <Link href={href} color="brand.500" isExternal>
          {String(value)}
        </Link>
      );
    }

    case "attachment": {
      const attachment = value as { name: string };
      return <Text>{attachment?.name || "-"}</Text>;
    }

    default:
      return <Text>{String(value)}</Text>;
  }
}
