/**
 * ダッシュボードウィジェットカードコンポーネント
 * 各アプリのデータをテーブル/リスト/グラフ形式で表示
 * 表示形式はアプリ設定で変更、ダッシュボードではDnDによる並び替えのみ
 */

import { ChartRenderer } from "@/components/charts";
import { RecordTable } from "@/components/records";
import { ListView } from "@/components/views";
import { useChartData, useRecords } from "@/hooks";
import type { DashboardWidget, WidgetSize } from "@/types";
import type { ChartDataRequest } from "@/types/chart";
import { getAppIcon } from "@/utils";
import {
  Badge,
  Box,
  Card,
  CardBody,
  CardHeader,
  Heading,
  HStack,
  Icon,
  IconButton,
  Spinner,
  Text,
  VStack,
} from "@chakra-ui/react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useMemo } from "react";
import { FiMove } from "react-icons/fi";

interface DashboardWidgetCardProps {
  widget: DashboardWidget;
  isDragging?: boolean;
}

/**
 * ウィジェットサイズに応じた高さを取得
 */
function getWidgetHeight(size: WidgetSize): string {
  switch (size) {
    case "small":
      return "250px";
    case "medium":
      return "350px";
    case "large":
      return "500px";
    default:
      return "350px";
  }
}

/**
 * ウィジェットサイズに応じたレコード数を取得
 */
function getRecordLimit(size: WidgetSize): number {
  switch (size) {
    case "small":
      return 5;
    case "medium":
      return 10;
    case "large":
      return 20;
    default:
      return 10;
  }
}

export function DashboardWidgetCard({
  widget,
  isDragging = false,
}: DashboardWidgetCardProps) {
  // チャートの設定を生成（x_axis, y_axis形式で）
  const chartConfig = useMemo<ChartDataRequest | null>(() => {
    if (widget.view_type !== "chart") return null;
    const firstField = widget.app?.fields?.[0]?.field_code || "";
    if (!firstField) return null;
    return {
      chart_type: "bar",
      x_axis: { field: firstField },
      y_axis: { field: firstField, aggregation: "count" },
    };
  }, [widget.view_type, widget.app?.fields]);

  // DnD設定
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging: isSortableDragging,
  } = useSortable({ id: widget.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging || isSortableDragging ? 0.5 : 1,
  };

  // レコードデータの取得（chartビュー以外の場合のみ）
  const { data: recordsData, isLoading: isRecordsLoading } = useRecords(
    widget.view_type !== "chart" ? widget.app_id : undefined,
    { page: 1, limit: getRecordLimit(widget.widget_size) }
  );

  // チャートデータの取得
  const { data: chartData, isLoading: isChartLoading } = useChartData(
    widget.app_id,
    widget.view_type === "chart" ? chartConfig : null
  );

  const fields = widget.app?.fields || [];
  const records = recordsData?.records || [];

  const isLoading =
    (widget.view_type !== "chart" && isRecordsLoading) ||
    (widget.view_type === "chart" && isChartLoading);

  return (
    <Card
      ref={setNodeRef}
      style={style}
      h={getWidgetHeight(widget.widget_size)}
      shadow={isDragging || isSortableDragging ? "lg" : "sm"}
      borderWidth={isDragging || isSortableDragging ? "2px" : "1px"}
      borderColor={isDragging || isSortableDragging ? "brand.400" : "gray.200"}
      overflow="hidden"
    >
      <CardHeader pb={2}>
        <HStack justify="space-between">
          <HStack spacing={3}>
            <IconButton
              {...attributes}
              {...listeners}
              icon={<FiMove />}
              aria-label="ドラッグして移動"
              size="sm"
              variant="ghost"
              cursor="grab"
              _active={{ cursor: "grabbing" }}
            />
            <Icon
              as={getAppIcon(widget.app?.icon || "default")}
              boxSize={6}
              color="brand.500"
            />
            <VStack align="start" spacing={0}>
              <HStack>
                <Heading size="sm" noOfLines={1}>
                  {widget.app?.name || "Unknown App"}
                </Heading>
                {widget.app?.is_external && (
                  <Badge colorScheme="purple" fontSize="xs">
                    外部
                  </Badge>
                )}
              </HStack>
              <Text fontSize="xs" color="gray.500">
                {widget.view_type === "table" && "テーブル表示"}
                {widget.view_type === "list" && "リスト表示"}
                {widget.view_type === "chart" && "グラフ表示"}
              </Text>
            </VStack>
          </HStack>
        </HStack>
      </CardHeader>

      <CardBody pt={0} overflow="auto">
        {isLoading ? (
          <Box
            display="flex"
            alignItems="center"
            justifyContent="center"
            h="full"
          >
            <Spinner />
          </Box>
        ) : (
          <>
            {widget.view_type === "table" && (
              <Box overflowX="auto">
                <RecordTable
                  records={records}
                  fields={fields}
                  selectedIds={[]}
                  onSelectRecord={() => {}}
                  onSelectAll={() => {}}
                  onView={() => {}}
                  isAdmin={false}
                />
              </Box>
            )}

            {widget.view_type === "list" && (
              <ListView
                records={records}
                fields={fields}
                onView={() => {}}
                isAdmin={false}
              />
            )}

            {widget.view_type === "chart" && chartData && (
              <Box h="full">
                <ChartRenderer
                  chartType={chartConfig?.chart_type || "bar"}
                  data={chartData}
                />
              </Box>
            )}

            {widget.view_type === "chart" && !chartData && (
              <Box
                display="flex"
                alignItems="center"
                justifyContent="center"
                h="full"
                color="gray.500"
              >
                <Text>グラフ設定が必要です</Text>
              </Box>
            )}

            {records.length === 0 && widget.view_type !== "chart" && (
              <Box
                display="flex"
                alignItems="center"
                justifyContent="center"
                h="full"
                color="gray.500"
              >
                <Text>データがありません</Text>
              </Box>
            )}
          </>
        )}
      </CardBody>
    </Card>
  );
}
