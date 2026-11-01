import {
  CHART_TYPE_LABELS,
  ChartAxis,
  ChartDataRequest,
  ChartType,
  Field as FieldModel,
} from "@/types";
import {
  Button,
  Field,
  HStack,
  Input,
  NativeSelect,
  VStack,
} from "@chakra-ui/react";
import { useState } from "react";

interface ChartConfigFormProps {
  fields: FieldModel[];
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
    <VStack as="form" onSubmit={handleSubmit} gap={4} align="stretch">
      <Field.Root required>
        <Field.Label>
          グラフタイプ
          <Field.RequiredIndicator />
        </Field.Label>
        <NativeSelect.Root>
          <NativeSelect.Field
            value={chartType}
            onChange={(e) => setChartType(e.target.value as ChartType)}
          >
            {Object.entries(CHART_TYPE_LABELS).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </NativeSelect.Field>
          <NativeSelect.Indicator />
        </NativeSelect.Root>
      </Field.Root>

      <Field.Root required>
        <Field.Label>
          X軸（分類）
          <Field.RequiredIndicator />
        </Field.Label>
        <NativeSelect.Root>
          <NativeSelect.Field
            value={xAxis.field}
            onChange={(e) => setXAxis({ ...xAxis, field: e.target.value })}
            placeholder="フィールドを選択"
          >
            {textFields.map((field) => (
              <option key={field.id} value={field.field_code}>
                {field.field_name}
              </option>
            ))}
          </NativeSelect.Field>
          <NativeSelect.Indicator />
        </NativeSelect.Root>
      </Field.Root>

      <HStack gap={4}>
        <Field.Root flex={1}>
          <Field.Label>Y軸（集計フィールド）</Field.Label>
          <NativeSelect.Root>
            <NativeSelect.Field
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
            </NativeSelect.Field>
            <NativeSelect.Indicator />
          </NativeSelect.Root>
        </Field.Root>

        <Field.Root flex={1}>
          <Field.Label>集計方法</Field.Label>
          <NativeSelect.Root>
            <NativeSelect.Field
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
            </NativeSelect.Field>
            <NativeSelect.Indicator />
          </NativeSelect.Root>
        </Field.Root>
      </HStack>

      <Field.Root>
        <Field.Label>グループ化（オプション）</Field.Label>
        <NativeSelect.Root>
          <NativeSelect.Field
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
          </NativeSelect.Field>
          <NativeSelect.Indicator />
        </NativeSelect.Root>
      </Field.Root>

      <Field.Root>
        <Field.Label>Y軸ラベル</Field.Label>
        <Input
          value={yAxis.label || ""}
          onChange={(e) => setYAxis({ ...yAxis, label: e.target.value })}
          placeholder="例: 売上金額"
        />
      </Field.Root>

      <Button
        type="submit"
        colorPalette="brand"
        loading={isLoading}
        loadingText="グラフを生成中..."
      >
        グラフを表示
      </Button>
    </VStack>
  );
}
