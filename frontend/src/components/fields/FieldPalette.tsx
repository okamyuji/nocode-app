import { FIELD_TYPE_LABELS, FieldType } from "@/types";
import { Box, Icon, SimpleGrid, Text, VStack } from "@chakra-ui/react";
import { useDraggable } from "@dnd-kit/core";
import {
  FiCalendar,
  FiCheckSquare,
  FiFile,
  FiHash,
  FiLink,
  FiList,
  FiType,
} from "react-icons/fi";
import { MdOutlineTextFields, MdRadioButtonChecked } from "react-icons/md";

interface FieldPaletteProps {
  onFieldSelect?: (fieldType: FieldType) => void;
}

const fieldTypeIcons: Record<FieldType, React.ComponentType> = {
  text: FiType,
  textarea: MdOutlineTextFields,
  number: FiHash,
  date: FiCalendar,
  datetime: FiCalendar,
  select: FiList,
  multiselect: FiList,
  checkbox: FiCheckSquare,
  radio: MdRadioButtonChecked,
  link: FiLink,
  attachment: FiFile,
};

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
      opacity={isDragging ? 0.5 : 1}
      _hover={{ borderColor: "brand.300", bg: "brand.50" }}
      transition="all 0.2s"
      onClick={() => onSelect?.(fieldType)}
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
