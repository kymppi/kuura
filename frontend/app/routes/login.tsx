import { useEffect, useState } from 'react';
import { useLocation } from 'react-router';
import LoginLayout from '../layouts/LoginLayout';
import { getServiceInfo, type ServiceInfo } from '../lib/service.client';
import LoginForm from '../login/LoginForm';
import type { Route } from './+types/home';

export function meta({}: Route.MetaArgs) {
  return [
    { title: 'Kuura' },
    {
      name: 'description',
      content: 'Log in to Kuura.',
    },
  ];
}

export default function Login() {
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

  const location = useLocation();
  const searchParams = new URLSearchParams(location.search);
  const returnTo = searchParams.get('return_to') || '/home';

  return (
    <LoginLayout>
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
    </LoginLayout>
  );
}
