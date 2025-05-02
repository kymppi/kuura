import { Button, Form, PasswordInput, Stack, TextInput } from '@carbon/react';
import { useEffect, useState } from 'react';
import { getServiceInfo, type ServiceInfo } from '../lib/service.client';

export default function LoginPage() {
  const [serviceData, setServiceData] = useState<ServiceInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchServiceInfo = async () => {
      try {
        const urlParams = new URLSearchParams(window.location.search);
        const serviceId = urlParams.get('service');

        if (serviceId) {
          setIsLoading(true);
          const data = await getServiceInfo(serviceId);
          setServiceData(data);
        } else {
          setError('No service ID provided');
        }
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
          <LoginForm info={serviceData} />
        ) : null}
      </div>
    </div>
  );
}

function LoginForm({ info }: { info: ServiceInfo }) {
  return (
    <Form style={{}}>
      <Stack gap={7}>
        <Stack gap="0.25rem">
          <h1>Log in to {info.name}</h1>
          <p>
            Don't have an account? Contact the{' '}
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
            required
          />

          <PasswordInput
            id="password"
            labelText="Password"
            placeholder="Enter your password"
            required
          />
        </Stack>

        <Button
          style={{
            justifySelf: 'flex-end',
          }}
          type="submit"
        >
          Log in
        </Button>
      </Stack>
    </Form>
  );
}
