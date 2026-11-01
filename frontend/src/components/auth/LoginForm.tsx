import { useAuth } from "@/hooks";
import {
  Box,
  Button,
  Field,
  Input,
  Link,
  Text,
  VStack,
} from "@chakra-ui/react";
import { toaster } from "@/components/ui/toaster";
import { useState } from "react";
import { Link as RouterLink } from "react-router-dom";

export function LoginForm() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<{ email?: string; password?: string }>(
    {}
  );
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { login } = useAuth();

  const validate = () => {
    const newErrors: { email?: string; password?: string } = {};
    if (!email) {
      newErrors.email = "メールアドレスを入力してください";
    } else if (!/\S+@\S+\.\S+/.test(email)) {
      newErrors.email = "有効なメールアドレスを入力してください";
    }
    if (!password) {
      newErrors.password = "パスワードを入力してください";
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    setIsSubmitting(true);
    try {
      await login(email, password);
      toaster.create({
        title: "ログイン成功",
        type: "success",
        duration: 3000,
        closable: true,
      });
    } catch {
      toaster.create({
        title: "ログイン失敗",
        description: "メールアドレスまたはパスワードが正しくありません",
        type: "error",
        duration: 5000,
        closable: true,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Box as="form" onSubmit={handleSubmit} w="100%">
      <VStack gap={4}>
        <Field.Root invalid={!!errors.email}>
          <Field.Label>メールアドレス</Field.Label>
          <Input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="email@example.com"
          />
          <Field.ErrorText>{errors.email}</Field.ErrorText>
        </Field.Root>

        <Field.Root invalid={!!errors.password}>
          <Field.Label>パスワード</Field.Label>
          <Input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="パスワードを入力"
          />
          <Field.ErrorText>{errors.password}</Field.ErrorText>
        </Field.Root>

        <Button
          type="submit"
          colorPalette="brand"
          w="100%"
          loading={isSubmitting}
          loadingText="ログイン中..."
        >
          ログイン
        </Button>

        <Text fontSize="sm" color="gray.600">
          アカウントをお持ちでない方は
          <Link asChild color="brand.500" ml={1}>
            <RouterLink to="/register">新規登録</RouterLink>
          </Link>
        </Text>
      </VStack>
    </Box>
  );
}
