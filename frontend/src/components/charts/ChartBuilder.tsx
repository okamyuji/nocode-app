import { Field } from "@/types";
import { ChartDataRequest, ChartType } from "@/types/chart";
import {
  Box,
  Button,
  FormControl,
  FormLabel,
  HStack,
  Icon,
  Radio,
  RadioGroup,
  Select,
  Text,
  VStack,
  Wrap,
  WrapItem,
} from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { BsGraphUp } from "react-icons/bs";
import {
  FiBarChart,
  FiBarChart2,
  FiPieChart,
  FiTrendingUp,
} from "react-icons/fi";
import { TbChartDonut } from "react-icons/tb";

interface ChartBuilderProps {
  fields: Field[];
  config?: ChartDataRequest;
  onConfigChange: (config: ChartDataRequest) => void;
  onSave?: () => void;
}

const chartTypes: {
  type: ChartType;
  label: string;
  icon: React.ComponentType;
}[] = [
  { type: "bar", label: "棒グラフ", icon: FiBarChart2 },
  { type: "horizontal_bar", label: "横棒", icon: FiBarChart },
  { type: "line", label: "折れ線", icon: FiTrendingUp },
  { type: "pie", label: "円グラフ", icon: FiPieChart },
  { type: "doughnut", label: "ドーナツ", icon: TbChartDonut },
  { type: "area", label: "面グラフ", icon: BsGraphUp },
];

const aggregationTypes = [
  { value: "count", label: "件数" },
  { value: "sum", label: "合計" },
  { value: "avg", label: "平均" },
  { value: "min", label: "最小" },
  { value: "max", label: "最大" },
];

export function ChartBuilder({
  fields,
  config,
  onConfigChange,
  onSave,
}: ChartBuilderProps) {
  const [chartType, setChartType] = useState<ChartType>(
    config?.chart_type || "bar"
  );
  const [xAxisField, setXAxisField] = useState(config?.x_axis?.field || "");
  const [aggregation, setAggregation] = useState(
    config?.y_axis?.aggregation || "count"
  );
  const [yAxisField, setYAxisField] = useState(config?.y_axis?.field || "");

  // Filter fields suitable for X-axis (categorical)
  const categoricalFields = fields.filter(
    (f) =>
      f.field_type === "text" ||
      f.field_type === "select" ||
      f.field_type === "radio" ||
      f.field_type === "date"
  );

  // Filter fields suitable for Y-axis aggregation (numeric)
  const numericFields = fields.filter((f) => f.field_type === "number");

  // Build and emit config when any value changes
  useEffect(() => {
    if (!xAxisField) return;

    const xField = fields.find((f) => f.field_code === xAxisField);
    const yField = fields.find((f) => f.field_code === yAxisField);

    const newConfig: ChartDataRequest = {
      chart_type: chartType,
      x_axis: {
        field: xAxisField,
        label: xField?.field_name || xAxisField,
      },
      y_axis: {
        field: aggregation === "count" ? "" : yAxisField,
        aggregation: aggregation as "count" | "sum" | "avg" | "min" | "max",
        label:
          yField?.field_name ||
          aggregationTypes.find((a) => a.value === aggregation)?.label ||
          aggregation,
      },
      filters: [],
    };

    onConfigChange(newConfig);
  }, [chartType, xAxisField, aggregation, yAxisField, fields, onConfigChange]);

  const handleChartTypeChange = (type: ChartType) => {
    setChartType(type);
  };

  const handleXAxisChange = (fieldCode: string) => {
    setXAxisField(fieldCode);
  };

  const handleAggregationChange = (agg: string) => {
    setAggregation(agg as "count" | "sum" | "avg" | "min" | "max");
    if (agg === "count") {
      setYAxisField("");
    }
  };

  const handleYAxisFieldChange = (fieldCode: string) => {
    setYAxisField(fieldCode);
  };

  return (
    <VStack align="stretch" spacing={4}>
      {/* Settings Row - Aligned at top */}
      <HStack spacing={6} align="flex-start" wrap="wrap">
        {/* Chart Type */}
        <FormControl w="auto" flexShrink={0}>
          <FormLabel fontSize="sm" fontWeight="bold" mb={2}>
            グラフ種類
          </FormLabel>
          <RadioGroup
            value={chartType}
            onChange={(v) => handleChartTypeChange(v as ChartType)}
          >
            <Wrap spacing={1}>
              {chartTypes.map((ct) => (
                <WrapItem key={ct.type}>
                  <Box
                    as="label"
                    px={2}
                    py={1.5}
                    border="1px"
                    borderColor={
                      chartType === ct.type ? "brand.500" : "gray.200"
                    }
                    borderRadius="md"
                    bg={chartType === ct.type ? "brand.50" : "white"}
                    cursor="pointer"
                    _hover={{ borderColor: "brand.300" }}
                    display="flex"
                    alignItems="center"
                    gap={1.5}
                    h="32px"
                  >
                    <Radio value={ct.type} display="none" />
                    <Icon
                      as={ct.icon}
                      boxSize={4}
                      color={chartType === ct.type ? "brand.500" : "gray.500"}
                    />
                    <Text
                      fontSize="xs"
                      fontWeight={chartType === ct.type ? "bold" : "normal"}
                    >
                      {ct.label}
                    </Text>
                  </Box>
                </WrapItem>
              ))}
            </Wrap>
          </RadioGroup>
        </FormControl>

        {/* X-Axis */}
        <FormControl w="160px" flexShrink={0}>
          <FormLabel fontSize="sm" fontWeight="bold" mb={2}>
            X軸
          </FormLabel>
          <Select
            size="sm"
            placeholder="フィールドを選択"
            value={xAxisField}
            onChange={(e) => handleXAxisChange(e.target.value)}
            h="32px"
          >
            {categoricalFields.map((field) => (
              <option key={field.id} value={field.field_code}>
                {field.field_name}
              </option>
            ))}
          </Select>
        </FormControl>

        {/* Y-Axis Aggregation */}
        <FormControl w="120px" flexShrink={0}>
          <FormLabel fontSize="sm" fontWeight="bold" mb={2}>
            Y軸（集計）
          </FormLabel>
          <Select
            size="sm"
            value={aggregation}
            onChange={(e) => handleAggregationChange(e.target.value)}
            h="32px"
          >
            {aggregationTypes.map((agg) => (
              <option key={agg.value} value={agg.value}>
                {agg.label}
              </option>
            ))}
          </Select>
        </FormControl>

        {/* Y-Axis Field (only for non-count aggregations) */}
        {aggregation !== "count" && (
          <FormControl w="160px" flexShrink={0}>
            <FormLabel fontSize="sm" fontWeight="bold" mb={2}>
              集計フィールド
            </FormLabel>
            <Select
              size="sm"
              placeholder="数値フィールド"
              value={yAxisField}
              onChange={(e) => handleYAxisFieldChange(e.target.value)}
              h="32px"
            >
              {numericFields.map((field) => (
                <option key={field.id} value={field.field_code}>
                  {field.field_name}
                </option>
              ))}
            </Select>
          </FormControl>
        )}

        {onSave && (
          <FormControl w="auto" flexShrink={0}>
            <FormLabel
              fontSize="sm"
              fontWeight="bold"
              mb={2}
              visibility="hidden"
            >
              &nbsp;
            </FormLabel>
            <Button
              size="sm"
              colorScheme="brand"
              onClick={onSave}
              isDisabled={!xAxisField}
              h="32px"
            >
              保存
            </Button>
          </FormControl>
        )}
      </HStack>
    </VStack>
  );
}
