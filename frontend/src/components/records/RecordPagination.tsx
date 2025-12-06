import { Pagination } from "@/types";
import { ChevronLeftIcon, ChevronRightIcon } from "@chakra-ui/icons";
import { Button, HStack, IconButton, Select, Text } from "@chakra-ui/react";

interface RecordPaginationProps {
  pagination: Pagination;
  onPageChange: (page: number) => void;
  onLimitChange: (limit: number) => void;
}

export function RecordPagination({
  pagination,
  onPageChange,
  onLimitChange,
}: RecordPaginationProps) {
  const { page, limit, total, total_pages } = pagination;

  const startRecord = (page - 1) * limit + 1;
  const endRecord = Math.min(page * limit, total);

  return (
    <HStack justify="space-between" w="100%" py={4}>
      <HStack spacing={2}>
        <Text fontSize="sm" color="gray.600">
          表示件数:
        </Text>
        <Select
          size="sm"
          w="80px"
          value={limit}
          onChange={(e) => onLimitChange(Number(e.target.value))}
        >
          <option value="10">10</option>
          <option value="20">20</option>
          <option value="50">50</option>
          <option value="100">100</option>
        </Select>
      </HStack>

      <HStack spacing={4}>
        <Text fontSize="sm" color="gray.600">
          {total > 0 ? `${startRecord}-${endRecord} / 全${total}件` : "0件"}
        </Text>

        <HStack spacing={1}>
          <IconButton
            icon={<ChevronLeftIcon />}
            aria-label="前のページ"
            size="sm"
            variant="outline"
            isDisabled={page <= 1}
            onClick={() => onPageChange(page - 1)}
          />

          {generatePageNumbers(page, total_pages).map((pageNum, index) =>
            pageNum === "..." ? (
              <Text key={`ellipsis-${index}`} px={2} color="gray.400">
                ...
              </Text>
            ) : (
              <Button
                key={pageNum}
                size="sm"
                variant={page === pageNum ? "solid" : "outline"}
                colorScheme={page === pageNum ? "brand" : "gray"}
                onClick={() => onPageChange(pageNum as number)}
              >
                {pageNum}
              </Button>
            )
          )}

          <IconButton
            icon={<ChevronRightIcon />}
            aria-label="次のページ"
            size="sm"
            variant="outline"
            isDisabled={page >= total_pages}
            onClick={() => onPageChange(page + 1)}
          />
        </HStack>
      </HStack>
    </HStack>
  );
}

function generatePageNumbers(
  currentPage: number,
  totalPages: number
): (number | string)[] {
  const pages: (number | string)[] = [];
  const maxVisible = 5;

  if (totalPages <= maxVisible) {
    for (let i = 1; i <= totalPages; i++) {
      pages.push(i);
    }
    return pages;
  }

  pages.push(1);

  const start = Math.max(2, currentPage - 1);
  const end = Math.min(totalPages - 1, currentPage + 1);

  if (start > 2) {
    pages.push("...");
  }

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  if (end < totalPages - 1) {
    pages.push("...");
  }

  pages.push(totalPages);

  return pages;
}
