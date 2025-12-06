import { LoginForm } from "@/components/auth";
import { Box, Card, CardBody, Heading, VStack } from "@chakra-ui/react";

export function LoginPage() {
  return (
    <Box
      minH="100vh"
      display="flex"
      alignItems="center"
      justifyContent="center"
      bg="gray.50"
      p={4}
    >
      <Card maxW="400px" w="100%">
        <CardBody>
          <VStack spacing={6}>
            <Heading size="lg" color="brand.500">
              Nocode App
            </Heading>
            <Heading size="md">ログイン</Heading>
            <LoginForm />
          </VStack>
        </CardBody>
      </Card>
    </Box>
  );
}
