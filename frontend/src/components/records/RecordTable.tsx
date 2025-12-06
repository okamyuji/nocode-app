import { FieldValue } from "@/components/fields";
import { Field, RecordItem } from "@/types";
import {
  Box,
  Checkbox,
  IconButton,
  Menu,
  MenuButton,
  MenuItem,
  MenuList,
  Table,
  Tbody,
  Td,
  Text,
  Th,
  Thead,
  Tr,
} from "@chakra-ui/react";
import { FiEdit2, FiEye, FiMoreVertical, FiTrash2 } from "react-icons/fi";

interface RecordTableProps {
  records: RecordItem[];
  fields: Field[];
  selectedIds: number[];
  onSelectRecord: (id: number) => void;
  onSelectAll: () => void;
  onView: (record: RecordItem) => void;
  onEdit: (record: RecordItem) => void;
  onDelete: (record: RecordItem) => void;
  isAdmin?: boolean;
}

export function RecordTable({
  records,
  fields,
  selectedIds,
  onSelectRecord,
  onSelectAll,
  onView,
  onEdit,
  onDelete,
  isAdmin = false,
}: RecordTableProps) {
  const allSelected =
    records.length > 0 && selectedIds.length === records.length;
  const someSelected =
    selectedIds.length > 0 && selectedIds.length < records.length;

  return (
    <Box overflowX="auto">
      <Table variant="simple" size="sm">
        <Thead bg="gray.50">
          <Tr>
            {isAdmin && (
              <Th w="40px" px={2}>
                <Checkbox
                  isChecked={allSelected}
                  isIndeterminate={someSelected}
                  onChange={onSelectAll}
                />
              </Th>
            )}
            <Th w="80px">ID</Th>
            {fields.map((field) => (
              <Th key={field.id} minW="150px">
                {field.field_name}
              </Th>
            ))}
            <Th w="60px">操作</Th>
          </Tr>
        </Thead>
        <Tbody>
          {records.length === 0 ? (
            <Tr>
              <Td
                colSpan={fields.length + (isAdmin ? 3 : 2)}
                textAlign="center"
                py={8}
              >
                <Text color="gray.500">レコードがありません</Text>
              </Td>
            </Tr>
          ) : (
            records.map((record) => (
              <Tr
                key={record.id}
                _hover={{ bg: "gray.50" }}
                bg={selectedIds.includes(record.id) ? "brand.50" : undefined}
              >
                {isAdmin && (
                  <Td px={2}>
                    <Checkbox
                      isChecked={selectedIds.includes(record.id)}
                      onChange={() => onSelectRecord(record.id)}
                    />
                  </Td>
                )}
                <Td fontWeight="medium" color="gray.600">
                  {record.id}
                </Td>
                {fields.map((field) => (
                  <Td key={field.id}>
                    <FieldValue
                      field={field}
                      value={record.data[field.field_code]}
                    />
                  </Td>
                ))}
                <Td>
                  <Menu>
                    <MenuButton
                      as={IconButton}
                      icon={<FiMoreVertical />}
                      variant="ghost"
                      size="sm"
                    />
                    <MenuList>
                      <MenuItem icon={<FiEye />} onClick={() => onView(record)}>
                        詳細
                      </MenuItem>
                      {isAdmin && (
                        <>
                          <MenuItem
                            icon={<FiEdit2 />}
                            onClick={() => onEdit(record)}
                          >
                            編集
                          </MenuItem>
                          <MenuItem
                            icon={<FiTrash2 />}
                            color="red.500"
                            onClick={() => onDelete(record)}
                          >
                            削除
                          </MenuItem>
                        </>
                      )}
                    </MenuList>
                  </Menu>
                </Td>
              </Tr>
            ))
          )}
        </Tbody>
      </Table>
    </Box>
  );
}
