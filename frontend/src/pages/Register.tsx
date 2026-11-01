import { RegisterForm } from "@/components/auth";
import { Box, Card, Heading, VStack } from "@chakra-ui/react";

export function RegisterPage() {
  return (
    <Box
      minH="100vh"
      display="flex"
      alignItems="center"
      justifyContent="center"
      bg="gray.50"
      p={4}
    >
      <Card.Root maxW="400px" w="100%">
        <Card.Body>
          <VStack gap={6}>
            <Heading size="lg" color="brand.500">
              Nocode App
            </Heading>
            <Heading size="md">新規登録</Heading>
            <RegisterForm />
          </VStack>
        </Card.Body>
      </Card.Root>
    </Box>
  );
}
