import {
  CHART_TYPE_LABELS,
  ChartAxis,
  ChartDataRequest,
  ChartType,
  Field,
} from "@/types";
import {
  Button,
  FormControl,
  FormLabel,
  HStack,
  Input,
  Select,
  VStack,
} from "@chakra-ui/react";
import { useState } from "react";

interface ChartConfigFormProps {
  fields: Field[];
  onSubmit: (config: ChartDataRequest) => void;
  initialConfig?: ChartDataRequest;
  isLoading?: boolean;
}

export function ChartConfigForm({
  fields,
  onSubmit,
  initialConfig,
  isLoading = false,
}: ChartConfigFormProps) {
  const [chartType, setChartType] = useState<ChartType>(
    initialConfig?.chart_type || "bar"
  );
  const [xAxis, setXAxis] = useState<ChartAxis>(
    initialConfig?.x_axis || { field: "", label: "" }
  );
  const [yAxis, setYAxis] = useState<ChartAxis>(
    initialConfig?.y_axis || { field: "", aggregation: "count", label: "" }
  );
  const [groupBy, setGroupBy] = useState(initialConfig?.group_by || "");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      chart_type: chartType,
      x_axis: xAxis,
      y_axis: yAxis,
      group_by: groupBy || undefined,
    });
  };

  const aggregationOptions = [
    { value: "count", label: "カウント" },
    { value: "sum", label: "合計" },
    { value: "avg", label: "平均" },
    { value: "min", label: "最小" },
    { value: "max", label: "最大" },
  ];

  const textFields = fields.filter((f) =>
    ["text", "select", "radio", "date"].includes(f.field_type)
  );
  const numberFields = fields.filter((f) => f.field_type === "number");

  return (
    <VStack as="form" onSubmit={handleSubmit} spacing={4} align="stretch">
      <FormControl isRequired>
        <FormLabel>グラフタイプ</FormLabel>
        <Select
          value={chartType}
          onChange={(e) => setChartType(e.target.value as ChartType)}
        >
          {Object.entries(CHART_TYPE_LABELS).map(([value, label]) => (
            <option key={value} value={value}>
              {label}
            </option>
          ))}
        </Select>
      </FormControl>

      <FormControl isRequired>
        <FormLabel>X軸（分類）</FormLabel>
        <Select
          value={xAxis.field}
          onChange={(e) => setXAxis({ ...xAxis, field: e.target.value })}
          placeholder="フィールドを選択"
        >
          {textFields.map((field) => (
            <option key={field.id} value={field.field_code}>
              {field.field_name}
            </option>
          ))}
        </Select>
      </FormControl>

      <HStack spacing={4}>
        <FormControl flex={1}>
          <FormLabel>Y軸（集計フィールド）</FormLabel>
          <Select
            value={yAxis.field}
            onChange={(e) => setYAxis({ ...yAxis, field: e.target.value })}
            placeholder="フィールドを選択（空の場合はカウント）"
          >
            <option value="">（レコード数をカウント）</option>
            {numberFields.map((field) => (
              <option key={field.id} value={field.field_code}>
                {field.field_name}
              </option>
            ))}
          </Select>
        </FormControl>

        <FormControl flex={1}>
          <FormLabel>集計方法</FormLabel>
          <Select
            value={yAxis.aggregation || "count"}
            onChange={(e) =>
              setYAxis({
                ...yAxis,
                aggregation: e.target.value as ChartAxis["aggregation"],
              })
            }
          >
            {aggregationOptions.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </Select>
        </FormControl>
      </HStack>

      <FormControl>
        <FormLabel>グループ化（オプション）</FormLabel>
        <Select
          value={groupBy}
          onChange={(e) => setGroupBy(e.target.value)}
          placeholder="なし"
        >
          <option value="">なし</option>
          {textFields
            .filter((f) => f.field_code !== xAxis.field)
            .map((field) => (
              <option key={field.id} value={field.field_code}>
                {field.field_name}
              </option>
            ))}
        </Select>
      </FormControl>

      <FormControl>
        <FormLabel>Y軸ラベル</FormLabel>
        <Input
          value={yAxis.label || ""}
          onChange={(e) => setYAxis({ ...yAxis, label: e.target.value })}
          placeholder="例: 売上金額"
        />
      </FormControl>

      <Button
        type="submit"
        colorScheme="brand"
        isLoading={isLoading}
        loadingText="グラフを生成中..."
      >
        グラフを表示
      </Button>
    </VStack>
  );
}
