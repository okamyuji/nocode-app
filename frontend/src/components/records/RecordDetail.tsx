import { FieldValue } from "@/components/fields";
import { Field, RecordItem } from "@/types";
import { formatDateTime } from "@/utils";
import { Badge, Box, Divider, HStack, Text, VStack } from "@chakra-ui/react";

interface RecordDetailProps {
  record: RecordItem;
  fields: Field[];
}

export function RecordDetail({ record, fields }: RecordDetailProps) {
  return (
    <VStack spacing={4} align="stretch">
      <HStack justify="space-between" pb={2}>
        <Text fontWeight="bold" color="gray.600">
          レコードID: {record.id}
        </Text>
        <Badge colorScheme="gray" fontSize="xs">
          更新日: {formatDateTime(record.updated_at)}
        </Badge>
      </HStack>

      <Divider />

      {fields.map((field) => (
        <Box key={field.id} py={2}>
          <Text fontSize="sm" fontWeight="bold" color="gray.500" mb={1}>
            {field.field_name}
            {field.required && (
              <Badge colorScheme="red" ml={1} fontSize="xs">
                必須
              </Badge>
            )}
          </Text>
          <Box fontSize="md">
            <FieldValue field={field} value={record.data[field.field_code]} />
          </Box>
        </Box>
      ))}

      <Divider />

      <HStack justify="space-between" fontSize="sm" color="gray.500">
        <Text>作成者ID: {record.created_by}</Text>
        <Text>作成日: {formatDateTime(record.created_at)}</Text>
      </HStack>
    </VStack>
  );
}
