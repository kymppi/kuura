import { Button, Form, PasswordInput, Stack, TextInput } from '@carbon/react';
import { useEffect, useState } from 'react';
import { SRPAuthClient } from '../lib/auth.client';
import { getServiceInfo, type ServiceInfo } from '../lib/service.client';
import { DefaultPrimeField } from '../lib/srp.client';

const client = new SRPAuthClient('', DefaultPrimeField);

export default function LoginPage({ returnTo }: { returnTo: string }) {
  const [serviceData, setServiceData] = useState<ServiceInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchServiceInfo = async () => {
      try {
        setIsLoading(true);
        const data = await getServiceInfo();
        setServiceData(data);
      } catch (err) {
        setError('Failed to load service information');
        console.error(err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchServiceInfo();
  }, []);

  return (
    <div
      style={{
        display: 'grid',
        placeItems: 'center',
        height: '100%',
        padding: '1rem',
        backgroundImage: 'url("/login.jpg")',
        backgroundSize: 'cover',
        backgroundRepeat: 'no-repeat',
        backgroundPosition: 'center',
      }}
    >
      <div
        style={{
          width: '100%',
          maxWidth: '32rem',
          padding: '1rem',
          backgroundColor: 'white',
        }}
      >
        {isLoading ? (
          <p>Loading service information...</p>
        ) : error ? (
          <p>{error}</p>
        ) : serviceData ? (
          <LoginForm info={serviceData} returnTo={returnTo} />
        ) : null}
      </div>
    </div>
  );
}

function LoginForm({
  info,
  returnTo,
}: {
  info: ServiceInfo;
  returnTo: string;
}) {
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
      <Stack gap={7}>
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
