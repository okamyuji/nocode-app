/**
 * カレンダービューコンポーネント
 * レコードを月間カレンダー形式で表示する
 */

import { Field, RecordItem } from "@/types";
import { ChevronLeftIcon, ChevronRightIcon } from "@chakra-ui/icons";
import {
  Box,
  Button,
  Grid,
  GridItem,
  Heading,
  HStack,
  IconButton,
  Select,
  Text,
  VStack,
} from "@chakra-ui/react";
import { useMemo, useState } from "react";

interface CalendarViewProps {
  records: RecordItem[];
  fields: Field[];
  dateFieldCode?: string;
  onRecordClick?: (record: RecordItem) => void;
  labelFieldCode?: string;
}

/** 曜日ラベル */
const WEEKDAYS = ["日", "月", "火", "水", "木", "金", "土"];

export function CalendarView({
  records,
  fields,
  dateFieldCode,
  onRecordClick,
  labelFieldCode,
}: CalendarViewProps) {
  const [currentDate, setCurrentDate] = useState(new Date());
  const [selectedDateField, setSelectedDateField] = useState(
    dateFieldCode || ""
  );

  // 日付フィールドを抽出
  const dateFields = useMemo(
    () =>
      fields.filter(
        (f) => f.field_type === "date" || f.field_type === "datetime"
      ),
    [fields]
  );

  // 日付フィールドが指定されていない場合、最初の日付フィールドを自動選択
  const activeDateField = selectedDateField || dateFields[0]?.field_code || "";

  // ラベルフィールドを取得（指定されていない場合は最初のテキストフィールド）
  const labelField = useMemo(() => {
    if (labelFieldCode) {
      return fields.find((f) => f.field_code === labelFieldCode);
    }
    return fields.find((f) => f.field_type === "text") || fields[0];
  }, [fields, labelFieldCode]);

  // 現在の月の日数と最初の曜日を計算
  const { daysInMonth, firstDayOfWeek, year, month } = useMemo(() => {
    const year = currentDate.getFullYear();
    const month = currentDate.getMonth();
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    return {
      year,
      month,
      daysInMonth: lastDay.getDate(),
      firstDayOfWeek: firstDay.getDay(),
    };
  }, [currentDate]);

  // レコードを日付でグループ化
  const recordsByDate = useMemo(() => {
    const map = new Map<string, RecordItem[]>();

    records.forEach((record) => {
      const dateValue = record.data[activeDateField];
      if (!dateValue) return;

      const date = new Date(dateValue as string);
      if (isNaN(date.getTime())) return;

      const dateKey = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
      const existing = map.get(dateKey) || [];
      map.set(dateKey, [...existing, record]);
    });

    return map;
  }, [records, activeDateField]);

  /** 月を移動する */
  const navigateMonth = (delta: number) => {
    setCurrentDate((prev) => {
      const newDate = new Date(prev);
      newDate.setMonth(prev.getMonth() + delta);
      return newDate;
    });
  };

  /** 今日に移動する */
  const goToToday = () => {
    setCurrentDate(new Date());
  };

  // カレンダーグリッドを生成
  const calendarDays = useMemo(() => {
    const days: (number | null)[] = [];

    // 月初日より前の空セルを追加
    for (let i = 0; i < firstDayOfWeek; i++) {
      days.push(null);
    }

    // 月の日数分のセルを追加
    for (let i = 1; i <= daysInMonth; i++) {
      days.push(i);
    }

    return days;
  }, [daysInMonth, firstDayOfWeek]);

  /** 指定日のレコードを取得 */
  const getRecordsForDay = (day: number) => {
    const dateKey = `${year}-${month}-${day}`;
    return recordsByDate.get(dateKey) || [];
  };

  /** 指定日が今日かどうかを判定 */
  const isToday = (day: number) => {
    const today = new Date();
    return (
      today.getFullYear() === year &&
      today.getMonth() === month &&
      today.getDate() === day
    );
  };

  // 日付フィールドがない場合のメッセージ
  if (dateFields.length === 0) {
    return (
      <Box p={8} textAlign="center">
        <Text color="gray.500">
          カレンダービューを使用するには、日付フィールドが必要です。
        </Text>
      </Box>
    );
  }

  return (
    <Box>
      {/* ヘッダー */}
      <HStack justify="space-between" mb={4}>
        <HStack spacing={4}>
          <HStack>
            <IconButton
              icon={<ChevronLeftIcon />}
              aria-label="前月"
              size="sm"
              onClick={() => navigateMonth(-1)}
            />
            <Heading size="md" minW="150px" textAlign="center">
              {year}年 {month + 1}月
            </Heading>
            <IconButton
              icon={<ChevronRightIcon />}
              aria-label="翌月"
              size="sm"
              onClick={() => navigateMonth(1)}
            />
          </HStack>
          <Button size="sm" onClick={goToToday}>
            今日
          </Button>
        </HStack>

        <HStack>
          <Text fontSize="sm" color="gray.600">
            日付フィールド:
          </Text>
          <Select
            size="sm"
            w="auto"
            value={activeDateField}
            onChange={(e) => setSelectedDateField(e.target.value)}
          >
            {dateFields.map((field) => (
              <option key={field.id} value={field.field_code}>
                {field.field_name}
              </option>
            ))}
          </Select>
        </HStack>
      </HStack>

      {/* 曜日ヘッダー */}
      <Grid templateColumns="repeat(7, 1fr)" gap={1} mb={2}>
        {WEEKDAYS.map((day, index) => (
          <GridItem
            key={day}
            textAlign="center"
            fontWeight="bold"
            fontSize="sm"
            color={
              index === 0 ? "red.500" : index === 6 ? "blue.500" : "gray.600"
            }
            py={2}
          >
            {day}
          </GridItem>
        ))}
      </Grid>

      {/* カレンダーグリッド */}
      <Grid templateColumns="repeat(7, 1fr)" gap={1}>
        {calendarDays.map((day, index) => {
          const dayRecords = day ? getRecordsForDay(day) : [];
          const dayOfWeek = index % 7;

          return (
            <GridItem
              key={index}
              minH="100px"
              bg={day ? "white" : "gray.50"}
              border="1px"
              borderColor="gray.200"
              borderRadius="md"
              p={2}
            >
              {day && (
                <VStack align="stretch" spacing={1}>
                  <Text
                    fontSize="sm"
                    fontWeight={isToday(day) ? "bold" : "normal"}
                    color={
                      isToday(day)
                        ? "brand.500"
                        : dayOfWeek === 0
                          ? "red.500"
                          : dayOfWeek === 6
                            ? "blue.500"
                            : "gray.700"
                    }
                    bg={isToday(day) ? "brand.50" : "transparent"}
                    borderRadius="full"
                    w="24px"
                    h="24px"
                    lineHeight="24px"
                    textAlign="center"
                  >
                    {day}
                  </Text>

                  {dayRecords.slice(0, 3).map((record) => (
                    <Box
                      key={record.id}
                      bg="brand.100"
                      px={2}
                      py={0.5}
                      borderRadius="sm"
                      fontSize="xs"
                      cursor="pointer"
                      _hover={{ bg: "brand.200" }}
                      onClick={() => onRecordClick?.(record)}
                      isTruncated
                    >
                      {labelField
                        ? String(
                            record.data[labelField.field_code] ||
                              `ID: ${record.id}`
                          )
                        : `ID: ${record.id}`}
                    </Box>
                  ))}

                  {dayRecords.length > 3 && (
                    <Text fontSize="xs" color="gray.500" textAlign="center">
                      +{dayRecords.length - 3}件
                    </Text>
                  )}
                </VStack>
              )}
            </GridItem>
          );
        })}
      </Grid>
    </Box>
  );
}
