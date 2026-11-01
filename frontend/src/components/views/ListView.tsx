import { FieldValue } from "@/components/fields";
import { Field, RecordItem } from "@/types";
import { formatShortDate } from "@/utils";
import { DeleteIcon } from "@chakra-ui/icons/Delete";
import { EditIcon } from "@chakra-ui/icons/Edit";
import { ViewIcon } from "@chakra-ui/icons/View";
import {
  Badge,
  Box,
  Card,
  Grid,
  GridItem,
  HStack,
  IconButton,
  SimpleGrid,
  Text,
  VStack,
} from "@chakra-ui/react";
import { Tooltip } from "@/components/ui/tooltip";

interface ListViewProps {
  records: RecordItem[];
  fields: Field[];
  onView?: (record: RecordItem) => void;
  onEdit?: (record: RecordItem) => void;
  onDelete?: (record: RecordItem) => void;
  isAdmin?: boolean;
}

export function ListView({
  records,
  fields,
  onView,
  onEdit,
  onDelete,
  isAdmin = false,
}: ListViewProps) {
  // Show only first few fields in card view
  const displayFields = fields.slice(0, 4);
  const primaryField = fields[0];

  if (records.length === 0) {
    return (
      <Box p={8} textAlign="center">
        <Text color="gray.500">レコードがありません</Text>
      </Box>
    );
  }

  return (
    <SimpleGrid columns={{ base: 1, md: 2, lg: 3, xl: 4 }} gap={3} p={4}>
      {records.map((record) => (
        <Card.Root
          key={record.id}
          variant="outline"
          size="sm"
          _hover={{ shadow: "md", borderColor: "brand.300" }}
          transition="all 0.2s"
          cursor="pointer"
          onClick={() => onView?.(record)}
        >
          <Card.Body py={3} px={4}>
            <VStack align="stretch" gap={2}>
              {/* Header */}
              <HStack justify="space-between" align="start">
                <VStack align="start" gap={0} flex={1} minW={0}>
                  <Badge colorPalette="gray" fontSize="2xs">
                    ID: {record.id}
                  </Badge>
                  {primaryField && (
                    <Text fontWeight="bold" fontSize="md" lineClamp={1}>
                      {String(record.data[primaryField.field_code] || "-")}
                    </Text>
                  )}
                </VStack>

                <HStack gap={0} onClick={(e) => e.stopPropagation()}>
                  <Tooltip content="詳細">
                    <IconButton
                      aria-label="詳細"
                      size="xs"
                      variant="ghost"
                      onClick={() => onView?.(record)}
                    >
                      <ViewIcon />
                    </IconButton>
                  </Tooltip>
                  {isAdmin && (
                    <>
                      <Tooltip content="編集">
                        <IconButton
                          aria-label="編集"
                          size="xs"
                          variant="ghost"
                          onClick={() => onEdit?.(record)}
                        >
                          <EditIcon />
                        </IconButton>
                      </Tooltip>
                      <Tooltip content="削除">
                        <IconButton
                          aria-label="削除"
                          size="xs"
                          variant="ghost"
                          colorPalette="red"
                          onClick={() => onDelete?.(record)}
                        >
                          <DeleteIcon />
                        </IconButton>
                      </Tooltip>
                    </>
                  )}
                </HStack>
              </HStack>

              {/* Fields */}
              <Grid templateColumns="auto 1fr" gap={1} fontSize="xs">
                {displayFields.slice(1).map((field) => (
                  <GridItem key={field.id} display="contents">
                    <Text color="gray.500" fontWeight="medium">
                      {field.field_name}:
                    </Text>
                    <Box truncate>
                      <FieldValue
                        field={field}
                        value={record.data[field.field_code]}
                      />
                    </Box>
                  </GridItem>
                ))}
              </Grid>

              {/* Footer */}
              <HStack justify="space-between" fontSize="2xs" color="gray.400">
                <Text>作成: {formatShortDate(record.created_at)}</Text>
                <Text>更新: {formatShortDate(record.updated_at)}</Text>
              </HStack>
            </VStack>
          </Card.Body>
        </Card.Root>
      ))}
    </SimpleGrid>
  );
}
