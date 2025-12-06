import { ChartConfigForm, ChartRenderer } from "@/components/charts";
import { Loading } from "@/components/common";
import { useApp, useChartData, useFields } from "@/hooks";
import { ChartDataRequest } from "@/types";
import { ChevronRightIcon } from "@chakra-ui/icons";
import {
  Box,
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  Card,
  CardBody,
  CardHeader,
  Heading,
  SimpleGrid,
  Text,
} from "@chakra-ui/react";
import { useState } from "react";
import { Link as RouterLink, useParams } from "react-router-dom";

export function AppChartPage() {
  const { appId } = useParams<{ appId: string }>();
  const numericAppId = Number(appId);

  const [chartConfig, setChartConfig] = useState<ChartDataRequest | null>(null);

  const { data: app, isLoading: isAppLoading } = useApp(numericAppId);
  const { data: fieldsData, isLoading: isFieldsLoading } =
    useFields(numericAppId);
  const { data: chartData, isLoading: isChartLoading } = useChartData(
    numericAppId,
    chartConfig
  );

  const fields = fieldsData?.fields || [];

  if (isAppLoading || isFieldsLoading) {
    return <Loading message="アプリを読み込み中..." />;
  }

  if (!app) {
    return <Box>アプリが見つかりません</Box>;
  }

  return (
    <Box>
      <Breadcrumb
        spacing="8px"
        separator={<ChevronRightIcon color="gray.500" />}
        mb={4}
      >
        <BreadcrumbItem>
          <BreadcrumbLink as={RouterLink} to="/apps">
            アプリ一覧
          </BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbItem>
          <BreadcrumbLink as={RouterLink} to={`/apps/${appId}/records`}>
            {app.name}
          </BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbItem isCurrentPage>
          <BreadcrumbLink>グラフ</BreadcrumbLink>
        </BreadcrumbItem>
      </Breadcrumb>

      <Heading size="lg" mb={6}>
        {app.name} - グラフ
      </Heading>

      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6}>
        <Card>
          <CardHeader>
            <Heading size="md">グラフ設定</Heading>
          </CardHeader>
          <CardBody>
            <ChartConfigForm
              fields={fields}
              onSubmit={setChartConfig}
              initialConfig={chartConfig || undefined}
              isLoading={isChartLoading}
            />
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <Heading size="md">プレビュー</Heading>
          </CardHeader>
          <CardBody>
            {isChartLoading ? (
              <Loading message="グラフを生成中..." />
            ) : chartData ? (
              <ChartRenderer
                data={chartData}
                chartType={chartConfig?.chart_type || "bar"}
              />
            ) : (
              <Box
                h="400px"
                display="flex"
                alignItems="center"
                justifyContent="center"
                border="2px dashed"
                borderColor="gray.200"
                borderRadius="md"
              >
                <Text color="gray.400">
                  グラフ設定を入力して「グラフを表示」をクリックしてください
                </Text>
              </Box>
            )}
          </CardBody>
        </Card>
      </SimpleGrid>
    </Box>
  );
}
