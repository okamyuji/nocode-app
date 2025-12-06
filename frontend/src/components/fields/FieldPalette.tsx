import { FIELD_TYPE_LABELS, FieldType } from "@/types";
import { Box, HStack, Icon, SimpleGrid, Text, VStack } from "@chakra-ui/react";
import { useDraggable } from "@dnd-kit/core";
import { fieldTypeIcons } from "./fieldTypeIcons";

interface FieldPaletteProps {
  onFieldSelect?: (fieldType: FieldType) => void;
}

interface DraggableFieldTypeProps {
  fieldType: FieldType;
  label: string;
  onSelect?: (fieldType: FieldType) => void;
}

function DraggableFieldType({
  fieldType,
  label,
  onSelect,
}: DraggableFieldTypeProps) {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: `palette-${fieldType}`,
    data: {
      type: "new-field",
      fieldType,
      label,
    },
  });

  const IconComponent = fieldTypeIcons[fieldType];

  return (
    <Box
      ref={setNodeRef}
      {...listeners}
      {...attributes}
      p={3}
      bg="white"
      border="1px"
      borderColor="gray.200"
      borderRadius="md"
      cursor="grab"
      opacity={isDragging ? 0.3 : 1}
      _hover={{ borderColor: "brand.300", bg: "brand.50" }}
      transition="all 0.2s"
      onClick={() => onSelect?.(fieldType)}
      userSelect="none"
    >
      <VStack spacing={1}>
        <Icon as={IconComponent} boxSize={5} color="brand.500" />
        <Text fontSize="xs" textAlign="center" noOfLines={1}>
          {label}
        </Text>
      </VStack>
    </Box>
  );
}

/**
 * ドラッグ中に表示されるパレットアイテムのオーバーレイ
 */
interface PaletteItemOverlayProps {
  fieldType: FieldType;
  label: string;
}

export function PaletteItemOverlay({
  fieldType,
  label,
}: PaletteItemOverlayProps) {
  const IconComponent = fieldTypeIcons[fieldType];

  return (
    <Box
      p={4}
      bg="brand.50"
      border="2px"
      borderColor="brand.500"
      borderRadius="md"
      shadow="xl"
      minW="120px"
      cursor="grabbing"
    >
      <VStack spacing={2}>
        <Icon as={IconComponent} boxSize={6} color="brand.600" />
        <Text
          fontSize="sm"
          fontWeight="bold"
          color="brand.700"
          textAlign="center"
        >
          {label}
        </Text>
        <Text fontSize="xs" color="gray.500">
          ドロップして追加
        </Text>
      </VStack>
    </Box>
  );
}

/**
 * ドラッグ中に表示される既存フィールドのオーバーレイ
 */
interface FieldItemOverlayProps {
  fieldName: string;
  fieldCode: string;
  fieldType: FieldType;
  fieldIndex: number;
}

export function FieldItemOverlay({
  fieldName,
  fieldCode,
  fieldType,
  fieldIndex,
}: FieldItemOverlayProps) {
  const IconComponent = fieldTypeIcons[fieldType];

  // 表示名の決定: 表示名 > フィールドコード > フィールド #n
  const displayName =
    fieldName.trim() || fieldCode.trim() || `フィールド #${fieldIndex + 1}`;

  return (
    <Box
      p={4}
      bg="white"
      border="2px"
      borderColor="brand.500"
      borderRadius="md"
      shadow="xl"
      minW="200px"
      cursor="grabbing"
    >
      <VStack spacing={1} align="start">
        <HStack spacing={2}>
          <Icon as={IconComponent} boxSize={4} color="brand.500" />
          <Text fontSize="sm" fontWeight="bold">
            {displayName}
          </Text>
        </HStack>
        <Text fontSize="xs" color="gray.500">
          {FIELD_TYPE_LABELS[fieldType]}
        </Text>
      </VStack>
    </Box>
  );
}

export function FieldPalette({ onFieldSelect }: FieldPaletteProps) {
  const fieldTypes = Object.entries(FIELD_TYPE_LABELS) as [FieldType, string][];

  return (
    <VStack align="stretch" spacing={4}>
      <Text fontWeight="bold" color="gray.700">
        フィールドパレット
      </Text>
      <Text fontSize="sm" color="gray.500">
        クリックまたはドラッグで追加
      </Text>
      <SimpleGrid columns={2} spacing={2}>
        {fieldTypes.map(([type, label]) => (
          <DraggableFieldType
            key={type}
            fieldType={type}
            label={label}
            onSelect={onFieldSelect}
          />
        ))}
      </SimpleGrid>
    </VStack>
  );
}
