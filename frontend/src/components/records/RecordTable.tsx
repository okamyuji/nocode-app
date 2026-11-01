import { FieldValue } from "@/components/fields";
import { Field, RecordItem } from "@/types";
import {
  Box,
  Checkbox,
  IconButton,
  Menu,
  Portal,
  Table,
  Text,
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
      <Table.Root variant="line" size="sm">
        <Table.Header bg="gray.50">
          <Table.Row>
            {isAdmin && (
              <Table.ColumnHeader w="40px" px={2}>
                <Checkbox.Root
                  checked={
                    allSelected ? true : someSelected ? "indeterminate" : false
                  }
                  onCheckedChange={onSelectAll}
                >
                  <Checkbox.HiddenInput />
                  <Checkbox.Control />
                </Checkbox.Root>
              </Table.ColumnHeader>
            )}
            <Table.ColumnHeader w="80px">ID</Table.ColumnHeader>
            {fields.map((field) => (
              <Table.ColumnHeader key={field.id} minW="150px">
                {field.field_name}
              </Table.ColumnHeader>
            ))}
            <Table.ColumnHeader w="64px" {...stickyHeaderSx}>
              操作
            </Table.ColumnHeader>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {records.length === 0 ? (
            <Table.Row>
              <Table.Cell
                colSpan={fields.length + (isAdmin ? 3 : 2)}
                textAlign="center"
                py={8}
              >
                <Text color="gray.500">レコードがありません</Text>
              </Table.Cell>
            </Table.Row>
          ) : (
            records.map((record) => {
              const isSelected = selectedIds.includes(record.id);
              const rowBg = isSelected ? "brand.50" : "white";
              return (
                <Table.Row
                  key={record.id}
                  role="group"
                  _hover={{ bg: "gray.50" }}
                  bg={isSelected ? "brand.50" : undefined}
                >
                  {isAdmin && (
                    <Table.Cell px={2}>
                      <Checkbox.Root
                        checked={isSelected}
                        onCheckedChange={() => onSelectRecord(record.id)}
                      >
                        <Checkbox.HiddenInput />
                        <Checkbox.Control />
                      </Checkbox.Root>
                    </Table.Cell>
                  )}
                  <Table.Cell fontWeight="medium" color="gray.600">
                    {record.id}
                  </Table.Cell>
                  {fields.map((field) => (
                    <Table.Cell key={field.id}>
                      <FieldValue
                        field={field}
                        value={record.data[field.field_code]}
                      />
                    </Table.Cell>
                  ))}
                  <Table.Cell
                    position="sticky"
                    right={0}
                    bg={rowBg}
                    zIndex={1}
                    borderLeft="1px solid"
                    borderLeftColor="gray.200"
                    boxShadow="-4px 0 6px -4px rgba(0, 0, 0, 0.05)"
                    _groupHover={{ bg: "gray.50" }}
                  >
                    <Menu.Root
                      positioning={{ placement: "left-start" }}
                      lazyMount
                    >
                      <Menu.Trigger asChild>
                        <IconButton
                          variant="ghost"
                          size="sm"
                          aria-label="行アクション"
                        >
                          <FiMoreVertical />
                        </IconButton>
                      </Menu.Trigger>
                      {/* Portal にレンダリングして、親 Card / main の overflow:hidden に切り取られないようにする */}
                      <Portal>
                        <Menu.Positioner>
                          <Menu.Content zIndex="popover">
                            <Menu.Item
                              value="view"
                              onClick={() => onView(record)}
                            >
                              <FiEye />
                              詳細
                            </Menu.Item>
                            {isAdmin && onEdit && (
                              <Menu.Item
                                value="edit"
                                onClick={() => onEdit(record)}
                              >
                                <FiEdit2 />
                                編集
                              </Menu.Item>
                            )}
                            {isAdmin && onDelete && (
                              <Menu.Item
                                value="delete"
                                color="red.500"
                                onClick={() => onDelete(record)}
                              >
                                <FiTrash2 />
                                削除
                              </Menu.Item>
                            )}
                          </Menu.Content>
                        </Menu.Positioner>
                      </Portal>
                    </Menu.Root>
                  </Table.Cell>
                </Table.Row>
              );
            })
          )}
        </Table.Body>
      </Table.Root>
    </Box>
  );
}
