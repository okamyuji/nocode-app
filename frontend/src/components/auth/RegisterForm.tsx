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

export function RegisterForm() {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [errors, setErrors] = useState<{
    name?: string;
    email?: string;
    password?: string;
    confirmPassword?: string;
  }>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { register } = useAuth();

  const validate = () => {
    const newErrors: typeof errors = {};
    if (!name) {
      newErrors.name = "名前を入力してください";
    }
    if (!email) {
      newErrors.email = "メールアドレスを入力してください";
    } else if (!/\S+@\S+\.\S+/.test(email)) {
      newErrors.email = "有効なメールアドレスを入力してください";
    }
    if (!password) {
      newErrors.password = "パスワードを入力してください";
    } else if (password.length < 6) {
      newErrors.password = "パスワードは6文字以上で入力してください";
    }
    if (password !== confirmPassword) {
      newErrors.confirmPassword = "パスワードが一致しません";
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    setIsSubmitting(true);
    try {
      await register(email, password, name);
      toaster.create({
        title: "登録完了",
        description: "アカウントが作成されました",
        type: "success",
        duration: 3000,
        closable: true,
      });
    } catch {
      toaster.create({
        title: "登録失敗",
        description:
          "アカウントの作成に失敗しました。メールアドレスが既に使用されている可能性があります。",
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
        <Field.Root invalid={!!errors.name}>
          <Field.Label>名前</Field.Label>
          <Input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="山田 太郎"
          />
          <Field.ErrorText>{errors.name}</Field.ErrorText>
        </Field.Root>

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
            placeholder="6文字以上で入力"
          />
          <Field.ErrorText>{errors.password}</Field.ErrorText>
        </Field.Root>

        <Field.Root invalid={!!errors.confirmPassword}>
          <Field.Label>パスワード（確認）</Field.Label>
          <Input
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="パスワードを再入力"
          />
          <Field.ErrorText>{errors.confirmPassword}</Field.ErrorText>
        </Field.Root>

        <Button
          type="submit"
          colorPalette="brand"
          w="100%"
          loading={isSubmitting}
          loadingText="登録中..."
        >
          新規登録
        </Button>

        <Text fontSize="sm" color="gray.600">
          既にアカウントをお持ちの方は
          <Link asChild color="brand.500" ml={1}>
            <RouterLink to="/login">ログイン</RouterLink>
          </Link>
        </Text>
      </VStack>
    </Box>
  );
}
