import { Button, PasswordInput, Stack, TextInput } from '@carbon/react';
import { useState } from 'react';
import { Form } from 'react-router';
import { useAuthentication } from '../hooks/useAuthentication';
import type { ServiceInfo } from '../lib/service.client';

export default function LoginForm({
  info,
  returnTo,
}: {
  readonly info: ServiceInfo;
  readonly returnTo: string;
}) {
  const { client } = useAuthentication();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [inlineError, setInlineError] = useState<string | null>(null);
  const [usernameInvalid, setUsernameInvalid] = useState(false);
  const [passwordInvalid, setPasswordInvalid] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    setUsernameInvalid(false);
    setPasswordInvalid(false);
    setInlineError(null);

    if (!username || !password) {
      if (!username) setUsernameInvalid(true);
      if (!password) setPasswordInvalid(true);
      setInlineError('Username and password are required');
      return;
    }

    const result = await client.login(info.id, username, password);

    if (!result.success) {
      setInlineError(result.error);
    } else {
      setInlineError(null);
      window.location.href = returnTo;
    }
  };

  return (
    <Form onSubmit={handleSubmit}>
      <Stack gap={8}>
        <Stack gap="0.25rem">
          <h1>Log in to {info.name}</h1>
          <p>
            Don't have an account? Contact{' '}
            <a href={info.contact.link} target="_blank">
              {info.contact.name}
            </a>
            .
          </p>
        </Stack>
        <Stack gap="1rem">
          <TextInput
            id="username"
            labelText="Username"
            placeholder="Enter your username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            invalid={usernameInvalid}
            invalidText="Username is required"
            required
          />
          <PasswordInput
            id="password"
            labelText="Password"
            placeholder="Enter your password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            invalid={passwordInvalid}
            invalidText="Password is required"
            required
          />
          {inlineError && (
            <p style={{ color: 'red', marginTop: '0.5rem' }}>{inlineError}</p>
          )}
        </Stack>

        <Button type="submit" style={{ justifySelf: 'flex-end' }}>
          Log in
        </Button>
      </Stack>
    </Form>
  );
}
