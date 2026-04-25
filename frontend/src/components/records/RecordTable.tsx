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
  Portal,
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
  onEdit?: (record: RecordItem) => void;
  onDelete?: (record: RecordItem) => void;
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

  // 操作列を右端に固定するための共通スタイル。
  // 多フィールド時に横スクロールが発生しても操作メニューが常に見える。
  const stickyHeaderSx = {
    position: "sticky" as const,
    right: 0,
    bg: "gray.50",
    zIndex: 2,
    borderLeft: "1px solid",
    borderLeftColor: "gray.200",
    boxShadow: "-4px 0 6px -4px rgba(0, 0, 0, 0.05)",
  };

  return (
    <Box overflowX="auto" maxW="100%" w="100%">
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
            <Th w="64px" sx={stickyHeaderSx}>
              操作
            </Th>
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
            records.map((record) => {
              const isSelected = selectedIds.includes(record.id);
              const rowBg = isSelected ? "brand.50" : "white";
              return (
                <Tr
                  key={record.id}
                  role="group"
                  _hover={{ bg: "gray.50" }}
                  bg={isSelected ? "brand.50" : undefined}
                >
                  {isAdmin && (
                    <Td px={2}>
                      <Checkbox
                        isChecked={isSelected}
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
                  <Td
                    sx={{
                      position: "sticky",
                      right: 0,
                      bg: rowBg,
                      zIndex: 1,
                      borderLeft: "1px solid",
                      borderLeftColor: "gray.200",
                      boxShadow: "-4px 0 6px -4px rgba(0, 0, 0, 0.05)",
                      _groupHover: { bg: "gray.50" },
                    }}
                  >
                    <Menu placement="left-start" isLazy>
                      <MenuButton
                        as={IconButton}
                        icon={<FiMoreVertical />}
                        variant="ghost"
                        size="sm"
                        aria-label="行アクション"
                      />
                      {/* Portal にレンダリングして、親 Card / main の overflow:hidden に切り取られないようにする */}
                      <Portal>
                        <MenuList zIndex="popover">
                          <MenuItem
                            icon={<FiEye />}
                            onClick={() => onView(record)}
                          >
                            詳細
                          </MenuItem>
                          {isAdmin && onEdit && (
                            <MenuItem
                              icon={<FiEdit2 />}
                              onClick={() => onEdit(record)}
                            >
                              編集
                            </MenuItem>
                          )}
                          {isAdmin && onDelete && (
                            <MenuItem
                              icon={<FiTrash2 />}
                              color="red.500"
                              onClick={() => onDelete(record)}
                            >
                              削除
                            </MenuItem>
                          )}
                        </MenuList>
                      </Portal>
                    </Menu>
                  </Td>
                </Tr>
              );
            })
          )}
        </Tbody>
      </Table>
    </Box>
  );
}
