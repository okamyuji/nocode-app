/**
 * ローディングスピナーコンポーネント
 */

import { Center, Spinner, Text, VStack } from "@chakra-ui/react";

interface LoadingProps {
  message?: string;
  fullScreen?: boolean;
}

export function Loading({
  message = "読み込み中...",
  fullScreen = false,
}: LoadingProps) {
  const content = (
    <VStack gap={4}>
      <Spinner
        borderWidth="4px"
        animationDuration="0.65s"
        css={{ "--spinner-track-color": "colors.gray.200" }}
        color="brand.500"
        size="xl"
      />
      <Text color="gray.500">{message}</Text>
    </VStack>
  );

  // フルスクリーン表示
  if (fullScreen) {
    return (
      <Center
        h="100vh"
        w="100vw"
        position="fixed"
        top={0}
        left={0}
        bg="white"
        zIndex={1000}
      >
        {content}
      </Center>
    );
  }

  return <Center py={12}>{content}</Center>;
}
